#!/bin/bash
# Populate Google Secret Manager with secrets for External Secrets Operator
# This script creates all secrets required by the microservices platform

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored message
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    print_error "gcloud CLI is not installed. Please install it first."
    print_info "Visit: https://cloud.google.com/sdk/docs/install"
    exit 1
fi

# Get project ID from terraform.tfvars or prompt
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TERRAFORM_DIR="$SCRIPT_DIR/../environments/prod"

if [ -f "$TERRAFORM_DIR/terraform.tfvars" ]; then
    PROJECT_ID=$(grep -E '^project_id\s*=' "$TERRAFORM_DIR/terraform.tfvars" | cut -d'"' -f2 | tr -d ' ')
fi

if [ -z "$PROJECT_ID" ]; then
    read -p "Enter your GCP Project ID: " PROJECT_ID
fi

if [ -z "$PROJECT_ID" ]; then
    print_error "Project ID is required"
    exit 1
fi

print_info "Using GCP Project: $PROJECT_ID"

# Set the project
gcloud config set project "$PROJECT_ID" > /dev/null 2>&1

# Check if user is authenticated
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q .; then
    print_error "Not authenticated with gcloud. Please run: gcloud auth login"
    exit 1
fi

print_success "Authenticated with gcloud"

# Enable Secret Manager API
print_info "Enabling Secret Manager API..."
gcloud services enable secretmanager.googleapis.com --quiet

# Function to create or update a secret
create_secret() {
    local secret_name="$1"
    local secret_description="$2"
    local prompt_message="$3"
    local is_file="${4:-false}"

    print_info "Processing secret: $secret_name"

    # Check if secret already exists
    if gcloud secrets describe "$secret_name" --project="$PROJECT_ID" &> /dev/null; then
        print_warning "Secret '$secret_name' already exists. Do you want to update it? (y/N)"
        read -r response
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            print_info "Skipping $secret_name"
            return 0
        fi
    else
        # Create the secret
        gcloud secrets create "$secret_name" \
            --replication-policy="automatic" \
            --labels="managed-by=terraform,component=external-secrets" \
            --project="$PROJECT_ID" > /dev/null 2>&1
        print_success "Created secret: $secret_name"
    fi

    # Add secret version
    print_info "$prompt_message"

    if [ "$is_file" = true ]; then
        read -p "Enter file path: " file_path
        if [ ! -f "$file_path" ]; then
            print_error "File not found: $file_path"
            return 1
        fi
        gcloud secrets versions add "$secret_name" \
            --data-file="$file_path" \
            --project="$PROJECT_ID" > /dev/null 2>&1
    else
        read -s -p "Enter secret value (hidden): " secret_value
        echo
        if [ -z "$secret_value" ]; then
            print_warning "Empty value provided for $secret_name. Skipping..."
            return 0
        fi
        echo -n "$secret_value" | gcloud secrets versions add "$secret_name" \
            --data-file=- \
            --project="$PROJECT_ID" > /dev/null 2>&1
    fi

    print_success "Added secret version for: $secret_name"
}

# Banner
echo
print_info "========================================="
print_info "  Google Secret Manager Setup"
print_info "  External Secrets Operator"
print_info "========================================="
echo

# Critical Secrets
print_info "========================================="
print_info "  CRITICAL SECRETS"
print_info "========================================="
echo

# Auth Service - JWT Keys
print_info "Auth Service JWT Keys"
print_info "Generate RSA keys using: ssh-keygen -t rsa -b 4096 -m PEM -f jwt-key"
create_secret "auth-service-jwt-private-key" "Auth service JWT private key" \
    "JWT Private Key (jwt-key)" true
create_secret "jwt-public-key" "JWT public key for all services" \
    "JWT Public Key (jwt-key.pub)" true

echo

# Payment Service - Stripe Keys
print_info "Payment Service - Stripe Keys"
print_info "Get your keys from: https://dashboard.stripe.com/apikeys"
create_secret "payment-service-stripe-secret-key" "Stripe secret key" \
    "Stripe Secret Key (starts with sk_)"
create_secret "payment-service-stripe-webhook-secret" "Stripe webhook secret" \
    "Stripe Webhook Secret (starts with whsec_)"

echo

# High Priority Secrets
print_info "========================================="
print_info "  HIGH PRIORITY SECRETS"
print_info "========================================="
echo

# Notification Service - SMTP
print_info "Notification Service - SMTP Configuration"
create_secret "notification-service-smtp-host" "SMTP server host" \
    "SMTP Host (e.g., smtp.gmail.com)"
create_secret "notification-service-smtp-username" "SMTP username" \
    "SMTP Username/Email"
create_secret "notification-service-smtp-password" "SMTP password" \
    "SMTP Password or App-Specific Password"
create_secret "notification-service-sendgrid-api-key" "SendGrid API key" \
    "SendGrid API Key (optional, press Enter to skip)"

echo

# Search Service - Elasticsearch
print_info "Search Service - Elasticsearch Credentials"
create_secret "search-service-elasticsearch-username" "Elasticsearch username" \
    "Elasticsearch Username (default: elastic)"
create_secret "search-service-elasticsearch-password" "Elasticsearch password" \
    "Elasticsearch Password"

echo

# Order Service - Temporal
print_info "Order Service - Temporal Configuration"
create_secret "order-service-temporal-api-key" "Temporal Cloud API key" \
    "Temporal API Key (optional if using self-hosted, press Enter to skip)"

echo

# Redis Cluster
print_info "Redis Cluster Configuration"
print_info "Generate a strong password using: openssl rand -base64 32"
create_secret "redis-cluster-password" "Redis cluster password" \
    "Redis Cluster Password"

echo

# Medium Priority Secrets
print_info "========================================="
print_info "  MEDIUM PRIORITY SECRETS"
print_info "========================================="
echo

# Grafana
print_info "Grafana Admin Password"
print_info "Generate a strong password using: openssl rand -base64 20"
create_secret "grafana-admin-password" "Grafana admin password" \
    "Grafana Admin Password"

echo

# ArgoCD
print_info "ArgoCD Admin Password"
print_info "Generate a strong password using: openssl rand -base64 20"
create_secret "argocd-admin-password" "ArgoCD admin password" \
    "ArgoCD Admin Password"

echo

# Summary
print_info "========================================="
print_info "  SETUP COMPLETE"
print_info "========================================="
echo

print_success "All secrets have been populated in Google Secret Manager!"
echo
print_info "Next steps:"
print_info "1. Apply Terraform configuration: ./terraform/scripts/apply-prod.sh"
print_info "2. Verify External Secrets Operator: kubectl get externalsecrets -A"
print_info "3. Check synced secrets: kubectl get secrets -A | grep external-secrets"
echo
print_info "To view secrets in Google Secret Manager:"
print_info "  gcloud secrets list --project=$PROJECT_ID"
echo
print_info "To view a specific secret version:"
print_info "  gcloud secrets versions access latest --secret=SECRET_NAME --project=$PROJECT_ID"
echo

print_success "Done!"
