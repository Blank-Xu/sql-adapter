// Copyright 2023 by Blank-Xu. All Rights Reserved.
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
)

func getDao(db *sql.DB, driverName, tableName string) (dao, error) {
	d := dao{
		db: db,

		tableName:   tableName,
		placeHolder: defaultPlaceholder,

		sqlCreateTable:   fmt.Sprintf(sqlCreateTable, tableName),
		sqlTruncateTable: fmt.Sprintf(sqlTruncateTable, tableName),

		sqlTableExist: fmt.Sprintf(sqlTableExist, tableName),

		sqlInsertRow:    fmt.Sprintf(sqlInsertRow, tableName),
		sqlUpdateRow:    fmt.Sprintf(sqlUpdateRow, tableName),
		sqlDeleteAll:    fmt.Sprintf(sqlDeleteAll, tableName),
		sqlDeleteRow:    fmt.Sprintf(sqlDeleteRow, tableName),
		sqlDeleteByArgs: fmt.Sprintf(sqlDeleteByArgs, tableName),

		sqlSelectAll:   fmt.Sprintf(sqlSelectAll, tableName),
		sqlSelectWhere: fmt.Sprintf(sqlSelectWhere, tableName),
	}

	var err error

	switch driverName {
	case "postgres", "pgx", "cloudsql-postgres":
		d.placeHolder = sqlPlaceholderPostgreSQL
		d.sqlCreateTable = fmt.Sprintf(sqlCreateTablePostgreSQL, tableName)
		d.sqlInsertRow = fmt.Sprintf(sqlInsertRowPostgreSQL, tableName)
		d.sqlUpdateRow = fmt.Sprintf(sqlUpdateRowPostgreSQL, tableName)
		d.sqlDeleteRow = fmt.Sprintf(sqlDeleteRowPostgreSQL, tableName)
	case "mysql":
		d.sqlCreateTable = fmt.Sprintf(sqlCreateTableMySQL, tableName)
	case "sqlite", "sqlite3":
		d.sqlCreateTable = fmt.Sprintf(sqlCreateTableSQLite3, tableName)
		d.sqlTruncateTable = fmt.Sprintf(sqlTruncateTableSQLite3, tableName)
	case "sqlserver":
		d.placeHolder = sqlPlaceholderSQLServer
		d.sqlCreateTable = fmt.Sprintf(sqlCreateTableSQLServer, tableName)
		d.sqlInsertRow = fmt.Sprintf(sqlInsertRowSQLServer, tableName)
		d.sqlUpdateRow = fmt.Sprintf(sqlUpdateRowSQLServer, tableName)
		d.sqlDeleteRow = fmt.Sprintf(sqlDeleteRowSQLServer, tableName)
	case "mssql":
		err = errors.New("driver name mssql not support, please use sqlserver")
	case "oci8", "ora", "goracle":
		err = errors.New("sqladapter: please checkout 'oracle' branch")
	default:
		err = fmt.Errorf("unsupported driver name: %s", driverName)
	}

	return d, err
}

type dao struct {
	db *sql.DB

	tableName string

	placeHolder string

	sqlCreateTable string

	sqlTableExist  string
	sqlSelectAll   string
	sqlSelectWhere string

	sqlInsertRow string
	sqlUpdateRow string

	// not necessary to use DDL for delete
	sqlTruncateTable string

	sqlDeleteAll    string
	sqlDeleteRow    string
	sqlDeleteByArgs string
}

func (d dao) TableName() string {
	return d.tableName
}

// rebindSQL rebind SQL by different database.
func (d dao) rebindSQL(query string) string {
	if d.placeHolder == defaultPlaceholder {
		return query
	}

	var idx, num int

	result := make([]byte, 0, len(query)+10)

	for {
		idx = strings.Index(query, defaultPlaceholder)
		if idx == -1 {
			break
		}

		num++

		result = append(result, query[:idx]...)
		result = append(result, d.placeHolder...)
		result = strconv.AppendInt(result, int64(num), 10)

		query = query[idx+1:]
	}

	return string(append(result, query...))
}

// querySQL query data by sql.
func (d dao) querySQL(ctx context.Context, query string, args ...interface{}) ([]rule, error) {
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]rule, 0, 128)

	for rows.Next() {
		var rule rule

		err = rows.Scan(&rule.PType, &rule.V0, &rule.V1, &rule.V2, &rule.V3, &rule.V4, &rule.V5)
		if err != nil {
			return nil, err
		}

		rules = append(rules, rule)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

// execSQL exec sql.
func (d dao) execSQL(ctx context.Context, query string, args ...interface{}) error {
	_, err := d.db.ExecContext(ctx, query, args...)

	return err
}

type txData struct {
	step  string
	query string
	args  []interface{}
}

// execTxSQL exec transaction sql rows.
func (d dao) execTxSQL(ctx context.Context, beforeTxData, afterTxData txData, query string, args [][]interface{}) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx err: %v", err)
	}

	var (
		step string
		stmt *sql.Stmt
	)

	if beforeTxData.query != "" {
		if _, err = tx.ExecContext(ctx, beforeTxData.query, beforeTxData.args...); err != nil {
			step = beforeTxData.step + " before prepare"
			goto ROLLBACK
		}
	}

	if stmt, err = tx.PrepareContext(ctx, query); err != nil {
		step = "prepare context"
		goto ROLLBACK
	}

	for _, arg := range args {
		if _, err = stmt.ExecContext(ctx, arg...); err != nil {
			step = "stmt exec context"
			goto ROLLBACK
		}
	}

	if err = stmt.Close(); err != nil {
		step = "stmt close"
		goto ROLLBACK
	}

	if afterTxData.query != "" {
		if _, err = tx.ExecContext(ctx, afterTxData.query, afterTxData.args...); err != nil {
			step = afterTxData.step + " after stmt"
			goto ROLLBACK
		}
	}

	if err = tx.Commit(); err != nil {
		step = "commit"
		goto ROLLBACK
	}

	return nil

ROLLBACK:

	if err1 := tx.Rollback(); err1 != nil {
		return fmt.Errorf("%s err: %v, rollback err: %v", step, err, err1)
	}

	return fmt.Errorf("%s err: %v", step, err)
}

// CreateTable create a table.
func (d dao) CreateTable(ctx context.Context) error {
	return d.execSQL(ctx, d.sqlCreateTable)
}

// IsTableExist check the table exists.
func (d dao) IsTableExist(ctx context.Context) bool {
	return d.execSQL(ctx, d.sqlTableExist) == nil
}

// SelectAll select all data from the table.
func (d dao) SelectAll(ctx context.Context) ([]rule, error) {
	return d.querySQL(ctx, d.sqlSelectAll)
}

// SelectRows select eligible data by args from the table.
func (d dao) SelectRows(ctx context.Context, query string, args ...interface{}) ([]rule, error) {
	if len(args) == 0 {
		return d.querySQL(ctx, query)
	}

	query = d.rebindSQL(query)

	return d.querySQL(ctx, query, args...)
}

// SelectByCondition
func (d dao) SelectByCondition(ctx context.Context, whereCondition string, args ...interface{}) ([]rule, error) {
	var buf bytes.Buffer

	buf.Grow(128)
	buf.WriteString(d.sqlSelectWhere)
	// this is for reuse the SQL
	buf.WriteString("p_type=?")
	buf.WriteString(whereCondition)

	query := d.rebindSQL(buf.String())

	return d.querySQL(ctx, query, args...)
}

// SelectByFilter select eligible data by Filter from the table.
func (d dao) SelectByFilter(ctx context.Context, filterData [maxParameterCount]filterData) (lines []rule, err error) {
	var (
		sqlBuf bytes.Buffer
		buf    bytes.Buffer
	)

	sqlBuf.Grow(64)
	sqlBuf.WriteString(d.sqlSelectWhere)

	args := make([]string, 0, maxParameterCount)

	for _, col := range filterData {
		l := len(col.arg)
		if l == 0 {
			continue
		}

		switch sqlBuf.Bytes()[sqlBuf.Len()-1] {
		case '?', ')':
			sqlBuf.WriteString(" AND ")
		}

		sqlBuf.WriteString(col.fieldName)

		if l == 1 {
			sqlBuf.WriteString("=?")
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

	params := make([]interface{}, len(args))
	for idx := range args {
		params[idx] = args[idx]
	}

	return d.SelectRows(ctx, sqlBuf.String(), params...)
}

// InsertRow insert one row to the table.
func (d dao) InsertRow(ctx context.Context, args ...interface{}) error {
	return d.execSQL(ctx, d.sqlInsertRow, args...)
}

// InsertRows insert multiple rows to the table by transaction
func (d dao) InsertRows(ctx context.Context, args [][]interface{}) error {
	return d.execTxSQL(ctx, txData{}, txData{}, d.sqlInsertRow, args)
}

// UpdateRow update one row to the table.
func (d dao) UpdateRow(ctx context.Context, args ...interface{}) error {
	return d.execSQL(ctx, d.sqlUpdateRow, args...)
}

// UpdateRows update multiple rows to the table by transaction
func (d dao) UpdateRows(ctx context.Context, args [][]interface{}) error {
	return d.execTxSQL(ctx, txData{}, txData{}, d.sqlUpdateRow, args)
}

// UpdateFilteredRows
func (d dao) UpdateFilteredRows(ctx context.Context, deleteCondition string, deleteArgs []interface{}, updateArgs [][]interface{}) error {
	deleteQuery := d.sqlDeleteByArgs + deleteCondition
	deleteQuery = d.rebindSQL(deleteQuery)

	return d.execTxSQL(ctx, txData{step: "delete rows", query: deleteQuery, args: deleteArgs}, txData{}, d.sqlInsertRow, updateArgs)
}

// DeleteAll clear the table.
// func (d dao) DeleteAll(ctx context.Context) error {
// 	return d.execSQL(ctx, d.sqlDeleteAll)
// }

// DeleteRows delete eligible data.
func (d dao) DeleteRows(ctx context.Context, args [][]interface{}) error {
	return d.execTxSQL(ctx, txData{}, txData{}, d.sqlDeleteRow, args)
}

// DeleteAllAndInsertRows clear table and insert new rows.
func (d dao) DeleteAllAndInsertRows(ctx context.Context, rules [][]interface{}) error {
	return d.execTxSQL(ctx, txData{step: "delete all", query: d.sqlDeleteAll}, txData{}, d.sqlInsertRow, rules)
}

// DeleteByArgs delete eligible data.
func (d dao) DeleteByArgs(ctx context.Context, ptype string, rule []string) error {
	var sqlBuf bytes.Buffer

	sqlBuf.Grow(128)
	sqlBuf.WriteString(d.sqlDeleteByArgs)

	args := make([]interface{}, 0, maxParameterCount)
	args = append(args, ptype)

	for idx, arg := range rule {
		if arg != "" {
			sqlBuf.WriteString(" AND v")
			sqlBuf.WriteString(strconv.Itoa(idx))
			sqlBuf.WriteString("=?")

			args = append(args, arg)
		}
	}

	query := d.rebindSQL(sqlBuf.String())

	return d.execSQL(ctx, query, args...)
}

// DeleteByCondition
func (d dao) DeleteByCondition(ctx context.Context, condition string, args ...interface{}) error {
	deleteQuery := d.sqlDeleteByArgs + condition
	deleteQuery = d.rebindSQL(deleteQuery)

	return d.execSQL(ctx, deleteQuery, args...)
}

// TruncateTable clear the table.
// func (d dao) TruncateTable(ctx context.Context) error {
// 	return d.execSQL(ctx, d.sqlTruncateTable)
// }

// TruncateAndInsertRows clear table and insert new rows.
// func (d dao) TruncateAndInsertRows(ctx context.Context, rules ...[]interface{}) error {
// 	return d.execTxSQL(ctx, txData{step: "truncate table", query: d.sqlTruncateTable}, txData{}, d.sqlInsertRow, rules...)
// }

// GenFilteredCondition
func (d dao) GenFilteredCondition(ptype string, fieldIndex int, fieldValues ...string) (string, []interface{}) {
	var whereConditionBuf bytes.Buffer

	whereConditionBuf.Grow(64)

	args := make([]interface{}, 0, maxParameterCount)
	args = append(args, ptype)

	var value string

	l := fieldIndex + len(fieldValues)

	for idx := 0; idx < maxParameterCount-1; idx++ {
		if fieldIndex <= idx && idx < l {
			value = fieldValues[idx-fieldIndex]

			if value != "" {
				whereConditionBuf.WriteString(" AND v")
				whereConditionBuf.WriteString(strconv.Itoa(idx))
				whereConditionBuf.WriteString("=?")

				args = append(args, value)
			}
		}
	}

	return whereConditionBuf.String(), args
}
