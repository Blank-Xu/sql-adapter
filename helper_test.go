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
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

const (
	testEnvFile = "test.env"

	testRbacModelFile  = "examples/rbac_model.conf"
	testRbacPolicyFile = "examples/rbac_policy.csv"
)

var (
	testDBs = map[string]*sql.DB{}

	testDefaultPolicy = [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}}
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()

	teardown()
	os.Exit(code)
}

func setup() {
	driverNames := sql.Drivers()
	if len(driverNames) == 0 {
		log.Fatal("empty drivers")
	}

	envMap := loadEnvfile()

	testDataSources := map[string]string{
		"sqlite": "./test.db",
		"mysql": fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			envMap["TEST_DATABASE_USER"],
			envMap["TEST_DATABASE_PASSWORD"],
			envMap["TEST_DATABASE_HOST"],
			envMap["TEST_DATABASE_PORT_MYSQL"],
			envMap["TEST_DATABASE_NAME"]),
		"postgres": fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
			envMap["TEST_DATABASE_USER"],
			envMap["TEST_DATABASE_PASSWORD"],
			envMap["TEST_DATABASE_HOST"],
			envMap["TEST_DATABASE_PORT_POSTGRES"],
			envMap["TEST_DATABASE_NAME"]),
		"sqlserver": fmt.Sprintf("sqlserver://sa:%s@%s:%s?database=%s&encrypt=disable&connection+timeout=30",
			envMap["TEST_DATABASE_PASSWORD"],
			envMap["TEST_DATABASE_HOST"],
			envMap["TEST_DATABASE_PORT_SQLSERVER"],
			envMap["TEST_DATABASE_NAME"]),
	}

	// log.Println(testDataSources)

	for _, driverName := range driverNames {
		if driverName == "mssql" {
			continue
		}

		dataSourceName, ok := testDataSources[driverName]
		if !ok {
			log.Printf("driver name [%s] not found\n", driverName)
			continue
		}

		db, err := connDB(driverName, dataSourceName)
		if err != nil {
			log.Fatalf("connect to database failed, driver name: [%s], data source: [%s], err: %v", driverName, dataSourceName, err)
		}

		testDBs[driverName] = db
	}
}

func teardown() {
	for driverName, db := range testDBs {
		if err := db.Close(); err != nil {
			log.Printf("db close failed, driver name: [%s], err: %v\n", driverName, err)
		}
	}
}

func loadEnvfile() map[string]string {
	f, err := os.Open(testEnvFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	m := make(map[string]string)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || text[0] == '#' {
			continue
		}

		s := strings.Split(text, "=")
		if len(s) < 2 {
			panic("invalid env file: " + testEnvFile)
		}

		m[s[0]] = strings.Join(s[1:], "")
	}

	if err = scanner.Err(); err != nil {
		panic("load env file failed, err: " + err.Error())
	}

	return m
}

func connDB(driverName, dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("sql.Open failed, err: %v", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("db.Ping failed, err: %v", err)
	}

	return db, nil
}
