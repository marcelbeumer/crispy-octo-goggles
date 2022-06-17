#!/bin/sh
set -o errexit
registry=streamproc-registry
registry_port=5000
cluster=streamproc

function get_cluster_yaml() {
  cat <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
    - |
      kind: InitConfiguration
      nodeRegistration:
        kubeletExtraArgs:
          node-labels: "ingress-ready=true"
  extraPortMappings:
    - containerPort: 80
      hostPort: 80
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."${registry}:${registry_port}"]
    endpoint = ["http://${registry}:5000"]
EOF
}

# create registry container unless it already exists
if [ "$(docker inspect -f '{{.State.Running}}' "${registry}" 2>/dev/null || true)" != 'true' ]; then
  docker run \
    -d --restart=always \
    -p "127.0.0.1:${registry_port}:${registry_port}" \
    --name "${registry}" \
    registry:2
fi

# create a cluster with the local registry enabled in containerd
echo "$(get_cluster_yaml)" | kind create cluster --name $cluster --config=-

# connect the registry to the cluster network if not already connected
if [ "$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' "${registry}")" = 'null' ]; then
  docker network connect "kind" "${registry}"
fi

# Document the local registry
# https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "${registry}:${registry_port}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

echo "Installing ingress"
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

# echo "Waiting for ingress to be ready"
# kubectl wait --namespace ingress-nginx \
#   --for=condition=ready pod \
#   --selector=app.kubernetes.io/component=controller \
#   --timeout=90s
