#!/bin/bash

# Check if appName is provided
if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ]; then
  echo "Usage: $0 <action> <appName> <compNum/providerNum>"
  exit 1
fi

ACTION=$1
APP_NAME=$2

# Defult value for providers num
PROVIDER_NUM=20
APP_NUM=$3

SKYAPP_NAME="$APP_NAME-c$APP_NUM-p$PROVIDER_NUM-skyapp"
DATAFLOW_NAME="$APP_NAME-$PROVIDER_NUM-dataflow"
OUTPUT_DIR="c$APP_NUM-p$PROVIDER_NUM"
APP_YAML_DIR="../skycluster-notebook/apps/C$APP_NUM"

# Create output directory if it does not exist
if [ ! -d "$OUTPUT_DIR" ]; then
  mkdir -p "$OUTPUT_DIR"
fi

APP_COMPONENTS=$(cat "$APP_YAML_DIR/generated_skyapp.yaml")
UPDATED_APP_COMPONENTS=$(echo "$APP_COMPONENTS" | sed "s/generated-app/${SKYAPP_NAME}/g")
echo "$UPDATED_APP_COMPONENTS" > "${OUTPUT_DIR}/${SKYAPP_NAME}.yaml"

APP_DATAFLOW=$(cat "$APP_YAML_DIR/generated_dataflow.yaml")
UPDATED_APP_DATAFLOW=$(echo "$APP_DATAFLOW" | sed "s/generated-app/${SKYAPP_NAME}/g")

echo "$UPDATED_APP_DATAFLOW" > "${OUTPUT_DIR}/${DATAFLOW_NAME}.yaml"

kubectl ${ACTION} -f "${OUTPUT_DIR}/${SKYAPP_NAME}.yaml" -f "${OUTPUT_DIR}/${DATAFLOW_NAME}.yaml"
