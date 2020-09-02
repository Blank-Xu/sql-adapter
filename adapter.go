// Copyright 2020 by Blank-Xu. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sqladapter

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

// defaultTableName  if tableName == "", the Adapter will use this default table name.
const defaultTableName = "CASBIN_RULE"

// CasbinRule  defines the casbin rule model.
// It used for save or load policy lines from oracle.
type CasbinRule struct {
	PType string         `db:"P_TYPE"`
	V0    string         `db:"V0"`
	V1    string         `db:"V1"`
	V2    sql.NullString `db:"V2"`
	V3    sql.NullString `db:"V3"`
	V4    sql.NullString `db:"V4"`
	V5    sql.NullString `db:"V5"`
}

// Adapter  define the adapter for Casbin.
// It can load policy lines or save policy lines from connected database.
type Adapter struct {
	db        *sql.DB
	tableName string

	isFiltered bool

	sqlCreateTable   []string
	sqlTruncateTable string
	sqlIsTableExist  string
	sqlInsertRow     []byte
	sqlDeleteAll     string
	sqlDeleteByArgs  []byte
	sqlSelectAll     string
	sqlSelectWhere   []byte

	cols         [][]byte
	placeholders [][]byte
}

// Filter  defines the filtering rules for a FilteredAdapter's policy.
// Empty values are ignored, but all others must match the filter.
type Filter struct {
	PType []string
	V0    []string
	V1    []string
	V2    []string
	V3    []string
	V4    []string
	V5    []string
}

// NewAdapter  the constructor for Adapter.
// db should connected to database and controlled by user.
// If tableName == "", the Adapter will automatically create a table named 'CASBIN_RULE'.
func NewAdapter(db *sql.DB, tableName string) (*Adapter, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	// check db connecting
	err := db.Ping()
	if err != nil {
		return nil, err
	}

	if tableName == "" {
		tableName = defaultTableName
	}

	adapter := Adapter{
		db:        db,
		tableName: tableName,
	}

	// generate sql
	adapter.genSQL()

	// generate sql params
	adapter.genParams()

	if !adapter.isTableExist() {
		if err = adapter.createTable(); err != nil {
			return nil, err
		}
	}

	return &adapter, nil
}

// genSQL  generate sql based on db driver name.
func (p *Adapter) genSQL() {
	p.tableName = strings.ToUpper(p.tableName)

	p.sqlCreateTable = []string{
		fmt.Sprintf(sqlCreateTable, p.tableName),
		fmt.Sprintf(sqlCreateIndex, p.tableName),
	}

	p.sqlTruncateTable = fmt.Sprintf(sqlTruncateTable, p.tableName)
	p.sqlIsTableExist = fmt.Sprintf(sqlIsTableExist, p.tableName)

	p.sqlInsertRow = []byte(fmt.Sprintf(sqlInsertRow, p.tableName))
	p.sqlDeleteAll = fmt.Sprintf(sqlDeleteAll, p.tableName)
	p.sqlDeleteByArgs = []byte(fmt.Sprintf(sqlDeleteByArgs, p.tableName))

	p.sqlSelectAll = fmt.Sprintf(sqlSelectAll, p.tableName)
	p.sqlSelectWhere = []byte(fmt.Sprintf(sqlSelectWhere, p.tableName))
}

// genParams  generate all cols and placeholders by db driver name.
func (p *Adapter) genParams() {
	var line CasbinRule

	t := reflect.TypeOf(line)
	l := t.NumField()

	var (
		colBuf bytes.Buffer
		phBuf  bytes.Buffer
	)

	colBuf.Grow(16)
	phBuf.Grow(16)

	p.cols = make([][]byte, l)
	p.placeholders = make([][]byte, l)

	for i := 0; i < l; i++ {
		tag := t.Field(i).Tag.Get("db")
		colBuf.WriteString(tag)
		p.cols[i] = []byte("(" + colBuf.String() + ")")
		colBuf.WriteByte(',')

		phBuf.Write(sqlPlaceHolder)
		phBuf.WriteString(strconv.Itoa(i + 1))
		p.placeholders[i] = []byte("(" + phBuf.String() + ")")
		phBuf.WriteByte(',')
	}
}

// createTable  create a not exists table.
func (p *Adapter) createTable() (err error) {
	for _, query := range p.sqlCreateTable {
		if _, err = p.db.Exec(query); err != nil {
			return
		}
	}
	return
}

// truncateTable  clear the table.
func (p *Adapter) truncateTable() error {
	_, err := p.db.Exec(p.sqlTruncateTable)

	return err
}

// isTableExist  check the table exists.
func (p *Adapter) isTableExist() bool {
	_, err := p.db.Query(p.sqlIsTableExist)

	return err == nil
}

// deleteRows  delete eligible data.
func (p *Adapter) deleteRows(query string, args ...interface{}) error {
	_, err := p.db.Exec(query, args...)

	return err
}

// truncateAndInsertRows  clear table and insert new rows.
func (p *Adapter) truncateAndInsertRows(args [][]interface{}) (err error) {
	if err = p.truncateTable(); err != nil {
		return
	}

	tx, err := p.db.Begin()
	if err != nil {
		return
	}

	var action string
	// if _, err = tx.Exec(p.sqlDeleteAll); err != nil {
	// 	action = "delete all"
	// 	goto ROLLBACK
	// }

	var sqlBuf bytes.Buffer

	for _, arg := range args {
		l := len(arg)
		if l == 0 {
			continue
		}

		sqlBuf.Grow(128)
		sqlBuf.Write(p.sqlInsertRow)
		sqlBuf.Write(p.cols[l-1])
		sqlBuf.WriteString(" VALUES ")
		sqlBuf.Write(p.placeholders[l-1])

		if _, err = tx.Exec(sqlBuf.String(), arg...); err != nil {
			action = "exec"
			goto ROLLBACK
		}

		sqlBuf.Reset()
	}

	if err = tx.Commit(); err != nil {
		action = "commit"
		goto ROLLBACK
	}

	return

ROLLBACK:

	if err1 := tx.Rollback(); err1 != nil {
		err = fmt.Errorf("%s err: %v, rollback err: %v", action, err, err1)
	}

	return
}

// selectRows  select all data from the table.
func (p *Adapter) selectRows(query string, args ...string) (lines []*CasbinRule, err error) {
	// make a slice with capacity
	lines = make([]*CasbinRule, 0, 64)

	params := make([]interface{}, len(args))
	for idx := range args {
		params[idx] = args[idx]
	}

	rows, err := p.db.Query(query, params...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var rule CasbinRule

		if err = rows.Scan(&rule.PType, &rule.V0, &rule.V1, &rule.V2, &rule.V3, &rule.V4, &rule.V5); err != nil {
			return nil, err
		}

		lines = append(lines, &rule)
	}

	return lines, nil
}

// selectWhereIn  select eligible data by filter from the table.
func (p *Adapter) selectWhereIn(filter *Filter) (lines []*CasbinRule, err error) {
	var sqlBuf bytes.Buffer

	sqlBuf.Grow(128)
	sqlBuf.Write(p.sqlSelectWhere)

	args := make([]string, 0, 4)

	var buf bytes.Buffer

	j := 1

	for _, col := range [7]struct {
		name string
		arg  []string
	}{
		{"P_TYPE", filter.PType},
		{"V0", filter.V0},
		{"V1", filter.V1},
		{"V2", filter.V2},
		{"V3", filter.V3},
		{"V4", filter.V4},
		{"V5", filter.V5},
	} {
		l := len(col.arg)
		if l == 0 {
			continue
		}

		sqlBuf.WriteString(col.name)

		if l == 1 {
			sqlBuf.WriteByte('=')
			sqlBuf.Write(sqlPlaceHolder)
			sqlBuf.WriteString(strconv.Itoa(j))

			args = append(args, col.arg[0])

			j++
		} else {
			buf.Grow(l * 4)
			for i := 0; i < l; i++ {
				buf.Write(sqlPlaceHolder)
				buf.WriteString(strconv.Itoa(j))
				buf.WriteByte(',')

				j++
			}
			buf.Truncate(buf.Len() - 1)

			sqlBuf.WriteString(" IN (")
			sqlBuf.Write(buf.Bytes())
			sqlBuf.WriteByte(')')

			args = append(args, col.arg...)

			buf.Reset()
		}

		sqlBuf.WriteString(sqlAND)
	}

	l := sqlBuf.Len()
	if l > 5 && sqlBuf.String()[l-5:] == sqlAND {
		sqlBuf.Truncate(l - 5)
	}

	return p.selectRows(sqlBuf.String(), args...)
}

// LoadPolicy  load all policy rules from the storage.
func (p *Adapter) LoadPolicy(model model.Model) error {
	lines, err := p.selectRows(p.sqlSelectAll)
	if err != nil {
		return err
	}

	for _, line := range lines {
		p.loadPolicyLine(line, model)
	}

	return nil
}

// SavePolicy  save policy rules to the storage.
func (p *Adapter) SavePolicy(model model.Model) error {
	args := make([][]interface{}, 0, 64)

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			arg := p.genArgs(ptype, rule)
			args = append(args, arg)
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			arg := p.genArgs(ptype, rule)
			args = append(args, arg)
		}
	}

	return p.truncateAndInsertRows(args)
}

// AddPolicy  add one policy rule to the storage.
func (p *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	args := p.genArgs(ptype, rule)
	idx := len(args) - 1

	var sqlBuf bytes.Buffer

	sqlBuf.Grow(128)
	sqlBuf.Write(p.sqlInsertRow)
	sqlBuf.Write(p.cols[idx])
	sqlBuf.WriteString(" VALUES ")
	sqlBuf.Write(p.placeholders[idx])

	_, err := p.db.Exec(sqlBuf.String(), args...)

	return err
}

// RemovePolicy  remove policy rules from the storage.
func (p *Adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	var sqlBuf bytes.Buffer

	sqlBuf.Grow(128)
	sqlBuf.Write(p.sqlDeleteByArgs)

	args := make([]interface{}, 0, len(rule)+1)
	args = append(args, ptype)
	j := 2

	for idx, arg := range rule {
		if arg != "" {
			sqlBuf.WriteString(" AND V")
			sqlBuf.WriteString(strconv.Itoa(idx))
			sqlBuf.WriteString(" = :arg")
			sqlBuf.WriteString(strconv.Itoa(j))

			args = append(args, arg)

			j++
		}
	}

	return p.deleteRows(sqlBuf.String(), args...)
}

// RemoveFilteredPolicy  remove policy rules that match the filter from the storage.
func (p *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	var sqlBuf bytes.Buffer

	sqlBuf.Grow(128)
	sqlBuf.Write(p.sqlDeleteByArgs)

	args := make([]interface{}, 0, 4)
	args = append(args, ptype)

	var value string

	l := fieldIndex + len(fieldValues)
	j := 2

	for idx := 0; idx < 6; idx++ {
		if fieldIndex <= idx && idx < l {
			value = fieldValues[idx-fieldIndex]

			if value != "" {
				sqlBuf.WriteString(" AND V")
				sqlBuf.WriteString(strconv.Itoa(idx))
				sqlBuf.WriteString(" = :arg")
				sqlBuf.WriteString(strconv.Itoa(j))

				args = append(args, value)

				j++
			}
		}
	}

	return p.deleteRows(sqlBuf.String(), args...)
}

// LoadFilteredPolicy  load policy rules that match the filter.
// filterPtr must be a pointer.
func (p *Adapter) LoadFilteredPolicy(model model.Model, filterPtr interface{}) error {
	if filterPtr == nil {
		return p.LoadPolicy(model)
	}

	filter, ok := filterPtr.(*Filter)
	if !ok {
		return errors.New("invalid filter type")
	}

	lines, err := p.selectWhereIn(filter)
	if err != nil {
		return err
	}

	for _, line := range lines {
		p.loadPolicyLine(line, model)
	}

	p.isFiltered = true

	return nil
}

// IsFiltered  returns true if the loaded policy rules has been filtered.
func (p *Adapter) IsFiltered() bool {
	return p.isFiltered
}

// loadPolicyLine  load a policy line to model.
func (Adapter) loadPolicyLine(line *CasbinRule, model model.Model) {
	var lineBuf bytes.Buffer

	lineBuf.Grow(128)
	lineBuf.WriteString(line.PType)

	args := [6]string{
		line.V0,
		line.V1,
		line.V2.String,
		line.V3.String,
		line.V4.String,
		line.V5.String,
	}

	for _, arg := range args {
		if arg != "" {
			lineBuf.WriteByte(',')
			lineBuf.WriteString(arg)
		}
	}

	persist.LoadPolicyLine(lineBuf.String(), model)
}

// genArg  generate args from pType and rule.
func (Adapter) genArgs(ptype string, rule []string) []interface{} {
	l := len(rule)
	args := make([]interface{}, l+1)

	args[0] = ptype

	for i := 0; i < l; i++ {
		args[i+1] = rule[i]
	}

	return args
}
