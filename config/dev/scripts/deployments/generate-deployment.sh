#!/bin/bash

# Check if appName is provided
if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ]; then
  echo "Usage: $0 <action> <appName> <providerNum>"
  exit 1
fi

ACTION=$1
APP_NAME=$2
PROVIDER_NUM=$3
SKYAPP_NAME="$APP_NAME-$PROVIDER_NUM-skyapp"
DATAFLOW_NAME="$APP_NAME-$PROVIDER_NUM-dataflow"
OUTPUT_DIR="p$PROVIDER_NUM"

# Create output directory if it does not exist
if [ ! -d "$OUTPUT_DIR" ]; then
  mkdir -p "$OUTPUT_DIR"
fi

# Create skyapp yaml
cat <<EOF > "${OUTPUT_DIR}/${SKYAPP_NAME}.yaml"
apiVersion: core.skycluster-manager.savitestbed.ca/v1alpha1
kind: SkyApp
metadata:
  name: ${SKYAPP_NAME}
spec:
  appName: ${SKYAPP_NAME}
  namespace: default
  appConfig:
    - name: frontend
      constraints:
        locationConstraints:
          - providerType: cloud
        virtualServiceConstraints:
          - virtualServiceName: vs1
    - name: backend
      constraints:
        locationConstraints:
          - providerType: cloud
        virtualServiceConstraints:
          - virtualServiceName: vs2
    - name: payment
      constraints:
        locationConstraints:
          - providerType: cloud
        virtualServiceConstraints:
          - virtualServiceName: vs3
EOF

# Create dataflow yaml
cat <<EOF > "${OUTPUT_DIR}/${DATAFLOW_NAME}.yaml"
apiVersion: core.skycluster-manager.savitestbed.ca/v1alpha1
kind: DataflowAttribute
metadata:
  name: ${DATAFLOW_NAME}
spec:
  appName: ${SKYAPP_NAME}
  connections:
    - source: frontend
      destinations:
        - name: backend
          constraints:
            latency: 500ms
    - source: backend
      destinations:
        - name: payment
          constraints:
            latency: 500ms
EOF

kubectl ${ACTION} -f "${OUTPUT_DIR}/${SKYAPP_NAME}.yaml" -f "${OUTPUT_DIR}/${DATAFLOW_NAME}.yaml"
