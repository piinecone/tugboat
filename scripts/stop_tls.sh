#!/bin/bash

CONTEXT=$1
SPECS_DIR=$2
APP_NAME=$3

kubectl config use-context $CONTEXT

# kube lego
kubectl delete -f $SPECS_DIR/kube-lego-deployment.yaml
kubectl delete -f $SPECS_DIR/tls-configmap.yaml

# cluster
kubectl delete -f $SPECS_DIR/$APP_NAME-tls-service.yaml
kubectl delete -f $SPECS_DIR/$APP_NAME-ingress.yaml
