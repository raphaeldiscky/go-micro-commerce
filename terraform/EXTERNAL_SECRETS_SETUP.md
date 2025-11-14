# External Secrets Operator Setup Guide

Complete guide for managing secrets with External Secrets Operator (ESO) and Google Secret Manager for the Go Micro-Commerce platform.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [What is External Secrets Operator?](#what-is-external-secrets-operator)
- [Setup Instructions](#setup-instructions)
- [Secret Inventory](#secret-inventory)
- [Populating Secrets](#populating-secrets)
- [Verification](#verification)
- [Updating Secrets](#updating-secrets)
- [Secret Rotation](#secret-rotation)
- [Troubleshooting](#troubleshooting)
- [Security Best Practices](#security-best-practices)
- [Cost Considerations](#cost-considerations)

## Overview

This setup uses **External Secrets Operator (ESO) v1.0.0** to synchronize secrets from **Google Secret Manager** to Kubernetes Secrets automatically. This approach provides:

- ✅ **Centralized secret management**: Single source of truth in Google Secret Manager
- ✅ **Automatic synchronization**: Secrets are synced to Kubernetes every hour
- ✅ **Workload Identity**: Secure authentication without service account keys
- ✅ **Audit logging**: All secret access is logged in Google Cloud
- ✅ **Encryption at rest**: Secrets encrypted by Google's infrastructure
- ✅ **GitOps friendly**: Secret definitions in Git, values in Secret Manager

## Architecture

Here's the diagram converted into clear, structured text:

**Google Secret Manager**  
Stores sensitive secrets centrally in Google Cloud Platform:

- JWT Private Key
- Stripe Secret
- SMTP Password

These secrets are accessed securely via **Workload Identity**, using a dedicated GCP Service Account (`external-secrets-sa`) that grants the External Secrets Operator the necessary permissions.

**External Secrets Operator (ESO)**  
An open-source Kubernetes operator that synchronizes secrets from external systems (like GCP Secret Manager) into Kubernetes Secrets.

Configuration:

- **ClusterSecretStore**:
  - Provider: Google Secret Manager
  - Authentication: Workload Identity
  - Service Account: `external-secrets-sa`

Manages **9 ExternalSecret resources**, each mapping a specific GCP secret to a corresponding Kubernetes Secret:

- `auth-service-jwt-keys`
- `payment-service-secrets`
- `notification-service-secrets`
- ... (6 additional secrets for other services)

Syncs secrets from GCP to Kubernetes **every hour** (configurable).

**Kubernetes Secrets**  
Automatically created and updated by the External Secrets Operator. Each corresponds to a GCP secret:

- `auth-service-jwt-keys`
- `payment-service-secrets`
- `notification-service-secrets`
- ... (total of 9 secrets)

These secrets are mounted as volumes or exposed as environment variables to application pods.

**Application Pods**  
Kubernetes workloads that consume the secrets:

- **Auth Service** — Uses `auth-service-jwt-keys` to sign/verify JWT tokens
- **Payment Service** — Uses `payment-service-secrets` (e.g., Stripe API key)
- **Notification Service** — Uses `notification-service-secrets` (e.g., SMTP password)
- ... (other services using the remaining secrets)

Secrets are injected securely at runtime via volume mounts or environment variables — never hardcoded.

**Flow Summary**:  
GCP Secret Manager → (via Workload Identity) → External Secrets Operator → Kubernetes Secrets → Application Pods

This architecture ensures secrets are never stored in code or version control, are centrally managed, and dynamically synced with minimal operational overhead.

## Prerequisites

Before setting up External Secrets Operator, ensure you have:

1. **GKE Cluster with Workload Identity enabled**
   - Already enabled in your cluster via `enable_workload_identity = true`

2. **gcloud CLI installed and authenticated**

   ```bash
   gcloud auth login
   gcloud config set project YOUR_PROJECT_ID
   ```

3. **kubectl configured for your GKE cluster**

   ```bash
   gcloud container clusters get-credentials go-micro-commerce-prod --zone=asia-southeast1-a
   ```

4. **Terraform >= 1.13.0 installed**

5. **Secret Manager API enabled** (automatically enabled by the setup script)

## What is External Secrets Operator?

**External Secrets Operator (ESO)** is a Kubernetes operator that integrates external secret management systems (like Google Secret Manager, AWS Secrets Manager, HashiCorp Vault) with Kubernetes.

### Key Concepts

**1. ClusterSecretStore**

- Cluster-wide resource that defines how to connect to Google Secret Manager
- Uses Workload Identity for authentication
- Configured once, used by all ExternalSecrets

**2. ExternalSecret**

- Namespace-scoped resource that defines which secrets to sync
- References the ClusterSecretStore
- Creates/updates a Kubernetes Secret automatically
- Syncs every hour (configurable via `refreshInterval`)

**3. Workload Identity**

- Allows Kubernetes service accounts to impersonate GCP service accounts
- No need to manage service account keys
- Follows Google Cloud best practices

## Setup Instructions

### Step 1: Deploy Infrastructure with Terraform

The External Secrets Operator module is already integrated into the production environment configuration.

1. **Navigate to the production environment:**

   ```bash
   cd terraform/environments/prod
   ```

2. **Ensure terraform.tfvars is configured:**

   ```hcl
   # External Secrets Operator configuration
   eso_namespace                   = "external-secrets"
   eso_chart_version               = "1.0.0"
   eso_replicas                    = 2
   eso_create_cluster_secret_store = true
   eso_cluster_secret_store_name   = "gcp-secret-manager"
   ```

3. **Apply Terraform configuration:**

   ```bash
   ./terraform/scripts/apply-prod.sh
   ```

   This will:
   - Install External Secrets Operator Helm chart v1.0.0
   - Create GCP service account: `external-secrets-operator`
   - Grant `roles/secretmanager.secretAccessor` role
   - Configure Workload Identity binding
   - Create Kubernetes service account: `external-secrets-sa`
   - Deploy ClusterSecretStore: `gcp-secret-manager`

### Step 2: Populate Google Secret Manager

Use the provided script to populate all secrets:

```bash
./terraform/scripts/populate-secrets.sh
```

The script will guide you through creating:

**Critical Secrets:**

- JWT private and public keys
- Stripe secret key and webhook secret

**High Priority:**

- SMTP credentials (username, password, host)
- SendGrid API key
- Elasticsearch credentials
- Temporal API key
- Redis cluster password

**Medium Priority:**

- Grafana admin password
- ArgoCD admin password

**Tips for secret generation:**

```bash
# Generate RSA key pair for JWT
ssh-keygen -t rsa -b 4096 -m PEM -f jwt-key
# Creates: jwt-key (private) and jwt-key.pub (public)

# Generate strong passwords
openssl rand -base64 32

# Generate shorter passwords
openssl rand -base64 20
```

### Step 3: Deploy ExternalSecret Resources

Apply the ExternalSecret manifests to your cluster:

```bash
kubectl apply -k deployments/k8s/base/external-secrets/
```

This creates 9 ExternalSecret resources that will sync secrets from Google Secret Manager.

### Step 4: Update Service Deployments

Update your service deployments to reference the ESO-managed secrets instead of hardcoded values.

**Example for auth-service:**

**Before:**

```yaml
volumes:
  - name: jwt-keys
    secret:
      secretName: auth-service-jwt-keys # Manually created secret
```

**After:**

```yaml
volumes:
  - name: jwt-keys
    secret:
      secretName: auth-service-jwt-keys # Created by ExternalSecret
```

The secret name remains the same, but now it's managed by ESO and automatically synced from Google Secret Manager.

### Step 5: Remove Hardcoded Secrets from Git

**IMPORTANT**: After verifying that ESO-managed secrets work, remove hardcoded secrets from Git:

```bash
# Remove local secret files (if any)
rm -rf deployments/k8s/overlays/local/secrets/*.pem
rm -rf deployments/k8s/overlays/local/secrets/*.key

# Update .gitignore to prevent future commits
echo "*.pem" >> .gitignore
echo "*.key" >> .gitignore
echo "terraform.tfvars" >> .gitignore
```

## Secret Inventory

### Critical Priority

| Secret Name                             | Description                         | Used By         | Rotation Frequency |
| --------------------------------------- | ----------------------------------- | --------------- | ------------------ |
| `auth-service-jwt-private-key`          | RSA private key for JWT signing     | auth-service    | Quarterly          |
| `jwt-public-key`                        | RSA public key for JWT verification | All services    | Quarterly          |
| `payment-service-stripe-secret-key`     | Stripe API secret key               | payment-service | Annually           |
| `payment-service-stripe-webhook-secret` | Stripe webhook signature secret     | payment-service | As needed          |

### High Priority

| Secret Name                             | Description            | Used By              | Rotation Frequency |
| --------------------------------------- | ---------------------- | -------------------- | ------------------ |
| `notification-service-smtp-host`        | SMTP server hostname   | notification-service | As needed          |
| `notification-service-smtp-username`    | SMTP username/email    | notification-service | Quarterly          |
| `notification-service-smtp-password`    | SMTP password          | notification-service | Quarterly          |
| `notification-service-sendgrid-api-key` | SendGrid API key       | notification-service | Quarterly          |
| `search-service-elasticsearch-username` | Elasticsearch username | search-service       | Quarterly          |
| `search-service-elasticsearch-password` | Elasticsearch password | search-service       | Quarterly          |
| `order-service-temporal-api-key`        | Temporal Cloud API key | order-service        | Quarterly          |
| `redis-cluster-password`                | Redis cluster password | All services         | Quarterly          |

### Medium Priority

| Secret Name              | Description            | Used By           | Rotation Frequency |
| ------------------------ | ---------------------- | ----------------- | ------------------ |
| `grafana-admin-password` | Grafana admin password | Grafana dashboard | Quarterly          |
| `argocd-admin-password`  | ArgoCD admin password  | ArgoCD UI         | Quarterly          |

## Populating Secrets

### Using the Automated Script

The easiest way to populate secrets is using the provided script:

```bash
./terraform/scripts/populate-secrets.sh
```

### Manual Secret Creation

If you prefer to create secrets manually:

```bash
# Set your project ID
export PROJECT_ID="your-gcp-project-id"

# Create a secret
gcloud secrets create SECRET_NAME \
  --replication-policy="automatic" \
  --project="$PROJECT_ID"

# Add a secret version from stdin
echo -n "secret-value" | gcloud secrets versions add SECRET_NAME \
  --data-file=- \
  --project="$PROJECT_ID"

# Add a secret version from file
gcloud secrets versions add SECRET_NAME \
  --data-file="path/to/file" \
  --project="$PROJECT_ID"
```

### Importing Existing Secrets

If you already have secrets in another system:

```bash
# Export from existing Kubernetes secret
kubectl get secret auth-service-jwt-keys -o jsonpath='{.data.private-key\.pem}' | base64 -d > jwt-key

# Import to Google Secret Manager
gcloud secrets versions add auth-service-jwt-private-key \
  --data-file=jwt-key \
  --project="$PROJECT_ID"

# Clean up local file
rm jwt-key
```

## Verification

### 1. Verify External Secrets Operator Installation

```bash
# Check ESO pods
kubectl get pods -n external-secrets
# Should show 2 replicas running

# Check ESO version
kubectl get deployment -n external-secrets external-secrets -o jsonpath='{.spec.template.spec.containers[0].image}'
# Should show: ghcr.io/external-secrets/external-secrets:v1.0.0 (or similar)
```

### 2. Verify ClusterSecretStore

```bash
# Check ClusterSecretStore status
kubectl get clustersecretstore gcp-secret-manager

# Expected output:
# NAME                  AGE   STATUS   READY
# gcp-secret-manager    1m    Valid    True
```

### 3. Verify ExternalSecrets

```bash
# List all ExternalSecrets
kubectl get externalsecrets -A

# Check status of a specific ExternalSecret
kubectl describe externalsecret auth-service-jwt-keys -n default

# Look for:
# Status:
#   Conditions:
#     Status: True
#     Type:   Ready
#   Sync Status: SecretSynced
```

### 4. Verify Synced Kubernetes Secrets

```bash
# List secrets created by ESO
kubectl get secrets -A | grep external-secrets

# Check a specific secret
kubectl get secret auth-service-jwt-keys -n default

# Verify secret data (be careful with sensitive data)
kubectl get secret auth-service-jwt-keys -n default -o jsonpath='{.data.private-key\.pem}' | base64 -d | head -c 50
# Should show: -----BEGIN RSA PRIVATE KEY-----
```

### 5. Test Application Access

```bash
# Check auth-service logs
kubectl logs -n default deployment/auth-service --tail=50

# Look for successful JWT key loading, no errors about missing secrets
```

## Updating Secrets

### Updating a Secret Value

When you need to update a secret (e.g., rotate password):

```bash
# Add a new version to the secret
echo -n "new-secret-value" | gcloud secrets versions add SECRET_NAME \
  --data-file=- \
  --project="$PROJECT_ID"

# ESO will automatically sync the new value within 1 hour
# Or force immediate sync by deleting the ExternalSecret
kubectl delete externalsecret SECRET_EXTERNAL_SECRET_NAME -n NAMESPACE
kubectl apply -k deployments/k8s/base/external-secrets/

# Restart pods to pick up the new secret
kubectl rollout restart deployment SERVICE_NAME -n NAMESPACE
```

### Immediate Secret Update

For critical updates that can't wait for the hourly sync:

```bash
# Option 1: Reduce refresh interval temporarily
kubectl patch externalsecret SECRET_NAME -n NAMESPACE --type=merge -p '{"spec":{"refreshInterval":"1m"}}'

# Wait 1-2 minutes, then restore original interval
kubectl patch externalsecret SECRET_NAME -n NAMESPACE --type=merge -p '{"spec":{"refreshInterval":"1h"}}'

# Option 2: Force sync by deleting and recreating
kubectl delete externalsecret SECRET_NAME -n NAMESPACE
kubectl apply -f deployments/k8s/base/external-secrets/SECRET_FILE.yaml
```

## Secret Rotation

### Recommended Rotation Schedule

| Secret Type          | Rotation Frequency  | Impact   | Automation     |
| -------------------- | ------------------- | -------- | -------------- |
| JWT Keys             | Quarterly (90 days) | High     | Manual         |
| Payment Gateway Keys | Annually            | Critical | Manual         |
| Database Passwords   | Quarterly           | High     | Semi-automated |
| SMTP Credentials     | Quarterly           | Medium   | Manual         |
| Admin Passwords      | Quarterly           | Medium   | Manual         |
| API Keys             | Quarterly           | Medium   | Manual         |

### JWT Key Rotation Procedure

1. **Generate new key pair:**

   ```bash
   ssh-keygen -t rsa -b 4096 -m PEM -f jwt-key-new
   ```

2. **Add new public key to all services first:**

   ```bash
   # Update jwt-public-key secret
   gcloud secrets versions add jwt-public-key \
     --data-file=jwt-key-new.pub \
     --project="$PROJECT_ID"

   # Wait for ESO to sync (1 hour) or force sync
   # Restart all services to load new public key
   kubectl rollout restart deployment/auth-service -n default
   kubectl rollout restart deployment/product-service -n default
   # ... restart all services
   ```

3. **Wait 24 hours** for all active tokens to expire

4. **Update private key:**

   ```bash
   gcloud secrets versions add auth-service-jwt-private-key \
     --data-file=jwt-key-new \
     --project="$PROJECT_ID"

   # Restart auth-service
   kubectl rollout restart deployment/auth-service -n default
   ```

5. **Clean up:**
   ```bash
   rm jwt-key-new jwt-key-new.pub
   ```

### Stripe Key Rotation

Follow Stripe's documented key rotation procedure:
https://stripe.com/docs/keys#rotate-keys

### Automated Rotation (Future Enhancement)

Consider implementing automated rotation using:

- **Google Secret Manager rotation**: Scheduled Cloud Functions
- **Kubernetes CronJobs**: Periodic key generation and update
- **External tools**: cert-manager, sealed-secrets with rotation policies

## Troubleshooting

### Issue 1: ExternalSecret Status is "SecretSyncedError"

**Symptoms:**

```bash
kubectl describe externalsecret SECRET_NAME
# Status: SecretSyncedError
# Message: secret not found in Google Secret Manager
```

**Solution:**

```bash
# Verify secret exists in Google Secret Manager
gcloud secrets describe SECRET_NAME --project="$PROJECT_ID"

# If not found, create it:
./terraform/scripts/populate-secrets.sh
```

### Issue 2: ClusterSecretStore Status is "Invalid"

**Symptoms:**

```bash
kubectl get clustersecretstore gcp-secret-manager
# STATUS: Invalid
```

**Solution:**

```bash
# Check Workload Identity binding
gcloud iam service-accounts get-iam-policy \
  external-secrets-operator@PROJECT_ID.iam.gserviceaccount.com \
  --project="$PROJECT_ID"

# Should show binding for:
# serviceAccount:PROJECT_ID.svc.id.goog[external-secrets/external-secrets-sa]

# If missing, re-apply Terraform:
cd terraform/environments/prod
terraform apply
```

### Issue 3: Permission Denied Errors

**Symptoms:**

```bash
# ESO logs show:
# Error: permission denied accessing secret
```

**Solution:**

```bash
# Verify IAM role binding
gcloud projects get-iam-policy PROJECT_ID \
  --flatten="bindings[].members" \
  --filter="bindings.members:serviceAccount:external-secrets-operator@PROJECT_ID.iam.gserviceaccount.com"

# Should show: roles/secretmanager.secretAccessor

# If missing, add the role:
gcloud projects add-iam-policy-binding PROJECT_ID \
  --member="serviceAccount:external-secrets-operator@PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

### Issue 4: Secrets Not Syncing After Update

**Symptoms:**

- Updated secret in Google Secret Manager
- Kubernetes secret still shows old value after 1+ hours

**Solution:**

```bash
# Check ESO pod logs
kubectl logs -n external-secrets deployment/external-secrets -f

# Force resync by deleting the ExternalSecret
kubectl delete externalsecret SECRET_NAME -n NAMESPACE
kubectl apply -f deployments/k8s/base/external-secrets/SECRET_FILE.yaml

# Restart the application pod
kubectl rollout restart deployment/SERVICE_NAME -n NAMESPACE
```

### Issue 5: Application Can't Read Secret

**Symptoms:**

- Application logs show: "file not found" or "permission denied"
- Secret exists and is synced

**Solution:**

```bash
# Verify secret is mounted correctly
kubectl describe pod POD_NAME -n NAMESPACE

# Check volume mounts
# Volumes:
#   jwt-keys:
#     Type:        Secret (a volume populated by a Secret)
#     SecretName:  auth-service-jwt-keys

# Verify files in container
kubectl exec -it POD_NAME -n NAMESPACE -- ls -la /app/keys/
# Should show: private-key.pem, public-key.pem

# Check file permissions
kubectl exec -it POD_NAME -n NAMESPACE -- cat /app/keys/private-key.pem | head -c 50
```

## Security Best Practices

### 1. Principle of Least Privilege

- ✅ ESO service account only has `secretmanager.secretAccessor` role
- ✅ No `secretmanager.admin` or write permissions
- ✅ Each service only accesses secrets it needs

### 2. Audit Logging

Enable Cloud Audit Logs for Secret Manager:

```bash
# View secret access logs
gcloud logging read "resource.type=secretmanager.googleapis.com" \
  --project="$PROJECT_ID" \
  --limit=50
```

### 3. Secret Versioning

- ✅ Google Secret Manager maintains all versions
- ✅ Can rollback to previous version if needed
- ✅ Audit trail of all changes

### 4. Workload Identity

- ✅ No service account keys in Git or containers
- ✅ Short-lived, automatically rotated tokens
- ✅ Follows Google Cloud best practices

### 5. Encryption at Rest

- ✅ All secrets encrypted by Google's infrastructure
- ✅ Optional: Use Customer-Managed Encryption Keys (CMEK)

```bash
# Enable CMEK (optional, additional cost)
gcloud secrets create SECRET_NAME \
  --replication-policy="automatic" \
  --kms-key-name="projects/PROJECT_ID/locations/LOCATION/keyRings/KEYRING/cryptoKeys/KEY"
```

### 6. Network Security

- ✅ GKE nodes are private (no public IPs)
- ✅ Secrets transmitted over Google's private network
- ✅ TLS encryption for all external communication

### 7. RBAC for Kubernetes Secrets

Restrict access to Kubernetes secrets:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: secret-reader
  namespace: default
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    resourceNames: ["auth-service-jwt-keys"] # Specific secret
    verbs: ["get"]
```

## Cost Considerations

### Google Secret Manager Pricing

**Storage Cost:**

- $0.06 per secret per month
- 50 secrets = **$3.00/month**

**Access Cost:**

- $0.03 per 10,000 accesses
- With 50 secrets, hourly sync, 9 ExternalSecrets = ~6,500 accesses/month
- **< $0.01/month**

**Total Estimated Cost: ~$3/month**

### Cost Optimization Tips

1. **Reduce refresh frequency** for static secrets:

   ```yaml
   spec:
     refreshInterval: 24h # Instead of 1h
   ```

2. **Use secret reuse**: Multiple ExternalSecrets can reference the same Google Secret Manager secret (e.g., `jwt-public-key`)

3. **Consolidate secrets**: Combine multiple related secrets into one JSON secret:
   ```bash
   echo -n '{"username":"user","password":"pass"}' | gcloud secrets versions add app-secrets --data-file=-
   ```

### Cost vs. Benefits

| Cost     | Benefit                           |
| -------- | --------------------------------- |
| $3/month | Centralized secret management     |
|          | Audit logging and compliance      |
|          | Automatic secret rotation support |
|          | Reduced security incidents        |
|          | Faster incident response          |

**ROI**: A single security incident costs **$1,000+** to investigate and remediate. This solution pays for itself many times over.

## Next Steps

After completing the External Secrets setup:

1. **Monitor secret sync status:**

   ```bash
   kubectl get externalsecrets -A --watch
   ```

2. **Set up alerting** for failed secret syncs:
   - Prometheus alerts for ESO metrics
   - Google Cloud Monitoring for Secret Manager access

3. **Document your secrets** in a secure location:
   - Which secrets exist
   - Rotation schedule
   - Who has access

4. **Implement secret rotation** for critical secrets:
   - Create rotation runbooks
   - Schedule rotation reminders
   - Consider automation

5. **Review and audit** regularly:
   - Quarterly access review
   - Remove unused secrets
   - Update permissions as needed

## Support and Resources

- **External Secrets Operator Docs**: https://external-secrets.io/
- **Google Secret Manager Docs**: https://cloud.google.com/secret-manager/docs
- **Workload Identity Docs**: https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity
- **This Repository's Issues**: For project-specific questions

---

**Questions?** Check the main Terraform README or open an issue in the repository.
