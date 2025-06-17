# sql-adapter

[![Go Report Card](https://goreportcard.com/badge/github.com/Blank-Xu/sql-adapter)](https://goreportcard.com/report/github.com/Blank-Xu/sql-adapter)
[![Build Status](https://github.com/Blank-Xu/sql-adapter/actions/workflows/tests.yaml/badge.svg)](https://github.com/Blank-Xu/sql-adapter/actions)
[![Coverage Status](https://coveralls.io/repos/github/Blank-Xu/sql-adapter/badge.svg?branch=master)](https://coveralls.io/github/Blank-Xu/sql-adapter?branch=master)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Blank-Xu/sql-adapter)](https://pkg.go.dev/github.com/Blank-Xu/sql-adapter)
[![Release](https://img.shields.io/github/release/Blank-Xu/sql-adapter.svg)](https://github.com/Blank-Xu/sql-adapter/releases/latest)
[![Sourcegraph](https://sourcegraph.com/github.com/Blank-Xu/sql-adapter/-/badge.svg)](https://sourcegraph.com/github.com/Blank-Xu/sql-adapter?badge)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

---

The `sql-adapter` is a `database/sql` adapter for [Casbin v2](https://github.com/casbin/casbin).

With this library, Casbin can load policy lines or save policy lines from supported databases.

## Test Databases

### `master` branch

- SQLite3: [modernc.org/sqlite](https://modernc.org/sqlite)
- MySQL(v8): [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)
- PostgreSQL(v15): [github.com/lib/pq](https://github.com/lib/pq)
- SQL Server(v2017): [github.com/microsoft/go-mssqldb](https://github.com/microsoft/go-mssqldb)

### `oracle` branch

- Oracle(v11.2): [github.com/mattn/go-oci8](https://github.com/mattn/go-oci8)

## Installation

```shell
go get github.com/Blank-Xu/sql-adapter
```

## Examples

- [database/sql](https://github.com/Blank-Xu/sql-adapter/blob/master/examples/database_sql)
- [sqlx](https://github.com/Blank-Xu/sql-adapter/tree/master/examples/sqlx)
- [xorm](https://github.com/Blank-Xu/sql-adapter/tree/master/examples/xorm)
- [gorm](https://github.com/Blank-Xu/sql-adapter/tree/master/examples/gorm)

### Simple example for MySQL

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

func main() {
    // connect to the database first.
    db, err := sql.Open("mysql", "YourUserName:YourPassword@tcp(127.0.0.1:3306)/YourDBName?charset=utf8")
    if err != nil {
        panic(err)
    }
    if err = db.Ping();err!=nil{
        panic(err)
    }
    defer db.Close()

    db.SetMaxOpenConns(20)
    db.SetMaxIdleConns(10)
    db.SetConnMaxLifetime(time.Minute * 10)

    // Initialize an adapter and use it in a Casbin enforcer:
    // The adapter will use the MySQL table name "casbin_rule_test",
    // the default table name is "casbin_rule" if it is not given.
    // If it doesn't exist, the adapter will create it automatically.
    a, err := sqladapter.NewAdapter(db, "mysql", "casbin_rule_test")
    if err != nil {
        panic(err)
    }

    e, err := casbin.NewEnforcer("test/testdata/rbac_model.conf", a)
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
```

## Getting Help

- [Casbin](https://github.com/casbin/casbin)

## License

This project is under Apache 2.0 License. See the [LICENSE](LICENSE) file for the full license text.
