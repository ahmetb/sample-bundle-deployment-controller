# sample-bundle-deployment-controller

A sample CRD that deploys a bundle of arbitrary Kubernetes resources with
pruning/kustomization capabilities.

Copyright 2022 Ahmet Alp Balkan.

## Setup

1. Make sure you have [Kind](https://kind.sigs.k8s.io/),
   Docker and [ko](https://ko.build/deployment/) installed.

1. Create Kind cluster:

       kind create cluster

1. Apply CRD manifests:

       kubectl apply -R -f config/

1. Build and deploy the `controller`:

       go run ./controller 

1. Apply the sample manifest:

       kubectl apply -f example-bundle-deployment.yaml
