#!/bin/bash

CONTEXT=$1
CONTAINER_IMAGE_NAME=$2
SITE_URL=$3
APPLICATION_ENV=$4
PROJECT_PREFIX=$5
GCLOUD_PROJECT_ID=$6
REGISTRY=$7
GO_APP_PATH=$8
RELAY_APP_PATH=$9

./scripts/start_docker.sh
kubectl config use-context $CONTEXT

# 1 - generate schema
cd $GO_APP_PATH
echo "----> Running update schema script..."
go run bin/update_schema.go -out $RELAY_APP_PATH/data/

# gather tags
TAG_PREFIX=$PROJECT_PREFIX/$CONTAINER_IMAGE_NAME
IMAGE=$TAG_PREFIX:latest
GCR_IMAGE=$REGISTRY/$GCLOUD_PROJECT_ID/$CONTAINER_IMAGE_NAME
# NEXT_VERSION=$(docker images --format '{{.Repository}} {{.Tag}}' | ag $CONTAINER_IMAGE_NAME | ag 0 | awk {'print $2'} | sed 's/ //' | sort -r | awk {'print $1 += .01; exit'})

# 2 - build go api service container
cd $GO_APP_PATH

echo "----> Building $IMAGE for $APPLICATION_ENV"
docker build \
  --build-arg WORKER=true \
  --build-arg SITE_URL=$SITE_URL \
  --build-arg APPLICATION_ENV=$APPLICATION_ENV \
  -t $IMAGE .
docker tag $IMAGE $GCR_IMAGE:latest

echo "----> Built:"
# docker images | ag "$CONTAINER_IMAGE_NAME" | ag $NEXT_VERSION
docker images | ag "$CONTAINER_IMAGE_NAME" | ag "latest"
