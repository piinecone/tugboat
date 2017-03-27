#!/bin/bash

CONTEXT=$1
SPECS_DIR=$2
APP_NAME=$3

kubectl config use-context $CONTEXT

# exit immediately if any commands fail
set -e

# temp
# kubectl delete -f $SPECS_DIR/kube-lego-deployment.yaml
# kubectl delete -f $SPECS_DIR/$APP_NAME-ingress.yaml
# kubectl delete -f $SPECS_DIR/tls-configmap.yaml

# kube lego
kubectl apply -f $SPECS_DIR/tls-configmap.yaml
kubectl apply -f $SPECS_DIR/kube-lego-deployment.yaml

# start app ingress and service
kubectl apply -f $SPECS_DIR/$APP_NAME-tls-service.yaml
kubectl apply -f $SPECS_DIR/$APP_NAME-ingress.yaml

echo "-----------------------------------------------------------------------"
echo "> Run 'kubectl get ingress --watch' and update DNS records accordingly"
echo "-----------------------------------------------------------------------"
