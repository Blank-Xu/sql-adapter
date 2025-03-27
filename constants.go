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

const (
	// defaultTableName  if tableName == "", the Adapter will use this default table name.
	defaultTableName = "casbin_rule"

	// maxParameterCount .
	maxParameterCount = 7

	// defaultPlaceholder .
	defaultPlaceholder = "?"
)

type adapterDriverNameIndex int

const (
	_SQLite adapterDriverNameIndex = iota + 1
	_MySQL
	_PostgreSQL
	_SQLServer
)

// general SQL for all supported databases.
const (
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
CREATE INDEX idx_%[1]s ON %[1]s (p_type,v0,v1);`
	sqlTableExist   = "SELECT 1 FROM %s WHERE 1=0"
	sqlInsertRow    = "INSERT INTO %s (p_type,v0,v1,v2,v3,v4,v5) VALUES (?,?,?,?,?,?,?)"
	sqlUpdateRow    = "UPDATE %s SET p_type=?,v0=?,v1=?,v2=?,v3=?,v4=?,v5=? WHERE p_type=? AND v0=? AND v1=? AND v2=? AND v3=? AND v4=? AND v5=?"
	sqlDeleteAll    = "DELETE FROM %s"
	sqlDeleteRow    = "DELETE FROM %s WHERE p_type=? AND v0=? AND v1=? AND v2=? AND v3=? AND v4=? AND v5=?"
	sqlDeleteByArgs = "DELETE FROM %s WHERE p_type=?"
	sqlSelectAll    = "SELECT p_type,v0,v1,v2,v3,v4,v5 FROM %s"
	sqlSelectWhere  = "SELECT p_type,v0,v1,v2,v3,v4,v5 FROM %s WHERE "
)

// for SQLite3.
const (
	sqlCreateTableSQLite3 = `
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
CREATE INDEX IF NOT EXISTS idx_%[1]s ON %[1]s (p_type,v0,v1);`
	sqlTruncateTableSQLite3 = "DROP TABLE IF EXISTS %[1]s;" + sqlCreateTableSQLite3
)

// for MySQL.
const (
	sqlCreateTableMySQL = `
CREATE TABLE IF NOT EXISTS %[1]s(
    p_type VARCHAR(32)  DEFAULT '' NOT NULL,
    v0     VARCHAR(255) DEFAULT '' NOT NULL,
    v1     VARCHAR(255) DEFAULT '' NOT NULL,
    v2     VARCHAR(255) DEFAULT '' NOT NULL,
    v3     VARCHAR(255) DEFAULT '' NOT NULL,
    v4     VARCHAR(255) DEFAULT '' NOT NULL,
    v5     VARCHAR(255) DEFAULT '' NOT NULL,
    INDEX idx_%[1]s (p_type,v0,v1)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;`
)

// for PostgreSQL.
const (
	sqlPlaceholderPostgreSQL = "$"
	sqlCreateTablePostgreSQL = `
CREATE TABLE IF NOT EXISTS %[1]s(
    p_type VARCHAR(32)  DEFAULT '' NOT NULL,
    v0     VARCHAR(255) DEFAULT '' NOT NULL,
    v1     VARCHAR(255) DEFAULT '' NOT NULL,
    v2     VARCHAR(255) DEFAULT '' NOT NULL,
    v3     VARCHAR(255) DEFAULT '' NOT NULL,
    v4     VARCHAR(255) DEFAULT '' NOT NULL,
    v5     VARCHAR(255) DEFAULT '' NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_%[1]s ON %[1]s (p_type,v0,v1);`
	sqlInsertRowPostgreSQL = "INSERT INTO %s (p_type,v0,v1,v2,v3,v4,v5) VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING"
	sqlUpdateRowPostgreSQL = "UPDATE %s SET p_type=$1,v0=$2,v1=$3,v2=$4,v3=$5,v4=$6,v5=$7 WHERE p_type=$8 AND v0=$9 AND v1=$10 AND v2=$11 AND v3=$12 AND v4=$13 AND v5=$14"
	sqlDeleteRowPostgreSQL = "DELETE FROM %s WHERE p_type=$1 AND v0=$2 AND v1=$3 AND v2=$4 AND v3=$5 AND v4=$6 AND v5=$7"
)

// for SQLServer.
const (
	sqlPlaceholderSQLServer = "@p"
	sqlCreateTableSQLServer = `
CREATE TABLE %[1]s(
    p_type NVARCHAR(32)  DEFAULT '' NOT NULL,
    v0     NVARCHAR(255) DEFAULT '' NOT NULL,
    v1     NVARCHAR(255) DEFAULT '' NOT NULL,
    v2     NVARCHAR(255) DEFAULT '' NOT NULL,
    v3     NVARCHAR(255) DEFAULT '' NOT NULL,
    v4     NVARCHAR(255) DEFAULT '' NOT NULL,
    v5     NVARCHAR(255) DEFAULT '' NOT NULL
);
CREATE INDEX idx_%[1]s ON %[1]s (p_type,v0,v1);`
	sqlInsertRowSQLServer = "INSERT INTO %s (p_type,v0,v1,v2,v3,v4,v5) VALUES (@p1,@p2,@p3,@p4,@p5,@p6,@p7)"
	sqlUpdateRowSQLServer = "UPDATE %s SET p_type=@p1,v0=@p2,v1=@p3,v2=@p4,v3=@p5,v4=@p6,v5=@p7 WHERE p_type=@p8 AND v0=@p9 AND v1=@p10 AND v2=@p11 AND v3=@p12 AND v4=@p13 AND v5=@p14"
	sqlDeleteRowSQLServer = "DELETE FROM %s WHERE p_type=@p1 AND v0=@p2 AND v1=@p3 AND v2=@p4 AND v3=@p5 AND v4=@p6 AND v5=@p7"
)
