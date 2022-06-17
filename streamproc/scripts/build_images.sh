#!/bin/sh
set -o errexit

services=( "aggregator" "consumer-high" "consumer-low" "event-api" "event-producer" "web-ui")
for name in "${services[@]}"
do
	echo ">>> building docker image for $name"
  docker build -t streamproc/k3d/$name:latest -f services/$name/Dockerfile ./services/$name
done
