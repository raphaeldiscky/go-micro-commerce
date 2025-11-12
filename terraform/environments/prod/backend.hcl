# Backend configuration for prod environment
# Usage: terraform init -backend-config=backend.hcl

bucket = "go-micro-commerce-terraform-state-prod"
prefix = "env/prod"

# Optional: Enable state locking (automatically enabled with GCS)
# Optional: Custom encryption key for enhanced security (RECOMMENDED for production)
# encryption_key = "projects/go-micro-commerce-prod/locations/asia-southeast1/keyRings/terraform/cryptoKeys/state"
