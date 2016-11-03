#!/bin/bash

DB=$1
DUMPFILE=$2

mysql -u root $DB < $DUMPFILE
