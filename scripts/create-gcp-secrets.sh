#!/bin/bash
# Create GCP Secret Manager secrets for go-micro-commerce
# This script creates all required secrets for 9 microservices

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Creating GCP Secret Manager Secrets${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}Error: gcloud CLI is not installed${NC}"
    exit 1
fi

# Check if openssl is installed
if ! command -v openssl &> /dev/null; then
    echo -e "${RED}Error: openssl is not installed${NC}"
    exit 1
fi

# Get current project
PROJECT_ID=$(gcloud config get-value project 2>/dev/null)
if [ -z "$PROJECT_ID" ]; then
    echo -e "${RED}Error: No GCP project configured${NC}"
    exit 1
fi

echo -e "${GREEN}Using GCP Project:${NC} $PROJECT_ID"
echo ""

# Function to create secret
create_secret() {
    local secret_name=$1
    local secret_value=$2

    if gcloud secrets describe $secret_name --project=$PROJECT_ID &>/dev/null; then
        echo -e "${YELLOW}  ⚠ Secret '$secret_name' already exists, skipping${NC}"
    else
        echo -n "$secret_value" | gcloud secrets create $secret_name \
            --data-file=- \
            --replication-policy="automatic" \
            --project=$PROJECT_ID \
            > /dev/null 2>&1
        echo -e "${GREEN}  ✓ Created: $secret_name${NC}"
    fi
}

# Array of services
SERVICES=(
    "auth"
    "product"
    "order"
    "payment"
    "cart"
    "fulfillment"
    "notification"
    "search"
    "chat"
)

echo -e "${YELLOW}Creating PostgreSQL credentials for 9 services...${NC}"
echo ""

# Create PostgreSQL credentials for each service
for service in "${SERVICES[@]}"; do
    echo "Service: $service"

    # Username
    username="${service}_user"
    create_secret "prod-postgres-${service}-username" "$username"

    # Password (32-byte random)
    password=$(openssl rand -base64 32)
    create_secret "prod-postgres-${service}-password" "$password"

    echo ""
done

echo -e "${YELLOW}Creating JWT secrets...${NC}"
echo ""

# JWT secret for auth service (64-byte random)
jwt_secret=$(openssl rand -base64 64)
create_secret "prod-jwt-auth-secret" "$jwt_secret"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}All Secrets Created Successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Total secrets created: 19"
echo "  - PostgreSQL usernames: 9"
echo "  - PostgreSQL passwords: 9"
echo "  - JWT secrets: 1"
echo ""
echo "Verify secrets in GCP Console:"
echo "https://console.cloud.google.com/security/secret-manager?project=$PROJECT_ID"
echo ""
echo "List all secrets:"
echo "  gcloud secrets list --project=$PROJECT_ID"
echo ""
echo "View a specific secret:"
echo "  gcloud secrets versions access latest --secret=prod-postgres-auth-username --project=$PROJECT_ID"
echo ""
