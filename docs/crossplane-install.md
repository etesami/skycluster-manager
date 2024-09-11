# Crossplane Setup 

## Required Providers
The skycluster-manager spans application across `AWS`, `GCP`, `Azure`
and any `Openstack` infrastructure setup. Install these providers by:

#### AWS Provider
```bash
cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-aws-ec2
spec:
  package: xpkg.upbound.io/upbound/provider-aws-ec2:v1.11.0
EOF
```

#### GCP Provider
```bash
cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-gcp-compute
spec:
  package: xpkg.upbound.io/upbound/provider-gcp-compute:v1.7.0
EOF
```

#### Azure Provider
```bash
cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-azure-compute
spec:
  package: xpkg.upbound.io/upbound/provider-azure-compute:v1.4.0
---
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-azure-network
spec:
  package: xpkg.upbound.io/upbound/provider-azure-network:v1.4.0
EOF
```

#### Openstack Provider
```bash
cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-openstack
spec:
  package: xpkg.upbound.io/crossplane-contrib/provider-openstack:v0.3.0-56.gee8b1d3
  # The latest version breaks due to the enforcement of some new fields
EOF
```

#### Kubernetes Provider
```bash
cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-kubernetes
spec:
  package: xpkg.upbound.io/crossplane-contrib/provider-kubernetes:v0.14.1
  runtimeConfigRef:
    apiVersion: pkg.crossplane.io/v1beta1
    kind: DeploymentRuntimeConfig
    name: provider-kubernetes
EOF
```

#### SSH Provider
```bash
cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-ssh
spec:
  package: docker.io/etesami/provider-ssh:latest
EOF
```