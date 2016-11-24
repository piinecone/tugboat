#!/bin/bash

CONTEXT=$1
CONTAINER_IMAGE_NAME=$2
SITE_URL=$3
APPLICATION_ENV=$4
PROJECT_PREFIX=$5
GCLOUD_PROJECT_ID=$6
REGISTRY=$7
APP_PATH=$8

./scripts/start_docker.sh
kubectl config use-context $CONTEXT

# container tags
TAG_PREFIX=$PROJECT_PREFIX/$CONTAINER_IMAGE_NAME
IMAGE=$TAG_PREFIX:latest
GCR_IMAGE=$REGISTRY/$GCLOUD_PROJECT_ID/$CONTAINER_IMAGE_NAME

# build and tag image
cd $APP_PATH
echo "----> Building $IMAGE for $APPLICATION_ENV"
docker build --build-arg SITE_URL=$SITE_URL --build-arg APPLICATION_ENV=$APPLICATION_ENV -t $IMAGE .
docker tag -f $IMAGE $GCR_IMAGE:latest

echo "----> Built:"
docker images | ag "$CONTAINER_IMAGE_NAME" | ag "latest"
