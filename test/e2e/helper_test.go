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

package sqladaptertest

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/microsoft/go-mssqldb"
)

const (
	testEnvFile = "../.env"
)

var (
	testDBs = map[string]*sql.DB{}
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
		log.Fatal("empty sql drivers")
	}

	envMap := loadEnvfile()
	dataSources := getDataSources(envMap)

	if db := os.Getenv("TEST_DB"); db != "" {
		log.Println(db)
		loadDB(dataSources, db)
		return
	}

	for _, driverName := range driverNames {
		if driverName == "mssql" {
			continue
		}

		loadDB(dataSources, driverName)
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

func getDataSources(envMap map[string]string) map[string]string {
	return map[string]string{
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

func loadDB(dataSources map[string]string, driverName string) {
	dataSourceName, ok := dataSources[driverName]
	if !ok {
		log.Printf("driver name [%s] not found\n", driverName)
		return
	}

	db, err := connDB(driverName, dataSourceName)
	if err != nil {
		log.Fatalf("connect to database failed, driver name: [%s], data source: [%s], err: %v", driverName, dataSourceName, err)
	}

	testDBs[driverName] = db
}
