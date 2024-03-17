#!/usr/bin/env bash
set -e
set -x

# enter demo dir relative to script
cd "$(dirname "${BASH_SOURCE[0]}")"
# go up one level
cd ..

authz_config="/files/authorization_config.yaml"

kind create cluster --config demo/kind-config-structured.yaml
docker exec kind-control-plane \
    sed -i "s,authorization-mode=.*,authorization-config=${authz_config}," \
    /etc/kubernetes/manifests/kube-apiserver.yaml
docker exec kind-control-plane \
    systemctl restart kubelet
sleep 5
kubectl -n kube-system describe pod \
    kube-apiserver-kind-control-plane | grep authorization
