# Skycluster Manager

<center style="margin: 0.95rem 0;">
<img src="assets/images/skycluster-logo-1-min.png" alt="SkyCluster Logo" width="300"/>
</center>

`skycluster-manager` is a custom Kubernetes controller designed to 
facilitate the deployment of Kubernetes resources (deployments, 
services, and config maps) in a multi-cloud or hybrid-cloud Kubernetes 
environment, specifically tailored for a given application.

The `skycluster-manager` operates within a management Kubernetes cluster. Users interact with this management cluster by submitting their application manifests, which include deployments, services, and config maps. The `skycluster-manager` then provisions a new multi-cloud or hybrid-cloud Kubernetes cluster and deploys the submitted application manifests into it.

The `skycluster-manager` relies on 
[Crossplane](https://github.com/crossplane/crossplane) 
for managing external cloud resources.

## Getting Started

### Pre-requisits

1. **Management Kubernetes Cluster**: A management Kubernetes cluster is
required to run the `skycluster-manager` and act as the point of 
contact for submitting your application. You can create a local 
management Kubernetes cluster using `kind` with the following command:

```bash
kind create cluster
```

2. **Crossplane**: Crossplane is a Kubernetes extension that allows 
Kubernetes to manage external cloud resources via standard Kubernetes 
APIs. To install Crossplane, use the following commands:

```bash
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update

# Tested with version 1.16.0
helm install crossplane \
  --namespace crossplane-system \
  --create-namespace crossplane-stable/crossplane \
  --version 1.16.0
```

Once Crossplane is installed, follow the instructions 
[here](./docs/crossplane-install.md) to install all the required 
providers, or simply run the following command:

```bash
sudo kubectl apply -f ./config/installation/crossplane-setup.yaml
```

3. **Providers Authentication Configuration**:
Cloud or hybrid providers require authentication to be used by the 
`skycluster-manager`. You will need to create a user with sufficient 
permissions and provide the necessary access credentials. Follow the 
instructions [here](./docs/crossplane-config.md) 
to configure authentication for `AWS`, `GCP`, and `Azure`.