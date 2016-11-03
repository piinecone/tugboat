#!/bin/bash

CONTEXT=$1
CONTAINER_IMAGE_NAME=$2
SITE_URL=$3
PROJECT_PREFIX=$4
GCLOUD_PROJECT_ID=$5
REGISTRY=$6
RELAY_APP_PATH=$7

./scripts/start_docker.sh
kubectl config use-context $CONTEXT

GO_APP_PATH=$GOPATH/src/github.com/$PROJECT_PREFIX/go-api/

# 1 - generate schema
cd $GO_APP_PATH
echo "----> Running update schema script..."
go run bin/update_schema.go -out $RELAY_APP_PATH/data/

# gather tags
TAG_PREFIX=piinecone/$PROJECT_PREFIX/$CONTAINER_IMAGE_NAME
IMAGE=$TAG_PREFIX:latest
GCR_IMAGE=$REGISTRY/$GCLOUD_PROJECT_ID/$CONTAINER_IMAGE_NAME
# NEXT_VERSION=$(docker images --format '{{.Repository}} {{.Tag}}' | ag $CONTAINER_IMAGE_NAME | ag 0 | awk {'print $2'} | sed 's/ //' | sort -r | awk {'print $1 += .01; exit'})

# 2 - build go api service container
cd $GO_APP_PATH
echo "----> Compiling go api binary..."
./bin/build_linux.sh

# docker build --build-arg SITE_URL=$SITE_URL -t $IMAGE .
# TODO squash this image, it's huge
# echo "----> Building $CONTAINER_IMAGE_NAME:$NEXT_VERSION for production..."
# docker build --build-arg SITE_URL=$SITE_URL --build-arg APPLICATION_ENV="production" -t $IMAGE -t $TAG_PREFIX:$NEXT_VERSION .
# docker tag -f $IMAGE $GCR_IMAGE:$NEXT_VERSION

echo "----> Building $CONTAINER_IMAGE_NAME:$NEXT_VERSION for production..."
docker build --build-arg SITE_URL=$SITE_URL --build-arg APPLICATION_ENV="production" -t $IMAGE .
docker tag -f $IMAGE $GCR_IMAGE:latest

echo "----> Built:"
# docker images | ag "$CONTAINER_IMAGE_NAME" | ag $NEXT_VERSION
docker images | ag "$CONTAINER_IMAGE_NAME" | ag "latest"
