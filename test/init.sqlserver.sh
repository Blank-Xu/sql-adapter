#!/bin/bash

# for local test
export PATH=/opt/mssql-tools18/bin:$PATH


# waiting for MSSQL server start
export STATUS=1
i=0

while [[ $STATUS -ne 0 ]] && [[ $i -lt 30 ]]; do
	i=$i+1
	sqlcmd -t 1 -U sa -P $MSSQL_SA_PASSWORD -C -Q "select 1" >> /dev/null
	STATUS=$?
done

if [ $STATUS -ne 0 ]; then 
	echo "Error: MSSQL SERVER took more than thirty seconds to start up."
	exit 1
fi
echo "-------------------------- MSSQL SERVER STARTED --------------------------"


# prepare database
INIT_SQL=$(cat << EOF
CREATE DATABASE $MSSQL_DB;
GO
USE $MSSQL_DB;
GO
CREATE LOGIN $MSSQL_USER WITH PASSWORD = '$MSSQL_PASSWORD';
GO
CREATE USER $MSSQL_USER FOR LOGIN $MSSQL_USER;
GO
ALTER SERVER ROLE sysadmin ADD MEMBER [$MSSQL_USER];
GO
EOF
)

sqlcmd -S localhost -U sa -P $MSSQL_SA_PASSWORD -C -d master -Q "$INIT_SQL"
echo "---------------- MSSQL CREATE DATABASE $MSSQL_DB SUCCESSFULLY ---------------- "
