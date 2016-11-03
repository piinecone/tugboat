#!/bin/bash

QUERY=$1

POD=$(kubectl get pods | ag $QUERY | ag 'Running' | awk '{print $1}')
echo $POD
