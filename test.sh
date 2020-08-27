#!/usr/bin/env bash

set -e

echo "installing build tools"
apt-get -qq -y update  2>&1 > /dev/null
apt-get -qq -y install git pkg-config gcc wget  2>&1 > /dev/null


echo "installing go"
cd /tmp/
wget -nv https://dl.google.com/go/go1.11.13.linux-amd64.tar.gz

mkdir -p /usr/local/goFiles1.11.x
tar -xf /tmp/go1.11.13.linux-amd64.tar.gz
mv /tmp/go /usr/local/go1.11.x


echo "starting Oracle"
/usr/sbin/startup.sh


echo "setting up Oracle"
export ORACLE_HOME=/u01/app/oracle/product/11.2.0/xe
export PATH=$ORACLE_HOME/bin:$PATH
export ORACLE_SID=XE
export LD_LIBRARY_PATH=/u01/app/oracle/product/11.2.0/xe/lib

DOCKER_IP=$(ifconfig eth0 | awk '/inet / { printf $2; exit }')
PORT="1521"

tnsping ${DOCKER_IP}

sqlplus -L -S "sys/oracle@${DOCKER_IP}:${PORT} as sysdba" <<SQL
CREATE USER scott IDENTIFIED BY tiger DEFAULT TABLESPACE users TEMPORARY TABLESPACE temp;
GRANT connect, resource, create view, create synonym TO scott;
GRANT execute ON SYS.DBMS_LOCK TO scott;
/
SQL


echo "creating oci8.pc"
mkdir -p /usr/local/pkg_config
cd /usr/local/pkg_config
export PKG_CONFIG_PATH=/usr/local/pkg_config
cat > oci8.pc <<PKGCONFIG
Name: oci8
Description: Oracle Call Interface
Version: 11.1
Cflags: -I${ORACLE_HOME}/rdbms/public
Libs: -L${ORACLE_HOME}/lib -Wl,-rpath,${ORACLE_HOME}/lib -lclntsh
PKGCONFIG

export PATH_SAVE=${PATH}



echo "set go env"
export PATH=/usr/local/go1.11.x/bin:${PATH_SAVE}
export GOROOT=/usr/local/go1.11.x
export GOPATH=/usr/local/goFiles1.11.x

echo "prepare required packages"
go get github.com/mattn/go-oci8
go get golang.org/x/tools/cmd/cover

echo "testing sql-adapter Go 1.11.x"
cd ${TESTDIR}/
go test -v . -covermode=count -coverprofile=coverage.out -args -dataSourceName="scott/tiger@${DOCKER_IP}:${PORT}/xe"
