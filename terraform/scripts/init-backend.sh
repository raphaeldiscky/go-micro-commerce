#!/bin/bash
# Initialize Terraform GCS backend for state storage
# This script creates the GCS bucket for Terraform state if it doesn't exist

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TERRAFORM_DIR="${SCRIPT_DIR}/.."

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    log_error "gcloud CLI is not installed. Please install it first."
    log_info "Visit: https://cloud.google.com/sdk/docs/install"
    exit 1
fi

# Check if terraform is installed
if ! command -v terraform &> /dev/null; then
    log_error "Terraform is not installed. Please install it first."
    log_info "Visit: https://developer.hashicorp.com/terraform/install"
    exit 1
fi

# Environment (default: prod)
ENV="${1:-prod}"

if [[ "$ENV" != "prod" ]]; then
    log_error "Invalid environment: $ENV"
    log_info "Usage: $0 [prod]"
    exit 1
fi

log_info "Initializing Terraform backend for environment: $ENV"

# Get project ID from gcloud config
PROJECT_ID=$(gcloud config get-value project 2>/dev/null)

if [[ -z "$PROJECT_ID" ]]; then
    log_error "No GCP project configured. Please run: gcloud config set project YOUR_PROJECT_ID"
    exit 1
fi

log_info "Using GCP project: $PROJECT_ID"

# Backend bucket name
BUCKET_NAME="go-micro-commerce-terraform-state-${ENV}"

# Check if bucket exists
if gsutil ls -b "gs://${BUCKET_NAME}" &> /dev/null; then
    log_info "Bucket gs://${BUCKET_NAME} already exists"
else
    log_info "Creating GCS bucket: gs://${BUCKET_NAME}"

    # Create bucket with versioning and encryption
    gsutil mb -p "${PROJECT_ID}" -l asia-southeast1 "gs://${BUCKET_NAME}"

    # Enable versioning for state file recovery
    gsutil versioning set on "gs://${BUCKET_NAME}"

    # Set lifecycle policy to keep last 10 versions
    cat > /tmp/lifecycle.json <<EOF
{
  "lifecycle": {
    "rule": [
      {
        "action": {"type": "Delete"},
        "condition": {
          "numNewerVersions": 10
        }
      }
    ]
  }
}
EOF
    gsutil lifecycle set /tmp/lifecycle.json "gs://${BUCKET_NAME}"
    rm /tmp/lifecycle.json

    log_info "Bucket created successfully with versioning enabled"
fi

# Initialize Terraform
ENV_DIR="${TERRAFORM_DIR}/environments/${ENV}"

if [[ ! -d "$ENV_DIR" ]]; then
    log_error "Environment directory not found: $ENV_DIR"
    exit 1
fi

cd "$ENV_DIR"

log_info "Initializing Terraform in: $ENV_DIR"

# Copy shared backend configuration
cp ../../shared/backend.tf .
cp ../../shared/versions.tf .

# Initialize with backend config
terraform init -backend-config=backend.hcl -reconfigure

log_info "Terraform backend initialized successfully!"
log_info "State will be stored in: gs://${BUCKET_NAME}/env/${ENV}"
log_warn "Remember to update terraform.tfvars with your project-specific values"
