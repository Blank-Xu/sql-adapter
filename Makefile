#!make

include test.env
export $(shell sed 's/=.*//' test.env)

define import_driver
	@echo 'import driver ${1}'
	@sed -i '16 i import _ "${1}"' helper_test.go
	@go get -u ${1}@${2}
endef

define delete_driver
	@sed -i 's#'${1}'#${DELETE_DRIVER_FLAG}#;/${DELETE_DRIVER_FLAG}/d' helper_test.go
endef

define run_test
	@echo "test start"
	@go mod tidy
	@CGO_ENABLED=1 go test -v -race -covermode=atomic -coverprofile=coverage.out ./...
endef

define start_mysql
	@docker run --name mysql_${TEST_DATABASE_NAME} \
		-p 3306:${TEST_DATABASE_PORT_MYSQL} \
		-e "MYSQL_DATABASE=${TEST_DATABASE_NAME}" \
		-e "MYSQL_ROOT_PASSWORD=${TEST_DATABASE_PASSWORD}" \
		-e "MYSQL_USER=${TEST_DATABASE_USER}" \
		-e "MYSQL_PASSWORD=${TEST_DATABASE_PASSWORD}" \
		-d --rm mysql:latest
endef

define start_postgres
	@docker run --name postgres_${TEST_DATABASE_NAME} \
		-p 5432:${TEST_DATABASE_PORT_POSTGRES} \
		-e "POSTGRES_DB=${TEST_DATABASE_NAME}" \
		-e "POSTGRES_USER=${TEST_DATABASE_USER}" \
		-e "POSTGRES_PASSWORD=${TEST_DATABASE_PASSWORD}" \
		-d --rm postgres:latest
endef

define start_sqlserver
	@docker run --name sqlserver_${TEST_DATABASE_NAME} \
		-p 1433:${TEST_DATABASE_PORT_SQLSERVER} \
		-e "ACCEPT_EULA=Y" \
		-e "MSSQL_DB=${TEST_DATABASE_NAME}" \
		-e "SA_PASSWORD=${TEST_DATABASE_PASSWORD}" \
		-e "MSSQL_USER=${TEST_DATABASE_USER}" \
		-e "MSSQL_PASSWORD=${TEST_DATABASE_PASSWORD}" \
		-d --rm mcmoe/mssqldocker:latest
endef


.PHONY: clean-drivers
clean-drivers:
	@echo "clean drivers"
	${call delete_driver,${DRIVER_SQLITE}}
	${call delete_driver,${DRIVER_MYSQL}}
	${call delete_driver,${DRIVER_POSTGRES}}
	${call delete_driver,${DRIVER_SQLSERVER}}

clean: clean-drivers
	@go mod tidy
	@go fmt ./...

run-test: clean-drivers
	${call import_driver,${DRIVER_SQLITE},${DRIVER_SQLITE_VERSION}}
	${call import_driver,${DRIVER_MYSQL},${DRIVER_MYSQL_VERSION}}
	${call import_driver,${DRIVER_POSTGRES},${DRIVER_POSTGRES_VERSION}}
	${call import_driver,${DRIVER_SQLSERVER},${DRIVER_SQLSERVER_VERSION}}
	${call run_test}

test-all: clean-drivers
	${call start_mysql}
	${call start_postgres}
	${call start_sqlserver}
	@sleep 10
	${call import_driver,${DRIVER_SQLITE},${DRIVER_SQLITE_VERSION}}
	${call import_driver,${DRIVER_MYSQL},${DRIVER_MYSQL_VERSION}}
	${call import_driver,${DRIVER_POSTGRES},${DRIVER_POSTGRES_VERSION}}
	${call import_driver,${DRIVER_SQLSERVER},${DRIVER_SQLSERVER_VERSION}}
	${call run_test}
	@docker stop mysql_${TEST_DATABASE_NAME}
	@docker stop postgres_${TEST_DATABASE_NAME}
	@docker stop sqlserver_${TEST_DATABASE_NAME}

test-sqlite: clean-drivers
	${call import_driver,${DRIVER_SQLITE},${DRIVER_SQLITE_VERSION}}
	${call run_test}

test-mysql: clean-drivers
	${call start_mysql}
	@sleep 10
	${call import_driver,${DRIVER_MYSQL},${DRIVER_MYSQL_VERSION}}
	${call run_test}
	@docker stop mysql_${TEST_DATABASE_NAME}

test-postgres: clean-drivers
	${call start_postgres}
	@sleep 10
	${call import_driver,${DRIVER_POSTGRES},${DRIVER_POSTGRES_VERSION}}
	${call run_test}
	@docker stop postgres_${TEST_DATABASE_NAME}

test-sqlserver: clean-drivers
	${call start_sqlserver}
	@sleep 10
	${call import_driver,${DRIVER_SQLSERVER},${DRIVER_SQLSERVER_VERSION}}
	${call run_test}
	@docker stop sqlserver_${TEST_DATABASE_NAME}

lint:
	golangci-lint run ./...

fmt:
	goimports -l -w -d -e .

build:
	go build -v .
