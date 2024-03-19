## non structured authz example

### build docker image
docker build -t kubernetes-sigs/referencegrant-poc/refauthz:latest -f Dockerfile .

### run the authorizer as a docker container
docker run --name authorizer --rm --network kind  -p 8081:8081 kubernetes-sigs/referencegrant-poc/refauthz:latest

### create kind cluster (authorizer needed to run before)
kind create cluster --retain --name=demo  --config=demo/kind-config.yaml -v 2

## apply crds
kubectl apply -f config/crd/
kubectl kustomize "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v1.0.0" | kubectl apply -f -

## Export kubeconfig file
kind get kubeconfig --internal --name=demo > ~/demo/kubeconfig

## Copy the file to the container
docker cp ~/demo/kubeconfig   authorizer:/demo/kubeconfig

## create demo namespace
kubectl create ns demo

## apply base manifests (SA and secret)
kubectl apply -f demo/manifests/base.yaml

## show the picture of the cluster
kubectl get gateway --> should show nothing

k get crg --> should show nothing

k get crc --> should show nothing

k get rg --> should show nothing

kubectl auth can-i get  secrets/demo-tls-secret --as system:serviceaccount:demo:demo-controller --> should return "NO"

## Deploy ClusterRefGrant and ClusterRefConsumer
cat demo/manifests/crg.yaml | yq

cat demo/manifests/crc.yaml | yq

k apply -f demo/manifests/crg.yaml

k apply -f demo/manifests/crc.yaml

--Look at the reconciliation logs to see the graph-- (should show nothing b/c nothing is referenced)

kubectl auth can-i get  secrets/demo-tls-secret --as system:serviceaccount:demo:demo-controller --> should return "NO"

## Deploy Gateway 
cat demo/manifests/gateway.yaml | yq

k apply -f demo/manifests/gateway.yaml

kubectl auth can-i get  secrets/demo-tls-secret --as system:serviceaccount:demo:demo-controller --> should return yes

## Deploy gateway-cross-ns
cat demo/manifests/gateway-secret-cross-ns.yaml | yq

k apply -f demo/manifests/gateway-secret-cross-ns.yaml

kubectl auth can-i get --namespace=default secrets/demo-tls-secret-default --as system:serviceaccount:demo:demo-controller --> should return NO

## Deploy ReferenceGrant
cat demo/manifests/referencegrant.yaml | yq

k apply -f demo/manifests/referencegrant.yaml

kubectl auth can-i get --namespace=default secrets/demo-tls-secret-default --as system:serviceaccount:demo:demo-controller --> should return yes

## removing the gateway
k delete -f demo/manifests/gateway-secret-cross-ns.yaml

kubectl auth can-i get --namespace=default secrets/demo-tls-secret-default --as system:serviceaccount:demo:demo-controller --> should return No