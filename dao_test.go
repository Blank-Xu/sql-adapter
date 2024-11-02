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
	"context"
	"testing"
)

func TestDao(t *testing.T) {
	for driverName, db := range testDBs {
		t.Logf("[%s] test dao start", driverName)

		d, err := getDao(db, driverName, "sqladapter_test_dao")
		if err != nil {
			t.Errorf("getDao failed, err: %v", err)
		}

		t.Run(driverName, func(t *testing.T) {
			testSQL(t, d)
		})
	}
}

func testSQL(t *testing.T, dao dao) {
	ctx := context.TODO()

	err := dao.CreateTable(ctx)
	if err != nil {
		t.Errorf("create table failed, err %v", err)
	}

	if !dao.IsTableExist(ctx) {
		t.Errorf("table[%s] not exist", dao.TableName())
	}

	testInsert(t, dao)
	testUpdate(t, dao)
	testSelect(t, dao)
	testDelete(t, dao)
}

func testInsert(t *testing.T, dao dao) {
	// todo
}

func testUpdate(t *testing.T, dao dao) {
	// todo
}

func testSelect(t *testing.T, dao dao) {
	// todo
}

func testDelete(t *testing.T, dao dao) {
	// todo
}
