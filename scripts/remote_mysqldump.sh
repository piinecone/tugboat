#!/bin/bash

POD=$1
PASSWORD=$2
DBNAME=$3

kubectl exec -i $POD -- bash -c "mysqldump -u root -p${PASSWORD} $DBNAME > $DBNAME.sql"
