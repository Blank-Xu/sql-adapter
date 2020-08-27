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

// for Oracle
var (
	sqlPlaceHolder = []byte(":arg")
)

const (
	sqlAND = " AND "

	sqlCreateTable = `
CREATE TABLE %s(
	P_TYPE NVARCHAR2(32) DEFAULT '' NOT NULL,
    V0     NVARCHAR2(255) DEFAULT '' NOT NULL,
    V1     NVARCHAR2(255) DEFAULT '' NOT NULL,
    V2     NVARCHAR2(255),
    V3     NVARCHAR2(255),
    V4     NVARCHAR2(255),
    V5     NVARCHAR2(255)
)`
	sqlCreateIndex = `CREATE INDEX IDX_%[1]s ON %[1]s (P_TYPE, V0, V1)`

	sqlTruncateTable = "TRUNCATE TABLE %s"
	sqlIsTableExist  = "SELECT 1 FROM %s"

	sqlInsertRow    = "INSERT INTO %s "
	sqlDeleteAll    = "DELETE FROM %s"
	sqlDeleteByArgs = "DELETE FROM %s WHERE P_TYPE = :arg1"

	sqlSelectAll   = "SELECT * FROM %s"
	sqlSelectWhere = "SELECT * FROM %s WHERE "
)
