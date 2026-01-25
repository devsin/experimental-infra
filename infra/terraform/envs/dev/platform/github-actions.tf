locals {
  github_oidc_issuer = "https://token.actions.githubusercontent.com"
}

resource "google_iam_workload_identity_pool" "github" {
  workload_identity_pool_id = "github"
  display_name              = "GitHub Actions"
  description               = "OIDC pool for GitHub Actions workflows"
}

resource "google_iam_workload_identity_pool_provider" "github" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = "github-oidc"
  display_name                       = "GitHub OIDC"
  description                        = "Trust GitHub Actions OIDC tokens for selected repo"
  attribute_mapping = {
    "google.subject"          = "assertion.sub"
    "attribute.actor"         = "assertion.actor"
    "attribute.aud"           = "assertion.aud"
    "attribute.repository"    = "assertion.repository"
    "attribute.repository_id" = "assertion.repository_id"
    "attribute.ref"           = "assertion.ref"
  }

  oidc {
    issuer_uri = local.github_oidc_issuer
  }

  # Restrict to the specified repo.
  attribute_condition = "assertion.repository == \"${var.github_repository}\""
}

resource "google_service_account" "github_actions" {
  account_id   = "gha-cicd"
  display_name = "GitHub Actions CI/CD"
  description  = "Build and push images from GitHub Actions"
}

resource "google_service_account_iam_binding" "github_actions_wiu" {
  service_account_id = google_service_account.github_actions.name
  role               = "roles/iam.workloadIdentityUser"
  members = [
    "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/${var.github_repository}"
  ]
}

resource "google_project_iam_member" "github_actions_ar_writer" {
  project = data.terraform_remote_state.bootstrap.outputs.project_id
  role    = "roles/artifactregistry.writer"
  member  = "serviceAccount:${google_service_account.github_actions.email}"
}

output "github_actions_service_account_email" {
  value       = google_service_account.github_actions.email
  description = "Service account email to set as GCP_SERVICE_ACCOUNT GitHub secret"
}

output "github_actions_workload_identity_provider" {
  value       = google_iam_workload_identity_pool_provider.github.name
  description = "OIDC provider resource name for GCP_WORKLOAD_IDENTITY_PROVIDER GitHub secret"
}