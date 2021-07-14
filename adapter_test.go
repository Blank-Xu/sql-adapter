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
	"database/sql"
	"strings"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/util"

	// _ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	// _ "github.com/mattn/go-sqlite3"
)

const (
	rbacModelFile  = "examples/rbac_model.conf"
	rbacPolicyFile = "examples/rbac_policy.csv"
)

var (
	dataSourceNames = map[string]string{
		// "sqlite3": ":memory:",
		"mysql":    "root:@tcp(127.0.0.1:3306)/sql_adapter_test",
		"postgres": "user=postgres host=127.0.0.1 port=5432 dbname=sql_adapter_test sslmode=disable",
		// "sqlserver": "sqlserver://sa:YourPassword@127.0.0.1:1433?database=sql_adapter_test&connection+timeout=30",
	}

	lines = []CasbinRule{
		{PType: "p", V0: "alice", V1: "data1", V2: "read"},
		{PType: "p", V0: "bob", V1: "data2", V2: "read"},
		{PType: "p", V0: "bob", V1: "data2", V2: "write"},
		{PType: "p", V0: "data2_admin", V1: "data1", V2: "read", V3: "test1", V4: "test2", V5: "test3"},
		{PType: "p", V0: "data2_admin", V1: "data2", V2: "write", V3: "test1", V4: "test2", V5: "test3"},
		{PType: "p", V0: "data1_admin", V1: "data2", V2: "write"},
		{PType: "g", V0: "alice", V1: "data2_admin"},
		{PType: "g", V0: "bob", V1: "data2_admin", V2: "test"},
		{PType: "g", V0: "bob", V1: "data1_admin", V2: "test2", V3: "test3", V4: "test4", V5: "test5"},
	}

	filter = Filter{
		PType: []string{"p"},
		V0:    []string{"bob", "data2_admin"},
		V1:    []string{"data1", "data2"},
		V2:    []string{"read", "write"},
		V3:    []string{"test1"},
		V4:    []string{"test2"},
		V5:    []string{"test3"},
	}
)

func TestAdapters(t *testing.T) {
	for key, value := range dataSourceNames {
		t.Logf("-------------------- test [%s] start, dataSourceName: [%s]", key, value)

		db, err := sql.Open(key, value)
		if err != nil {
			t.Fatalf("sql.Open failed, err: %v", err)
		}
		if err = db.Ping(); err != nil {
			t.Fatalf("db.Ping failed, err: %v", err)
		}

		t.Log("---------- testTableName start")
		testTableName(t, db, key)
		t.Log("---------- testTableName finished")

		t.Log("---------- testSQL start")
		testSQL(t, db, key, "sqladapter_sql")
		t.Log("---------- testSQL finished")

		t.Log("---------- testSaveLoad start")
		testSaveLoad(t, db, key, "sqladapter_save_load")
		t.Log("---------- testSaveLoad finished")

		t.Log("---------- testAutoSave start")
		testAutoSave(t, db, key, "sqladapter_auto_save")
		t.Log("---------- testAutoSave finished")

		t.Log("---------- testFilteredPolicy start")
		testFilteredPolicy(t, db, key, "sqladapter_filtered_policy")
		t.Log("---------- testFilteredPolicy finished")

		t.Log("---------- testUpdatePolicy start")
		testUpdatePolicy(t, db, key, "sqladapter_filtered_policy")
		t.Log("---------- testUpdatePolicy finished")

		t.Log("---------- testUpdatePolicies start")
		testUpdatePolicies(t, db, key, "sqladapter_filtered_policy")
		t.Log("---------- testUpdatePolicies finished")

		t.Log("---------- testUpdateFilteredPolicies start")
		testUpdateFilteredPolicies(t, db, key, "sqladapter_filtered_policy")
		t.Log("---------- testUpdateFilteredPolicies finished")
	}
}

func testTableName(t *testing.T, db *sql.DB, driverName string) {
	_, err := NewAdapter(db, driverName, "")
	if err != nil {
		t.Fatalf("NewAdapter failed, err: %v", err)
	}
}

func testSQL(t *testing.T, db *sql.DB, driverName, tableName string) {
	var err error
	logErr := func(action string) {
		if err != nil {
			t.Errorf("%s test failed, err: %v", action, err)
		}
	}

	equalValue := func(line1, line2 CasbinRule) bool {
		if line1.PType != line2.PType ||
			line1.V0 != line2.V0 ||
			line1.V1 != line2.V1 ||
			line1.V2 != line2.V2 ||
			line1.V3 != line2.V3 ||
			line1.V4 != line2.V4 ||
			line1.V5 != line2.V5 {
			return false
		}
		return true
	}

	var a *Adapter
	a, err = NewAdapter(db, driverName, tableName)
	logErr("NewAdapter")

	// createTable test has passed when adapter create
	// logErr("createTable",  a.createTable())

	if b := a.isTableExist(); b == false {
		t.Fatal("isTableExist test failed")
	}

	rules := make([][]interface{}, len(lines))
	for idx, rule := range lines {
		args := a.genArgs(rule.PType, []string{rule.V0, rule.V1, rule.V2, rule.V3, rule.V4, rule.V5})
		rules[idx] = args
	}

	err = a.truncateAndInsertRows(rules)
	logErr("truncateAndInsertRows")

	err = a.deleteAllAndInsertRows(rules)
	logErr("truncateAndInsertRows")

	err = a.deleteRows(a.sqlDeleteByArgs, "g")
	logErr("deleteRows sqlDeleteByArgs g")

	err = a.deleteRows(a.sqlDeleteAll)
	logErr("deleteRows sqlDeleteAll")

	_ = a.truncateAndInsertRows(rules)

	records, err := a.selectRows(a.sqlSelectAll)
	logErr("selectRows sqlSelectAll")
	for idx, record := range records {
		line := lines[idx]
		if !equalValue(*record, line) {
			t.Fatalf("selectRows records test not equal, query record: %+v, need record: %+v", record, line)
		}
	}

	records, err = a.selectWhereIn(&filter)
	logErr("selectWhereIn")
	i := 3
	for _, record := range records {
		line := lines[i]
		if !equalValue(*record, line) {
			t.Fatalf("selectWhereIn records test not equal, query record: %+v, need record: %+v", record, line)
		}
		i++
	}

	err = a.truncateTable()
	logErr("truncateTable")
}

func initPolicy(t *testing.T, db *sql.DB, driverName, tableName string) {
	// Because the DB is empty at first,
	// so we need to load the policy from the file adapter (.CSV) first.
	e, _ := casbin.NewEnforcer(rbacModelFile, rbacPolicyFile)

	a, err := NewAdapter(db, driverName, tableName)
	if err != nil {
		t.Fatal("NewAdapter test failed, err: ", err)
	}

	// This is a trick to save the current policy to the DB.
	// We can't call e.SavePolicy() because the adapter in the enforcer is still the file adapter.
	// The current policy means the policy in the Casbin enforcer (aka in memory).
	err = a.SavePolicy(e.GetModel())
	if err != nil {
		t.Fatal("SavePolicy test failed, err: ", err)
	}

	// Clear the current policy.
	e.ClearPolicy()
	testGetPolicy(t, e, [][]string{})

	// Load the policy from DB.
	err = a.LoadPolicy(e.GetModel())
	if err != nil {
		t.Fatal("LoadPolicy test failed, err: ", err)
	}
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}

func testSaveLoad(t *testing.T, db *sql.DB, driverName, tableName string) {
	// Initialize some policy in DB.
	initPolicy(t, db, driverName, tableName)
	// Note: you don't need to look at the above code
	// if you already have a working DB with policy inside.

	// Now the DB has policy, so we can provide a normal use case.
	// Create an adapter and an enforcer.
	// NewEnforcer() will load the policy automatically.
	a, _ := NewAdapter(db, driverName, tableName)
	e, _ := casbin.NewEnforcer(rbacModelFile, a)
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}

func testAutoSave(t *testing.T, db *sql.DB, driverName, tableName string) {
	// Initialize some policy in DB.
	initPolicy(t, db, driverName, tableName)
	// Note: you don't need to look at the above code
	// if you already have a working DB with policy inside.

	// Now the DB has policy, so we can provide a normal use case.
	// Create an adapter and an enforcer.
	// NewEnforcer() will load the policy automatically.
	a, _ := NewAdapter(db, driverName, tableName)
	e, _ := casbin.NewEnforcer(rbacModelFile, a)

	// AutoSave is enabled by default.
	// Now we disable it.
	e.EnableAutoSave(false)

	var err error
	logErr := func(action string) {
		if err != nil {
			t.Errorf("%s test failed, err: %v", action, err)
		}
	}

	// Because AutoSave is disabled, the policy change only affects the policy in Casbin enforcer,
	// it doesn't affect the policy in the storage.
	_, err = e.AddPolicy("alice", "data1", "write")
	logErr("AddPolicy1")
	// Reload the policy from the storage to see the effect.
	err = e.LoadPolicy()
	logErr("LoadPolicy1")
	// This is still the original policy.
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	_, err = e.AddPolicies([][]string{{"alice_1", "data_1", "read_1"}, {"bob_1", "data_1", "write_1"}})
	logErr("AddPolicies1")
	// Reload the policy from the storage to see the effect.
	err = e.LoadPolicy()
	logErr("LoadPolicy2")
	// This is still the original policy.
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	// Now we enable the AutoSave.
	e.EnableAutoSave(true)

	// Because AutoSave is enabled, the policy change not only affects the policy in Casbin enforcer,
	// but also affects the policy in the storage.
	_, err = e.AddPolicy("alice", "data1", "write")
	logErr("AddPolicy2")
	// Reload the policy from the storage to see the effect.
	err = e.LoadPolicy()
	logErr("LoadPolicy3")
	// The policy has a new rule: {"alice", "data1", "write"}.
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}, {"alice", "data1", "write"}})

	_, err = e.AddPolicies([][]string{{"alice_2", "data_2", "read_2"}, {"bob_2", "data_2", "write_2"}})
	logErr("AddPolicies2")
	// Reload the policy from the storage to see the effect.
	err = e.LoadPolicy()
	logErr("LoadPolicy4")
	// This is still the original policy.
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}, {"alice", "data1", "write"},
		{"alice_2", "data_2", "read_2"}, {"bob_2", "data_2", "write_2"}})

	_, err = e.RemovePolicies([][]string{{"alice_2", "data_2", "read_2"}, {"bob_2", "data_2", "write_2"}})
	logErr("RemovePolicies")
	err = e.LoadPolicy()
	logErr("LoadPolicy5")
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}, {"alice", "data1", "write"}})

	// Remove the added rule.
	_, err = e.RemovePolicy("alice", "data1", "write")
	logErr("RemovePolicy")
	err = e.LoadPolicy()
	logErr("LoadPolicy6")
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	// Remove "data2_admin" related policy rules via a filter.
	// Two rules: {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"} are deleted.
	_, err = e.RemoveFilteredPolicy(0, "data2_admin")
	logErr("RemoveFilteredPolicy")
	err = e.LoadPolicy()
	logErr("LoadPolicy7")
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}})
}

func testFilteredPolicy(t *testing.T, db *sql.DB, driverName, tableName string) {
	// Initialize some policy in DB.
	initPolicy(t, db, driverName, tableName)
	// Note: you don't need to look at the above code
	// if you already have a working DB with policy inside.

	// Now the DB has policy, so we can provide a normal use case.
	// Create an adapter and an enforcer.
	// NewEnforcer() will load the policy automatically.
	a, _ := NewAdapter(db, driverName, tableName)
	e, _ := casbin.NewEnforcer(rbacModelFile, a)
	// Now set the adapter
	e.SetAdapter(a)

	var err error
	logErr := func(action string) {
		if err != nil {
			t.Errorf("%s test failed, err: %v", action, err)
		}
	}

	// Load only alice's policies
	err = e.LoadFilteredPolicy(&Filter{V0: []string{"alice"}})
	logErr("LoadFilteredPolicy alice")
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}})

	// Load only bob's policies
	err = e.LoadFilteredPolicy(&Filter{V0: []string{"bob"}})
	logErr("LoadFilteredPolicy bob")
	testGetPolicy(t, e, [][]string{{"bob", "data2", "write"}})

	// Load policies for data2_admin
	err = e.LoadFilteredPolicy(&Filter{V0: []string{"data2_admin"}})
	logErr("LoadFilteredPolicy data2_admin")
	testGetPolicy(t, e, [][]string{{"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	// Load policies for alice and bob
	err = e.LoadFilteredPolicy(&Filter{V0: []string{"alice", "bob"}})
	logErr("LoadFilteredPolicy alice bob")
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}})

	_, err = e.AddPolicy("bob", "data1", "write", "test1", "test2", "test3")
	logErr("AddPolicy")

	err = e.LoadFilteredPolicy(&filter)
	logErr("LoadFilteredPolicy filter")
	testGetPolicy(t, e, [][]string{{"bob", "data1", "write", "test1", "test2", "test3"}})
}

func testUpdatePolicy(t *testing.T, db *sql.DB, driverName, tableName string) {
	// Initialize some policy in DB.
	initPolicy(t, db, driverName, tableName)

	a, _ := NewAdapter(db, driverName, tableName)
	e, _ := casbin.NewEnforcer(rbacModelFile, a)

	e.EnableAutoSave(true)
	e.UpdatePolicy([]string{"alice", "data1", "read"}, []string{"alice", "data1", "write"})
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "write"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}

func testUpdatePolicies(t *testing.T, db *sql.DB, driverName, tableName string) {
	// Initialize some policy in DB.
	initPolicy(t, db, driverName, tableName)

	a, _ := NewAdapter(db, driverName, tableName)
	e, _ := casbin.NewEnforcer(rbacModelFile, a)

	e.EnableAutoSave(true)
	e.UpdatePolicies([][]string{{"alice", "data1", "write"}, {"bob", "data2", "write"}}, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "read"}})
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "read"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}

func testUpdateFilteredPolicies(t *testing.T, db *sql.DB, driverName, tableName string) {
	// Initialize some policy in DB.
	initPolicy(t, db, driverName, tableName)

	a, _ := NewAdapter(db, driverName, tableName)
	e, _ := casbin.NewEnforcer(rbacModelFile, a)

	e.EnableAutoSave(true)
	e.UpdateFilteredPolicies([][]string{{"alice", "data1", "write"}}, 0, "alice", "data1", "read")
	e.UpdateFilteredPolicies([][]string{{"bob", "data2", "read"}}, 0, "bob", "data2", "write")
	e.LoadPolicy()
	testGetPolicyWithoutOrder(t, e, [][]string{{"alice", "data1", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}, {"bob", "data2", "read"}})
}

func testGetPolicy(t *testing.T, e *casbin.Enforcer, res [][]string) {
	t.Helper()
	myRes := e.GetPolicy()
	t.Log("Policy: ", myRes)

	m := make(map[string]struct{}, len(myRes))
	for _, record := range myRes {
		key := strings.Join(record, ",")
		m[key] = struct{}{}
	}

	for _, record := range res {
		key := strings.Join(record, ",")
		if _, ok := m[key]; !ok {
			t.Error("Policy: \n", myRes, ", supposed to be \n", res)
			break
		}
	}
}

func testGetPolicyWithoutOrder(t *testing.T, e *casbin.Enforcer, res [][]string) {
	myRes := e.GetPolicy()
	// log.Print("Policy: \n", myRes)

	if !arrayEqualsWithoutOrder(myRes, res) {
		t.Error("Policy: \n", myRes, ", supposed to be \n", res)
	}
}

func arrayEqualsWithoutOrder(a [][]string, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}

	mapA := make(map[int]string)
	mapB := make(map[int]string)
	order := make(map[int]struct{})
	l := len(a)

	for i := 0; i < l; i++ {
		mapA[i] = util.ArrayToString(a[i])
		mapB[i] = util.ArrayToString(b[i])
	}

	for i := 0; i < l; i++ {
		for j := 0; j < l; j++ {
			if _, ok := order[j]; ok {
				if j == l-1 {
					return false
				} else {
					continue
				}
			}
			if mapA[i] == mapB[j] {
				order[j] = struct{}{}
				break
			} else if j == l-1 {
				return false
			}
		}
	}
	return true
}
