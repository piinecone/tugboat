#!/bin/bash

CONTEXT=$1
CLUSTER_NAME=$2
DEPLOYMENT_NAME=$3
APP_NAME=$4
CONTAINER_NAME=$5
REGISTRY=$6

START=$(date "+%s")
echo "container name: $CONTAINER_NAME"

# get container version
# 
# LATEST_TAG=$(docker images --format '{{.Repository}} {{.Tag}}' | ag $CONTAINER_NAME | ag 0 | awk {'print $2'} | sed 's/ //' | sort -r | awk {'print $1; exit'})
# echo "DEPLOYING VERSION: $LATEST_TAG"
# 
# kubectl config use-context $CONTEXT
# 
# EXISTING=$(kubectl get deployments | ag "$APP_NAME" | awk {'print $1'})
# 
# if [ "$EXISTING" == "" ]
# then
#   echo "Creating new deployment..."
#   kubectl create -f ./clusters/$CLUSTER_NAME/specs/$DEPLOYMENT_NAME.yaml
# else
#   echo "Updating existing deployment..."
#   kubectl set image deployment/$APP_NAME $APP_NAME=$APP_NAME:$CONTAINER_VERSION
#   kubectl replace -f ./clusters/$CLUSTER_NAME/specs/$DEPLOYMENT_NAME.yaml
# fi

kubectl config use-context $CONTEXT
kubectl delete -f ./clusters/$CLUSTER_NAME/specs/$DEPLOYMENT_NAME.yaml
kubectl create -f ./clusters/$CLUSTER_NAME/specs/$DEPLOYMENT_NAME.yaml

END=$(date "+%s")
ELAPSED=$(($END - $START))
BLUE='\033[1;34m'
NC='\033[0m' # No Color
printf "${BLUE}[Deploy Complete]${NC} deploy finished in $ELAPSED s\n"
