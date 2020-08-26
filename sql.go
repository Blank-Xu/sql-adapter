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

// general sql
const (
	sqlPlaceHolder = "?"
	sqlCreateTable = `
CREATE TABLE %[1]s(
    p_type VARCHAR(32),
    v0     VARCHAR(255),
    v1     VARCHAR(255),
    v2     VARCHAR(255),
    v3     VARCHAR(255),
    v4     VARCHAR(255),
    v5     VARCHAR(255)
);
CREATE INDEX idx_%[1]s ON %[1]s (p_type, v0, v1);`
	sqlTruncateTable = "TRUNCATE TABLE %s"
	sqlIsTableExist  = "SELECT 1 FROM %s"
	sqlInsertRow     = "INSERT INTO %s (p_type, v0, v1, v2, v3, v4, v5) VALUES (?, ?, ?, ?, ?, ?, ?)"
	sqlDeleteAll     = "DELETE FROM %s"
	sqlDeleteByArgs  = "DELETE FROM %s WHERE p_type = ?"
	sqlSelectAll     = "SELECT p_type,v0,v1,v2,v3,v4,v5 FROM %s"
	sqlSelectWhere   = "SELECT p_type,v0,v1,v2,v3,v4,v5 FROM %s WHERE "
)

// for Sqlite3
const (
	sqlCreateTableSqlite3 = `
CREATE TABLE IF NOT EXISTS %[1]s(
    p_type VARCHAR(32)  DEFAULT '' NOT NULL,
    v0     VARCHAR(255) DEFAULT '' NOT NULL,
    v1     VARCHAR(255) DEFAULT '' NOT NULL,
    v2     VARCHAR(255) DEFAULT '' NOT NULL,
    v3     VARCHAR(255) DEFAULT '' NOT NULL,
    v4     VARCHAR(255) DEFAULT '' NOT NULL,
    v5     VARCHAR(255) DEFAULT '' NOT NULL,
    CHECK (TYPEOF("p_type") = "text" AND
           LENGTH("p_type") <= 32),
    CHECK (TYPEOF("v0") = "text" AND
           LENGTH("v0") <= 255),
    CHECK (TYPEOF("v1") = "text" AND
           LENGTH("v1") <= 255),
    CHECK (TYPEOF("v2") = "text" AND
           LENGTH("v2") <= 255),
    CHECK (TYPEOF("v3") = "text" AND
           LENGTH("v3") <= 255),
    CHECK (TYPEOF("v4") = "text" AND
           LENGTH("v4") <= 255),
    CHECK (TYPEOF("v5") = "text" AND
           LENGTH("v5") <= 255)
);
CREATE INDEX IF NOT EXISTS idx_%[1]s ON %[1]s (p_type, v0, v1);`
	sqlTruncateTableSqlite3 = "DROP TABLE IF EXISTS %[1]s;" + sqlCreateTableSqlite3
)

// for Mysql
const (
	sqlCreateTableMysql = `
CREATE TABLE IF NOT EXISTS %[1]s(
    p_type VARCHAR(32)  DEFAULT '' NOT NULL,
    v0     VARCHAR(255) DEFAULT '' NOT NULL,
    v1     VARCHAR(255) DEFAULT '' NOT NULL,
    v2     VARCHAR(255) DEFAULT '' NOT NULL,
    v3     VARCHAR(255) DEFAULT '' NOT NULL,
    v4     VARCHAR(255) DEFAULT '' NOT NULL,
    v5     VARCHAR(255) DEFAULT '' NOT NULL,
    INDEX idx_%[1]s (p_type, v0, v1)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;`
)

// for Postgres
const (
	sqlPlaceHolderPostgres = "$"
	sqlCreateTablePostgres = `
CREATE TABLE IF NOT EXISTS %[1]s(
    p_type VARCHAR(32)  DEFAULT '' NOT NULL,
    v0     VARCHAR(255) DEFAULT '' NOT NULL,
    v1     VARCHAR(255) DEFAULT '' NOT NULL,
    v2     VARCHAR(255) DEFAULT '' NOT NULL,
    v3     VARCHAR(255) DEFAULT '' NOT NULL,
    v4     VARCHAR(255) DEFAULT '' NOT NULL,
    v5     VARCHAR(255) DEFAULT '' NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_%[1]s ON %[1]s (p_type, v0, v1);`
	sqlInsertRowPostgres = "INSERT INTO %s (p_type, v0, v1, v2, v3, v4, v5) VALUES ($1, $2, $3, $4, $5, $6, $7)"
)

// for Sqlserver
const (
	sqlPlaceHolderSqlserver = "@p"
	sqlCreateTableSqlserver = `
CREATE TABLE %[1]s(
    p_type NVARCHAR(32)  DEFAULT '' NOT NULL,
    v0     NVARCHAR(255) DEFAULT '' NOT NULL,
    v1     NVARCHAR(255) DEFAULT '' NOT NULL,
    v2     NVARCHAR(255) DEFAULT '' NOT NULL,
    v3     NVARCHAR(255) DEFAULT '' NOT NULL,
    v4     NVARCHAR(255) DEFAULT '' NOT NULL,
    v5     NVARCHAR(255) DEFAULT '' NOT NULL
);
CREATE INDEX idx_%[1]s ON %[1]s (p_type, v0, v1);`
	sqlInsertRowSqlserver = "INSERT INTO %s (p_type, v0, v1, v2, v3, v4, v5) VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7)"
)
