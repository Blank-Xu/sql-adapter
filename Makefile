#!make

include test/.env
export $(shell sed 's/=.*//' test/.env)

define run_unit_test
	@go test -v -race -covermode=atomic -coverprofile=coverage.out ./... -coverpkg=./...
endef

define run_e2e_test
	@echo "test start with DB[${1}]"
	@cd test && CGO_ENABLED=1 TEST_DB=${1} go test -v -race -covermode=atomic -coverprofile=coverage.out ./e2e/... -coverpkg=../...
	@if [ -f "coverage.out" ]; then tail -n +2 test/coverage.out >> coverage.out; else mv test/coverage.out coverage.out; fi
endef

define start_mysql
	@docker run --name mysql_${TEST_DATABASE_NAME} \
		-e "MYSQL_DATABASE=${TEST_DATABASE_NAME}" \
		-e "MYSQL_ROOT_PASSWORD=${TEST_DATABASE_PASSWORD}" \
		-e "MYSQL_USER=${TEST_DATABASE_USER}" \
		-e "MYSQL_PASSWORD=${TEST_DATABASE_PASSWORD}" \
		-p 3306:${TEST_DATABASE_PORT_MYSQL} \
		-d --rm mysql:latest
endef

define start_postgres
	@docker run --name postgres_${TEST_DATABASE_NAME} \
		-e "POSTGRES_DB=${TEST_DATABASE_NAME}" \
		-e "POSTGRES_USER=${TEST_DATABASE_USER}" \
		-e "POSTGRES_PASSWORD=${TEST_DATABASE_PASSWORD}" \
		-p 5432:${TEST_DATABASE_PORT_POSTGRES} \
		-d --rm postgres:latest
endef

define start_sqlserver
	@docker run --name sqlserver_${TEST_DATABASE_NAME} \
		-e "ACCEPT_EULA=Y" \
		-e "MSSQL_DB=${TEST_DATABASE_NAME}" \
		-e "SA_PASSWORD=${TEST_DATABASE_PASSWORD}" \
		-e "MSSQL_USER=${TEST_DATABASE_USER}" \
		-e "MSSQL_PASSWORD=${TEST_DATABASE_PASSWORD}" \
		-p 1433:${TEST_DATABASE_PORT_SQLSERVER} \
		-d --rm mcr.microsoft.com/mssql/server:2022-CU15-GDR1-ubuntu-22.04
	@sleep 5
	@docker exec -i sqlserver_${TEST_DATABASE_NAME} /bin/bash < test/init.sqlserver.sh
endef

run-unit-test:
	${call run_unit_test}

run-e2e-test:
	${call run_e2e_test}

test-all:
	${call run_unit_test}
	${call start_sqlserver}
	${call start_mysql}
	${call start_postgres}
	@sleep 10
	${call run_e2e_test}
	@docker stop mysql_${TEST_DATABASE_NAME}
	@docker stop postgres_${TEST_DATABASE_NAME}
	@docker stop sqlserver_${TEST_DATABASE_NAME}

test-sqlite:
	${call run_e2e_test,sqlite}

test-mysql:
	${call start_mysql}
	@sleep 10
	${call run_e2e_test,mysql}
	@docker stop mysql_${TEST_DATABASE_NAME}

test-postgres:
	${call start_postgres}
	@sleep 10
	${call run_e2e_test,postgres}
	@docker stop postgres_${TEST_DATABASE_NAME}

test-sqlserver:
	${call start_sqlserver}
	@sleep 5
	${call run_e2e_test,sqlserver}
	@docker stop sqlserver_${TEST_DATABASE_NAME}

lint:
	golangci-lint run -v ./...

fmt:
	goimports -l -w -d -e .

build:
	go build -v .
