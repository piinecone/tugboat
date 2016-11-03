#!/bin/bash

# check that vm exists
DOCKER_VM_STATE=$(docker-machine ls | ag docker-vm | awk '{print $4}')

if [ "$DOCKER_VM_STATE" != "Running" ]
then
  # { echo "docker-vm is not running" 1>&2 ; exit 1; }
  # docker-machine -D ssh docker-vm sudo /etc/init.d/docker restart
  # docker-machine start docker-vm
  # docker-machine restart docker-vm
  printf "Starting docker-vm because it is not running..."
  docker-machine start docker-vm
  eval "$(docker-machine env docker-vm)"
else
  printf "Configuring local docker-machine..."
  eval "$(docker-machine env docker-vm)"
  # printenv | ag DOCKER
fi
