#!/bin/bash

CONTEXT=$1
CONTAINER_IMAGE_NAME=$2
SITE_URL=$3
APPLICATION_ENV=$4
PROJECT_PREFIX=$5
GCLOUD_PROJECT_ID=$6
REGISTRY=$7
GO_APP_PATH=$8

kubectl config use-context $CONTEXT

cd $GO_APP_PATH

TAG_PREFIX=$PROJECT_PREFIX/$CONTAINER_IMAGE_NAME
IMAGE=$TAG_PREFIX:latest
GCR_IMAGE=$REGISTRY/$GCLOUD_PROJECT_ID/$CONTAINER_IMAGE_NAME

cd $GO_APP_PATH
echo "----> Building $IMAGE for $APPLICATION_ENV"
docker build \
  --no-cache \
  --build-arg WORKER=false \
  --build-arg SITE_URL=$SITE_URL \
  --build-arg APPLICATION_ENV=$APPLICATION_ENV \
  -t $IMAGE .
docker tag $IMAGE $GCR_IMAGE:latest

echo "----> Built:"
docker images | ag "$CONTAINER_IMAGE_NAME" | ag "latest"
