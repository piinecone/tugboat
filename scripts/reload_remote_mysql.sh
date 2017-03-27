#!/bin/bash

POD=$1
PASSWORD=$2
DUMPFILE=$3
DBNAME=$4

kubectl exec -i $POD -- bash -c "cat > ${DBNAME}.sql" < $DUMPFILE
kubectl exec -i $POD -- bash -c "mysql -u root -p${PASSWORD} -e 'DROP DATABASE ${DBNAME}; CREATE DATABASE ${DBNAME};'"
kubectl exec -i $POD -- bash -c "mysql -u root -p${PASSWORD} ${DBNAME} < ${DBNAME}.sql"
