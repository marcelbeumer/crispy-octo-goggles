#!/bin/sh
set -o errexit
registry=streamproc-registry
cluster=streamproc

container_id=$(docker ps -f name=$registry -l -q)
if [ "$container_id" ]; then 
  echo "killing existing container $container_id"
  docker kill $container_id  &> /dev/null
  docker rm $container_id
fi

kind delete clusters $cluster
