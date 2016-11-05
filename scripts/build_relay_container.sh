#!/bin/bash

CONTEXT=$1
CONTAINER_IMAGE_NAME=$2
SITE_URL=$3
APP_ENV=$4
PROJECT_PREFIX=$5
GCLOUD_PROJECT_ID=$6
REGISTRY=$7
RELAY_APP_PATH=$8

./scripts/start_docker.sh
kubectl config use-context $CONTEXT

printf 'Compling app for production...'
cd $RELAY_APP_PATH
rm -rf dist
npm run compile 2>&1

IMAGE=piinecone/$PROJECT_PREFIX/$CONTAINER_IMAGE_NAME:latest
GCR_IMAGE=$REGISTRY/$GCLOUD_PROJECT_ID/$CONTAINER_IMAGE_NAME:latest

docker build -t $IMAGE .
docker tag -f $IMAGE $GCR_IMAGE
