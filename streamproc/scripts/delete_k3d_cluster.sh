#!/bin/sh
registry=streamproc-registry
cluster=streamproc

k3d registry delete $registry &> /dev/null
k3d cluster delete $cluster &> /dev/null
