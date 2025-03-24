module sql-example

go 1.21
toolchain go1.24.1

replace github.com/Blank-Xu/sql-adapter => ../../.

require (
	github.com/Blank-Xu/sql-adapter v0.0.0-00010101000000-000000000000
	github.com/casbin/casbin/v2 v2.103.0
	github.com/go-sql-driver/mysql v1.9.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.7.1 // indirect
	github.com/casbin/govaluate v1.3.0 // indirect
)
