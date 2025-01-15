#!/bin/bash

if [[ ! -f $AZURE_CONFIG_PATH ]]; then
  echo "Azure config file not found at $AZURE_CONFIG_PATH"
  exit 1
fi

cont_enc=$(echo $AZURE_CONFIG_PATH | base64 -w0)

cat <<EOF | kubectl apply -f -
apiVersion: azure.upbound.io/v1beta1
metadata:
  name: provider-cfg-azure
kind: ProviderConfig
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: secret-azure
      key: creds
---
apiVersion: v1
kind: Secret
metadata:
  name: secret-azure
  namespace: crossplane-system
type: Opaque
data:
  creds: $cont_enc
EOF