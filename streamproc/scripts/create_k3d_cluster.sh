#!/bin/sh
registry=streamproc-registry
cluster=streamproc
nodes=1

k3d cluster create $cluster \
  -a $nodes \
  --api-port "127.0.0.1:6650" \
  -p "127.0.0.1:80:80@loadbalancer" \
  --registry-create="$registry:127.0.0.1:5000"
