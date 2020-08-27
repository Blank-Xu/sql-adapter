# sql-adapter

[![Go Report Card](https://goreportcard.com/badge/github.com/Blank-Xu/sql-adapter)](https://goreportcard.com/report/github.com/Blank-Xu/sql-adapter)
[![Build Status](https://travis-ci.org/Blank-Xu/sql-adapter.svg?branch=master)](https://travis-ci.org/Blank-Xu/sql-adapter)
[![Coverage Status](https://coveralls.io/repos/github/Blank-Xu/sql-adapter/badge.svg?branch=master)](https://coveralls.io/github/Blank-Xu/sql-adapter?branch=master)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Blank-Xu/sql-adapter)](https://pkg.go.dev/github.com/Blank-Xu/sql-adapter)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
---

sql-adapter is a database/sql Adapter for [Casbin V2](https://github.com/casbin/casbin/v2). 

With this library, Casbin can load policy lines or save policy lines from supported databases.


## Tested Databases
### `master` branch
- SQLite3: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
- Mysql(v5.5): [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)
- MariaDB(v10.2): [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)
- Postgres(v9.6): [github.com/lib/pq](https://github.com/lib/pq)
- SqlServer(v2008R2-SP3): [github.com/denisenkom/go-mssqldb](https://github.com/denisenkom/go-mssqldb)

### `oracle` branch
- Oracle(v11.2): [github.com/mattn/go-oci8](https://github.com/mattn/go-oci8)


## Installation

    go get github.com/Blank-Xu/sql-adapter


## Simple Examples
### Sqlite3
```go
package main

import (
	"database/sql"
	"log" 

	sqladapter "github.com/Blank-Xu/sql-adapter"
	"github.com/casbin/casbin/v2"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// connect to the database first.
	db, err := sql.Open("sqlite3", "file:test.db")
	if err != nil {
		panic(err)
	}
	if err = db.Ping();err!=nil{
		panic(err)
	}

	// Initialize an adapter and use it in a Casbin enforcer:
	// The adapter will use the Sqlite3 table name "casbin_rule_test",
	// the default table name is "casbin_rule".
	// If it doesn't exist, the adapter will create it automatically.
	a, err := sqladapter.NewAdapter(db, "sqlite3", "casbin_rule_test")
	if err != nil {
		panic(err)
	}

	e, err := casbin.NewEnforcer("examples/rbac_model.conf", a)
	if err != nil {
		panic(err)
	}

	// Load the policy from DB.
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
```


### Mysql

```go
package main

import (
	"database/sql"
	"log"
	"runtime"
	"time"

	sqladapter "github.com/Blank-Xu/sql-adapter"
	"github.com/casbin/casbin/v2"
	_ "github.com/go-sql-driver/mysql"
)

func finalizer(db *sql.DB) {
	err := db.Close()
	if err != nil {
		panic(err)
	}
}

func main() {
	// connect to the database first.
	db, err := sql.Open("mysql", "root:YourPassword@tcp(127.0.0.1:3306)/YourDBName")
	if err != nil {
		panic(err)
	}
	if err = db.Ping();err!=nil{
		panic(err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Minute * 10)

	// need to control by user, not the package
	runtime.SetFinalizer(db, finalizer)

	// Initialize an adapter and use it in a Casbin enforcer:
	// The adapter will use the Sqlite3 table name "casbin_rule_test",
	// the default table name is "casbin_rule".
	// If it doesn't exist, the adapter will create it automatically.
	a, err := sqladapter.NewAdapter(db, "mysql", "casbin_rule_test")
	if err != nil {
		panic(err)
	}

	e, err := casbin.NewEnforcer("examples/rbac_model.conf", a)
	if err != nil {
		panic(err)
	}

	// Load the policy from DB.
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
```


## Getting Help

- [Casbin](https://github.com/casbin/casbin)


## License

This project is under Apache 2.0 License. See the [LICENSE](LICENSE) file for the full license text.
