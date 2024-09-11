# Crossplane Configuration

Providers installed in Crossplane require authentication to manage 
external resources. For each cloud provider that you want to integrate 
into the system, a separate configuration must be created.

### AWS Configuration
In the AWS Console, navigate to Identity and Access Management (IAM) 
and create a new user. Ensure the user has the following 
permission policy: `AmazonEC2FullAceess`. 
Next, in the Security Credentials section, generate an access key. 
After obtaining the Access Key ID and Secret Access Key, use the script 
below to create configuration for AWS:

```bash
AWS_ACCESS_KEY_ID=abcd....xwyz # replace with your ID
AWS_SECRET_ACCESS_KEY=abcd....xwyz # replace with your Key

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
```

### GCP Configuration
Create a new project in Google Cloud, then add a service account. 
Generate a service account key file in JSON format. After that, 
use the script below:

```bash
kubectl create secret generic secret-gcp -n crossplane-system --from-file=creds=./sv-acc.json

# Apply the provider configuration
cat <<EOF | kubectl apply -f -
apiVersion: gcp.upbound.io/v1beta1
kind: ProviderConfig
metadata:
  name: provider-cfg-gcp
spec:
  projectID: learned-cosine-391615
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: secret-gcp
      key: creds
EOF
```

### Azure Configuration
Create a subscription and note your Subscription ID. Next, create a 
service principal and configure its access to Azure resources. 
This can be done using the `az` CLI tool. Follow the script below:

```bash
export SUBS=<subsc-id> # replace with your subscription id
az account set --subscription $SUBS
cont_json=$(az ad sp create-for-rbac --sdk-auth --role Owner \
  --scopes /subscriptions/$SUBS)
cont_enc=$(echo $cont_json | base64 -w0)

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
```