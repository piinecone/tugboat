#!/bin/bash

CONTAINER_NAME=$1
PROJECT_PREFIX=$2

./scripts/start_docker.sh

echo "Removing $PROJECT_PREFIX/$CONTAINER_NAME images..."
docker rmi -f $(docker images | ag "$PROJECT_PREFIX/$CONTAINER_NAME" | awk {'print $3'})
echo "Done."
