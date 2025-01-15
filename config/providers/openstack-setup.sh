#!/bin/bash

# Check if any of these variables are not set, if so exist
if [[ -z $AUTH_URL || -z $USERNAME || -z $PASSWORD || -z $TENANT_NAME || \
  -z $REGION || -z $USER_DOMAIN_NAME || -z $PROJECT_DOMAIN_NAME ]]; then
  echo "One or more required variables are not set."
  exit 1
fi

cat <<EOF | kubectl apply -f -
apiVersion: openstack.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: provider-cfg-os-${REGION}
spec:
  credentials:
    source: Secret
    secretRef:
      name: secret-os-${REGION}
      namespace: crossplane-system
      key: configs
---
apiVersion: v1
kind: Secret
metadata:
  name: secret-os-${REGION}
  namespace: crossplane-system
type: Opaque
stringData:
  configs: |
    {
      "auth_url": $AUTH_URL,
      "user_name": $USERNAME,
      "password": $PASSWORD,
      "tenant_name": $TENANT_NAME,
      "region": $REGION,
      "user_domain_name": $USER_DOMAIN_NAME,
      "project_domain_name": $PROJECT_DOMAIN_NAME
    }
EOF