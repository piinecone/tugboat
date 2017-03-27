#!/bin/bash

CONTAINER_NAME=$1
GCLOUD_PROJECT_ID=$2
REGISTRY=$3

# LATEST_TAG=$(docker images --format '{{.Repository}} {{.Tag}}' | ag $CONTAINER_NAME | ag 0 | awk {'print $2'} | sed 's/ //' | sort -r | awk {'print $1; exit'})
LATEST_TAG="latest"

# echo "$(gcloud auth print-access-token)"
docker login -e hughes.nick@gmail.com -u _token -p "$(gcloud auth print-access-token)" https://$REGISTRY
# docker rmi $(docker images | grep none | awk {'print $3'})

echo "Pushing image $REGISTRY/$GCLOUD_PROJECT_ID/$CONTAINER_NAME:$LATEST_TAG..."

n=0
until [ $n -ge 10 ]
do
  docker push $REGISTRY/$GCLOUD_PROJECT_ID/$CONTAINER_NAME:$LATEST_TAG && break
  n=$[$n+1]
  sleep 1
  echo "Failed, retrying..."
done
