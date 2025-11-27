#!/bin/bash
# Populate Google Secret Manager with secrets from secrets.json
# This script reads from secrets.json and creates all secrets in Google Secret Manager

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

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    print_error "jq is not installed. Please install it first."
    print_info "Visit: https://stedolan.github.io/jq/download/"
    exit 1
fi

# Get project ID from terraform.tfvars or prompt
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TERRAFORM_DIR="$SCRIPT_DIR/../environments/prod"
SECRETS_DIR="$SCRIPT_DIR/secrets"
SECRETS_FILE="$SECRETS_DIR/secrets.json"

# Check if secrets.json exists
if [ ! -f "$SECRETS_FILE" ]; then
    print_error "secrets.json not found at $SECRETS_FILE"
    print_info "Please create it using secrets.json.example as a template"
    exit 1
fi

# Get project ID
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

# Function to read from JSON
read_secret_from_json() {
    local json_key="$1"
    jq -r ".$json_key // empty" "$SECRETS_FILE"
}

# Function to create or update a secret from JSON
create_secret_from_json() {
    local secret_name="$1"
    local json_key="$2"
    local is_file="${3:-false}"

    print_info "Processing secret: $secret_name"

    # Check if secret already exists
    if gcloud secrets describe "$secret_name" --project="$PROJECT_ID" &> /dev/null; then
        print_info "Updating existing secret: $secret_name"
    else
        # Create the secret
        gcloud secrets create "$secret_name" \
            --replication-policy="automatic" \
            --labels="managed-by=terraform,component=external-secrets" \
            --project="$PROJECT_ID" > /dev/null 2>&1
        print_success "Created secret: $secret_name"
    fi

    # Get value from JSON and add secret version
    if [ "$is_file" = true ]; then
        file_path=$(read_secret_from_json "$json_key")
        if [ -z "$file_path" ]; then
            print_warning "No file path provided for $json_key in secrets.json. Skipping..."
            return 0
        fi

        # Resolve relative paths relative to SECRETS_DIR
        if [[ "$file_path" != /* ]]; then
            file_path="$SECRETS_DIR/$file_path"
        fi

        if [ ! -f "$file_path" ]; then
            print_error "File not found: $file_path"
            return 1
        fi
        gcloud secrets versions add "$secret_name" \
            --data-file="$file_path" \
            --project="$PROJECT_ID" > /dev/null 2>&1
    else
        secret_value=$(read_secret_from_json "$json_key")
        if [ -z "$secret_value" ]; then
            print_warning "No value found for $json_key in secrets.json. Skipping..."
            return 0
        fi
        echo -n "$secret_value" | gcloud secrets versions add "$secret_name" \
            --data-file=- \
            --project="$PROJECT_ID" > /dev/null 2>&1
    fi

    print_success "Added secret version for: $secret_name"
}

# Function to check and generate SSH/RSA keys if missing
check_and_generate_keys() {
    local key_type="$1"        # "jwt" or "argocd-git"
    local key_name="$2"        # "jwt-key" or "argocd-git-key"
    local description="$3"     # Human-readable description

    local private_key="$SECRETS_DIR/$key_name"
    local public_key="$SECRETS_DIR/$key_name.pub"

    # Check if keys exist
    if [ -f "$private_key" ] && [ -f "$public_key" ]; then
        print_success "$description keys already exist"
        return 0
    fi

    if [ -f "$private_key" ] || [ -f "$public_key" ]; then
        print_warning "$description keys partially exist. Please check manually."
        return 1
    fi

    # Keys don't exist - ask user
    echo
    print_warning "$description keys not found at:"
    print_warning "  - $private_key"
    print_warning "  - $public_key"
    echo

    read -p "Would you like to generate them now? (y/n): " generate_keys

    if [[ "$generate_keys" =~ ^[Yy]$ ]]; then
        print_info "Generating $description keys..."

        if [ "$key_type" = "jwt" ]; then
            # Generate RSA key pair for JWT (4096 bits for security)
            ssh-keygen -t rsa -b 4096 -C "jwt-auth-service" -f "$private_key" -N "" -q
        elif [ "$key_type" = "argocd-git" ]; then
            # Generate Ed25519 key pair for Git SSH (modern, secure, fast)
            ssh-keygen -t ed25519 -C "argocd-gke-production" -f "$private_key" -N "" -q
        else
            print_error "Unknown key type: $key_type"
            return 1
        fi

        # Set correct permissions
        chmod 600 "$private_key"
        chmod 644 "$public_key"

        print_success "$description keys generated successfully"
        echo
        print_info "Public key:"
        cat "$public_key"
        echo

        if [ "$key_type" = "argocd-git" ]; then
            print_warning "IMPORTANT: Add the public key above to GitHub Deploy Keys:"
            print_warning "  https://github.com/raphaeldiscky/go-micro-commerce/settings/keys"
        fi

        return 0
    else
        echo
        print_info "To generate keys manually, run:"
        if [ "$key_type" = "jwt" ]; then
            echo "  ssh-keygen -t rsa -b 4096 -C \"jwt-auth-service\" -f \"$private_key\" -N \"\""
        elif [ "$key_type" = "argocd-git" ]; then
            echo "  ssh-keygen -t ed25519 -C \"argocd-gke-production\" -f \"$private_key\" -N \"\""
        fi
        echo "  chmod 600 $private_key"
        echo "  chmod 644 $public_key"
        echo
        return 1
    fi
}

# Banner
echo
print_info "========================================="
print_info "  Google Secret Manager Setup"
print_info "  External Secrets Operator"
print_info "  Loading from secrets.json"
print_info "========================================="
echo

# Check and generate keys if needed
print_info "========================================="
print_info "  Checking Required Keys"
print_info "========================================="
echo

# Check JWT keys
check_and_generate_keys "jwt" "jwt-key" "Auth Service JWT"

# Check ArgoCD Git keys
check_and_generate_keys "argocd-git" "argocd-git-key" "ArgoCD Git SSH"

echo
print_info "========================================="
print_info "  Uploading Secrets to GSM"
print_info "========================================="
echo

# Auth Service - JWT Keys
print_info "Auth Service JWT Keys"
create_secret_from_json "auth-service-jwt-private-key" "auth_service_jwt_private_key_file" true
create_secret_from_json "jwt-public-key" "auth_service_jwt_public_key_file" true

echo

# ArgoCD - Git Credentials
print_info "ArgoCD Git Credentials"
create_secret_from_json "argocd-git-ssh-private-key" "argocd_git_ssh_private_key_file" true

echo

# Payment Service - Stripe Keys
print_info "Payment Service - Stripe Keys"
create_secret_from_json "payment-service-stripe-secret-key" "payment_stripe_secret_key"
create_secret_from_json "payment-service-stripe-webhook-secret" "payment_stripe_webhook_secret"

echo

# Redis Cluster
print_info "Redis Cluster Configuration"
create_secret_from_json "redis-cluster-password" "redis_cluster_password"

echo

# Grafana
print_info "Grafana Admin Password"
create_secret_from_json "grafana-admin-password" "grafana_admin_password"

echo

# ArgoCD
print_info "ArgoCD Admin Password"
create_secret_from_json "argocd-admin-password" "argocd_admin_password"

echo

# GitHub Container Registry
print_info "GitHub Container Registry - Authentication"
create_secret_from_json "github-container-registry-username" "github_username"
create_secret_from_json "github-container-registry-token" "github_token"

echo

# Notification Service - SendGrid
print_info "Notification Service - SendGrid Configuration"
create_secret_from_json "notification-service-sendgrid-api-key" "sendgrid_api_key"

echo

# Summary
print_info "========================================="
print_info "  SETUP COMPLETE"
print_info "========================================="
echo

print_success "All secrets have been populated in Google Secret Manager!"
echo
print_info "Next steps:"
print_info "1. Verify secrets: gcloud secrets list --project=$PROJECT_ID"
print_info "2. Apply Terraform configuration: ./terraform/scripts/apply-prod.sh"
print_info "3. Verify External Secrets Operator: kubectl get externalsecrets -A"
print_info "4. Check synced secrets: kubectl get secrets -A | grep external-secrets"
echo

print_success "Done!"
