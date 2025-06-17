package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	sqladapter "github.com/Blank-Xu/sql-adapter"
	"github.com/casbin/casbin/v2"

	"xorm.io/xorm"
)

const (
	driverName = "mysql"
	tableName  = "casbin_rule_example"

	envFile = "../test/.env"

	rbacModelFile = "../test/testdata/rbac_model.conf"
)

var dataSource = func() string {
	envMap := loadEnvfile(envFile)

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8",
		envMap["TEST_DATABASE_USER"],
		envMap["TEST_DATABASE_PASSWORD"],
		envMap["TEST_DATABASE_HOST"],
		envMap["TEST_DATABASE_PORT_MYSQL"],
		envMap["TEST_DATABASE_NAME"])
}()

func main() {
	// connect to the database first.
	engine, err := xorm.NewEngine(driverName, dataSource)
	if err != nil {
		panic(err)
	}
	if err = engine.Ping(); err != nil {
		panic(err)
	}
	defer engine.Close()

	engine.SetMaxOpenConns(20)
	engine.SetMaxIdleConns(10)
	engine.SetConnMaxLifetime(time.Minute * 10)

	// Initialize an adapter and use it in a Casbin enforcer:
	// The adapter will use the MySQL table name "casbin_rule_example",
	// the default table name is "casbin_rule".
	// If it doesn't exist, the adapter will create it automatically.
	a, err := sqladapter.NewAdapter(engine.DB().DB, driverName, tableName)
	if err != nil {
		panic(err)
	}

	e, err := casbin.NewEnforcer(rbacModelFile, a)
	if err != nil {
		panic(err)
	}

	// Load the policies from DB.
	if err = e.LoadPolicy(); err != nil {
		log.Println("LoadPolicy failed, err: ", err)
	}

	// Check the permission.
	has, err := e.Enforce("alice", "data1", "read")
	if err != nil {
		log.Println("Enforce failed, err: ", err)
	}
	if !has {
		log.Println("do not have permission")
	}

	// Modify the policy.
	// e.AddPolicy(...)
	// e.RemovePolicy(...)

	// Save the policy back to DB.
	if err = e.SavePolicy(); err != nil {
		log.Println("SavePolicy failed, err: ", err)
	}
}

func loadEnvfile(envFile string) map[string]string {
	f, err := os.Open(envFile)
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
			panic("invalid env file: " + envFile)
		}

		m[s[0]] = strings.Join(s[1:], "")
	}

	if err = scanner.Err(); err != nil {
		panic("load env file failed, err: " + err.Error())
	}

	return m
}
