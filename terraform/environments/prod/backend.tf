# Backend configuration for Terraform state management
# This file should be referenced from each environment using backend-config

terraform {
  backend "gcs" {
    # bucket: Configured per environment via backend.hcl
    # prefix: Configured per environment via backend.hcl

    # Optional: Enable encryption at rest (recommended for production)
    # encryption_key = "projects/<PROJECT>/locations/<LOCATION>/keyRings/<KEYRING>/cryptoKeys/<KEY>"
  }
}
