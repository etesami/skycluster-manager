#!/bin/bash

# If env variables are not set, exit
if [ -z "$AWS_ACCESS_KEY_ID" ] || [ -z "$AWS_SECRET_ACCESS_KEY" ]; then
  echo "AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY must be set."
  exit 1
fi

# Create the content of the credentials in a variable
creds_content="[default]
aws_access_key_id = $AWS_ACCESS_KEY_ID
aws_secret_access_key = $AWS_SECRET_ACCESS_KEY"

# Echo the content and pipe it to base64 for encoding
creds_enc=$(echo "$creds_content" | base64 -w0)

cat <<EOF | kubectl apply -f -
apiVersion: aws.upbound.io/v1beta1
kind: ProviderConfig
metadata:
  name: provider-cfg-aws
spec:
  credentials:
    source: Secret
    secretRef:
      name: secret-aws
      namespace: crossplane-system
      key: creds
---
apiVersion: v1
kind: Secret
metadata:
  name: secret-aws
  namespace: crossplane-system
type: Opaque
data:
  creds: $creds_enc
EOF