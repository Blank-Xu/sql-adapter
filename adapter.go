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
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

const (
	// defaultTableName  if tableName == "", the Adapter will use this default table name.
	defaultTableName = "casbin_rule"

	// maxParamLength  .
	maxParamLength = 7
)

// CasbinRule  defines the casbin rule model.
// It used for save or load policy lines from connected database.
type CasbinRule struct {
	PType string
	V0    string
	V1    string
	V2    string
	V3    string
	V4    string
	V5    string
}

// Adapter  defines the database adapter for Casbin.
// It can load policy lines from connected database or save policy lines.
type Adapter struct {
	db         *sql.DB
	ctx        context.Context
	driverName string
	tableName  string

	isFiltered bool

	sqlPlaceHolder   string
	sqlCreateTable   string
	sqlTruncateTable string
	sqlIsTableExist  string
	sqlInsertRow     string
	sqlDeleteAll     string
	sqlDeleteByArgs  string
	sqlSelectAll     string
	sqlSelectWhere   string
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
// If tableName == "", the Adapter will automatically create a table named "casbin_rule".
func NewAdapter(db *sql.DB, driverName, tableName string) (*Adapter, error) {
	return NewAdapterContext(context.Background(), db, driverName, tableName)
}

// NewAdapterContext  the constructor for Adapter.
// db should connected to database and controlled by user.
// If tableName == "", the Adapter will automatically create a table named "casbin_rule".
func NewAdapterContext(ctx context.Context, db *sql.DB, driverName, tableName string) (*Adapter, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	switch driverName {
	case "":
		return nil, errors.New("need driverName param")
	case "oci8", "ora", "goracle":
		return nil, errors.New("sqladapter: please checkout 'oracle' branch")
	}

	// check db connecting
	err := db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	if tableName == "" {
		tableName = defaultTableName
	}

	adapter := Adapter{
		db:         db,
		ctx:        ctx,
		driverName: driverName,
		tableName:  tableName,
	}

	// generate different databases sql
	adapter.genSQL()

	if !adapter.isTableExist() {
		if err = adapter.createTable(); err != nil {
			return nil, err
		}
	}

	return &adapter, nil
}

// genSQL  generate sql based on db driver name.
func (p *Adapter) genSQL() {
	p.sqlPlaceHolder = sqlPlaceHolder
	p.sqlCreateTable = fmt.Sprintf(sqlCreateTable, p.tableName)
	p.sqlTruncateTable = fmt.Sprintf(sqlTruncateTable, p.tableName)

	p.sqlIsTableExist = fmt.Sprintf(sqlIsTableExist, p.tableName)

	p.sqlInsertRow = fmt.Sprintf(sqlInsertRow, p.tableName)
	p.sqlDeleteAll = fmt.Sprintf(sqlDeleteAll, p.tableName)
	p.sqlDeleteByArgs = fmt.Sprintf(sqlDeleteByArgs, p.tableName)

	p.sqlSelectAll = fmt.Sprintf(sqlSelectAll, p.tableName)
	p.sqlSelectWhere = fmt.Sprintf(sqlSelectWhere, p.tableName)

	switch p.driverName {
	case "postgres", "pgx", "pq-timeouts", "cloudsqlpostgres":
		p.sqlPlaceHolder = sqlPlaceHolderPostgres
		p.sqlCreateTable = fmt.Sprintf(sqlCreateTablePostgres, p.tableName)
		p.sqlInsertRow = fmt.Sprintf(sqlInsertRowPostgres, p.tableName)
	case "mysql":
		p.sqlCreateTable = fmt.Sprintf(sqlCreateTableMysql, p.tableName)
	case "sqlite3":
		p.sqlCreateTable = fmt.Sprintf(sqlCreateTableSqlite3, p.tableName)
		p.sqlTruncateTable = fmt.Sprintf(sqlTruncateTableSqlite3, p.tableName)
	case "sqlserver":
		p.sqlPlaceHolder = sqlPlaceHolderSqlserver
		p.sqlCreateTable = fmt.Sprintf(sqlCreateTableSqlserver, p.tableName)
		p.sqlInsertRow = fmt.Sprintf(sqlInsertRowSqlserver, p.tableName)
	}
}

func (p *Adapter) sqlRebind(query string) string {
	if p.sqlPlaceHolder == sqlPlaceHolder {
		return query
	}

	var idx, num int

	result := make([]byte, 0, len(query)+10)

	for {
		idx = strings.Index(query, sqlPlaceHolder)
		if idx == -1 {
			break
		}

		num++

		result = append(result, query[:idx]...)
		result = append(result, p.sqlPlaceHolder...)
		result = strconv.AppendInt(result, int64(num), 10)

		query = query[idx+1:]
	}

	return string(append(result, query...))
}

// createTable  create a not exists table.
func (p *Adapter) createTable() error {
	_, err := p.db.ExecContext(p.ctx, p.sqlCreateTable)

	return err
}

// truncateTable  clear the table.
func (p *Adapter) truncateTable() error {
	_, err := p.db.ExecContext(p.ctx, p.sqlTruncateTable)

	return err
}

// isTableExist  check the table exists.
func (p *Adapter) isTableExist() bool {
	_, err := p.db.ExecContext(p.ctx, p.sqlIsTableExist)

	return err == nil
}

// deleteRows  delete eligible data.
func (p *Adapter) deleteRows(query string, args ...interface{}) error {
	query = p.sqlRebind(query)

	_, err := p.db.ExecContext(p.ctx, query, args...)

	return err
}

// truncateAndInsertRows  clear table and insert new rows.
func (p *Adapter) truncateAndInsertRows(rules [][]interface{}) (err error) {
	if err = p.truncateTable(); err != nil {
		return
	}

	tx, err := p.db.BeginTx(p.ctx, nil)
	if err != nil {
		return
	}

	var action string
	var stmt *sql.Stmt
	// if _, err = tx.Exec(p.sqlDeleteAll); err != nil {
	// 	action = "delete all"
	// 	goto ROLLBACK
	// }

	stmt, err = tx.PrepareContext(p.ctx, p.sqlInsertRow)
	if err != nil {
		action = "prepare context"
		goto ROLLBACK
	}

	for _, rule := range rules {
		if _, err = stmt.ExecContext(p.ctx, rule...); err != nil {
			action = "stmt exec context"
			goto ROLLBACK
		}
	}

	if err = stmt.Close(); err != nil {
		action = "stmt close"
		goto ROLLBACK
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

// queryFunc  define func for query
var queryFunc = func(ctx context.Context, db *sql.DB, query string, args ...string) ([]*CasbinRule, error) {
	params := make([]interface{}, len(args))
	for idx := range args {
		params[idx] = args[idx]
	}

	lines := make([]*CasbinRule, 0, 64)

	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var rule CasbinRule

		err = rows.Scan(&rule.PType, &rule.V0, &rule.V1, &rule.V2, &rule.V3, &rule.V4, &rule.V5)
		if err != nil {
			return nil, err
		}

		lines = append(lines, &rule)
	}

	return lines, nil
}

// selectRows  select eligible data by args from the table.
func (p *Adapter) selectRows(query string, args ...string) ([]*CasbinRule, error) {
	if len(args) == 0 {
		return queryFunc(p.ctx, p.db, query)
	}

	query = p.sqlRebind(query)

	return queryFunc(p.ctx, p.db, query, args...)
}

// selectWhereIn  select eligible data by filter from the table.
func (p *Adapter) selectWhereIn(filter *Filter) (lines []*CasbinRule, err error) {
	var (
		sqlBuf bytes.Buffer
		buf    bytes.Buffer
	)

	sqlBuf.Grow(64)
	sqlBuf.WriteString(p.sqlSelectWhere)

	args := make([]string, 0, 4)

	for _, col := range [maxParamLength]struct {
		name string
		arg  []string
	}{
		{"p_type", filter.PType},
		{"v0", filter.V0},
		{"v1", filter.V1},
		{"v2", filter.V2},
		{"v3", filter.V3},
		{"v4", filter.V4},
		{"v5", filter.V5},
	} {
		l := len(col.arg)
		if l == 0 {
			continue
		}

		switch sqlBuf.Bytes()[sqlBuf.Len()-1] {
		case '?', ')':
			sqlBuf.WriteString(" AND ")
		}

		sqlBuf.WriteString(col.name)

		if l == 1 {
			sqlBuf.WriteString(" = ?")
			args = append(args, col.arg[0])
		} else {
			buf.Grow(l * 2)
			for i := 0; i < l; i++ {
				buf.WriteString("?,")
			}
			buf.Truncate(buf.Len() - 1)

			sqlBuf.WriteString(" IN (")
			sqlBuf.Write(buf.Bytes())
			sqlBuf.WriteByte(')')

			args = append(args, col.arg...)

			buf.Reset()
		}
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
	args := make([][]interface{}, 0, 32)

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

	_, err := p.db.ExecContext(p.ctx, p.sqlInsertRow, args...)

	return err
}

// RemovePolicy  remove policy rules from the storage.
func (p *Adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	var sqlBuf bytes.Buffer

	sqlBuf.Grow(64)
	sqlBuf.WriteString(p.sqlDeleteByArgs)

	args := make([]interface{}, 0, 4)
	args = append(args, ptype)

	for idx, arg := range rule {
		if arg != "" {
			sqlBuf.WriteString(" AND v")
			sqlBuf.WriteString(strconv.Itoa(idx))
			sqlBuf.WriteString(" = ?")

			args = append(args, arg)
		}
	}

	return p.deleteRows(sqlBuf.String(), args...)
}

// RemoveFilteredPolicy  remove policy rules that match the filter from the storage.
func (p *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	var sqlBuf bytes.Buffer

	sqlBuf.Grow(64)
	sqlBuf.WriteString(p.sqlDeleteByArgs)

	args := make([]interface{}, 0, 4)
	args = append(args, ptype)

	var value string

	l := fieldIndex + len(fieldValues)

	for idx := 0; idx < 6; idx++ {
		if fieldIndex <= idx && idx < l {
			value = fieldValues[idx-fieldIndex]

			if value != "" {
				sqlBuf.WriteString(" AND v")
				sqlBuf.WriteString(strconv.Itoa(idx))
				sqlBuf.WriteString(" = ?")

				args = append(args, value)
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
	if line == nil {
		return
	}

	var lineBuf bytes.Buffer

	lineBuf.Grow(64)
	lineBuf.WriteString(line.PType)

	args := [6]string{line.V0, line.V1, line.V2, line.V3, line.V4, line.V5}
	for _, arg := range args {
		if arg != "" {
			lineBuf.WriteByte(',')
			lineBuf.WriteString(arg)
		}
	}

	persist.LoadPolicyLine(lineBuf.String(), model)
}

// genArgs  generate args from ptype and rule.
func (Adapter) genArgs(ptype string, rule []string) []interface{} {
	args := make([]interface{}, maxParamLength)
	args[0] = ptype

	for idx := range rule {
		args[idx+1] = rule[idx]
	}

	for idx := len(rule) + 1; idx < maxParamLength; idx++ {
		args[idx] = ""
	}

	return args
}
