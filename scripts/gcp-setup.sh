#!/bin/bash
# GCP Setup Script for Talos Kubernetes Cluster
# This script enables required GCP APIs and creates service accounts

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}======================================${NC}"
echo -e "${GREEN}GCP Setup for Talos Kubernetes Cluster${NC}"
echo -e "${GREEN}======================================${NC}"
echo ""

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}Error: gcloud CLI is not installed${NC}"
    echo "Install from: https://cloud.google.com/sdk/docs/install"
    exit 1
fi

# Get current project
PROJECT_ID=$(gcloud config get-value project 2>/dev/null)
if [ -z "$PROJECT_ID" ]; then
    echo -e "${RED}Error: No GCP project configured${NC}"
    echo "Run: gcloud config set project YOUR_PROJECT_ID"
    exit 1
fi

echo -e "${GREEN}Using GCP Project:${NC} $PROJECT_ID"
echo ""

# Step 1: Enable Required APIs
echo -e "${YELLOW}Step 1: Enabling Required GCP APIs...${NC}"
APIS=(
    "compute.googleapis.com"
    "secretmanager.googleapis.com"
    "storage-api.googleapis.com"
    "iam.googleapis.com"
)

for api in "${APIS[@]}"; do
    echo "  - Enabling $api..."
    gcloud services enable $api --project=$PROJECT_ID
done

echo -e "${GREEN}✓ APIs enabled successfully${NC}"
echo ""

# Step 2: Create Service Account for External Secrets Operator
echo -e "${YELLOW}Step 2: Creating Service Account for External Secrets Operator...${NC}"

SA_NAME="external-secrets-operator"
SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"

# Check if service account already exists
if gcloud iam service-accounts describe $SA_EMAIL --project=$PROJECT_ID &>/dev/null; then
    echo -e "${YELLOW}  Service account $SA_NAME already exists${NC}"
else
    echo "  - Creating service account: $SA_NAME"
    gcloud iam service-accounts create $SA_NAME \
        --display-name="External Secrets Operator" \
        --description="Service account for ESO to access Secret Manager" \
        --project=$PROJECT_ID
    echo -e "${GREEN}  ✓ Service account created${NC}"
fi

# Step 3: Grant IAM Permissions
echo -e "${YELLOW}Step 3: Granting IAM Permissions...${NC}"

echo "  - Granting roles/secretmanager.secretAccessor..."
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="roles/secretmanager.secretAccessor" \
    --condition=None \
    > /dev/null 2>&1

echo -e "${GREEN}✓ IAM permissions granted${NC}"
echo ""

# Step 4: Create Service Account Key
echo -e "${YELLOW}Step 4: Creating Service Account Key...${NC}"

KEY_FILE="${HOME}/external-secrets-sa-key.json"

if [ -f "$KEY_FILE" ]; then
    echo -e "${YELLOW}  Warning: Key file already exists at $KEY_FILE${NC}"
    read -p "  Do you want to create a new key? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "  Skipping key creation"
    else
        rm -f "$KEY_FILE"
        echo "  - Creating new service account key..."
        gcloud iam service-accounts keys create "$KEY_FILE" \
            --iam-account=$SA_EMAIL \
            --project=$PROJECT_ID
        echo -e "${GREEN}  ✓ Service account key created${NC}"
    fi
else
    echo "  - Creating service account key..."
    gcloud iam service-accounts keys create "$KEY_FILE" \
        --iam-account=$SA_EMAIL \
        --project=$PROJECT_ID
    echo -e "${GREEN}  ✓ Service account key created${NC}"
fi

echo ""
echo -e "${GREEN}Key saved to:${NC} $KEY_FILE"
echo -e "${YELLOW}⚠️  IMPORTANT: Keep this key secure and never commit it to git!${NC}"
echo ""

# Step 5: Get VM Information
echo -e "${YELLOW}Step 5: Fetching VM Information...${NC}"
echo ""
echo "Your GCP Compute Engine VMs in asia-southeast2-a:"
echo ""
gcloud compute instances list --zones=asia-southeast2-a --project=$PROJECT_ID \
    --format="table(name,networkInterfaces[0].networkIP,networkInterfaces[0].accessConfigs[0].natIP,status)" \
    2>/dev/null || echo "No VMs found or insufficient permissions"

echo ""
echo -e "${GREEN}======================================${NC}"
echo -e "${GREEN}Setup Complete!${NC}"
echo -e "${GREEN}======================================${NC}"
echo ""
echo "Next steps:"
echo "1. Copy the VM internal IPs above to terraform.tfvars"
echo "2. Run: ./scripts/create-gcp-secrets.sh to create secrets"
echo "3. Install Talos on VMs: ./scripts/talos-kexec-install.sh <vm-name>"
echo "4. Deploy cluster: cd terraform/environments/production && terraform apply"
echo ""
