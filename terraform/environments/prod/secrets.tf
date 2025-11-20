# GitHub Container Registry Authentication
# Stores GitHub username and Personal Access Token for pulling private container images

# Create secret for GitHub username
resource "google_secret_manager_secret" "github_username" {
  project   = var.project_id
  secret_id = "github-container-registry-username"

  labels = {
    environment = "production"
    managed-by  = "terraform"
    purpose     = "container-registry-auth"
  }

  replication {
    auto {}
  }
}

# Store the actual username value
resource "google_secret_manager_secret_version" "github_username" {
  secret = google_secret_manager_secret.github_username.id

  secret_data = var.github_username
}

# Grant External Secrets Operator service account access to read the username secret
resource "google_secret_manager_secret_iam_member" "eso_github_username" {
  project   = var.project_id
  secret_id = google_secret_manager_secret.github_username.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.eso_gcp_service_account_name}@${var.project_id}.iam.gserviceaccount.com"

  depends_on = [google_secret_manager_secret.github_username]
}

# Create secret for GitHub token
resource "google_secret_manager_secret" "github_token" {
  project   = var.project_id
  secret_id = "github-container-registry-token"

  labels = {
    environment = "production"
    managed-by  = "terraform"
    purpose     = "container-registry-auth"
  }

  replication {
    auto {}
  }
}

# Store the actual token value
resource "google_secret_manager_secret_version" "github_token" {
  secret = google_secret_manager_secret.github_token.id

  secret_data = var.github_token
}

# Grant External Secrets Operator service account access to read the token secret
resource "google_secret_manager_secret_iam_member" "eso_github_token" {
  project   = var.project_id
  secret_id = google_secret_manager_secret.github_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.eso_gcp_service_account_name}@${var.project_id}.iam.gserviceaccount.com"

  depends_on = [google_secret_manager_secret.github_token]
}
