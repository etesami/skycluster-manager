# Skycluster Manager

<img style="margin: 0.95rem 0;" src="assets/images/skycluster-logo-1-min.png" alt="SkyCluster Logo" width="300"/>

`skycluster-manager` is a custom Kubernetes controller designed to 
facilitate the deployment of Kubernetes resources (deployments, 
services, and config maps) in a multi-cloud or hybrid-cloud Kubernetes 
environment, specifically tailored for a given application.

The `skycluster-manager` operates within a management Kubernetes cluster. Users interact with this management cluster by submitting their application manifests, which include deployments, services, and config maps. The `skycluster-manager` then provisions a new multi-cloud or hybrid-cloud Kubernetes cluster and deploys the submitted application manifests into it.

The `skycluster-manager` relies on 
[Crossplane](https://github.com/crossplane/crossplane) 
for managing external cloud resources.

Please refer to [docs](https://skycluster.io) for installation and configurations.
