#!/bin/sh
set -o errexit
registry=streamproc-registry
cluster=streamproc
host=localhost

container_id=$(docker ps -f name=$registry -l -q)
if [ -z "$container_id" ]; then echo "Could not find registry container"; exit 1; fi

port=$(docker inspect $container_id -f "{{ range .HostConfig.PortBindings }}{{range .}}{{.HostPort}}{{end}}{{end}}" )
if [ -z "$port" ]; then echo "Could not find host port for container $container_id"; exit 1; fi

echo ">>> pushing images to $host:$port"

services=( "aggregator" "consumer-high" "consumer-low" "event-api" "event-producer")
for name in "${services[@]}"
do
	echo ">>> pushing image for $name"
  docker tag streamproc/k3d/$name:latest $host:$port/$name:latest
  docker push $host:$port/$name:latest
done
