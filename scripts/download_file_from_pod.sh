#!/bin/bash

POD=$1
REMOTE_FILENAME=$2
LOCAL_FILENAME=$3

kubectl exec -i $POD -- bash -c "cat ${REMOTE_FILENAME}" > $LOCAL_FILENAME
