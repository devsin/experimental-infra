variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "repository_id" {
  type = string
}

variable "format" {
  type    = string
  default = "DOCKER"
}

resource "google_artifact_registry_repository" "docker" {
  project       = var.project_id
  location      = var.region
  repository_id = var.repository_id
  format        = var.format
}

output "repository_id" {
  value = google_artifact_registry_repository.docker.repository_id
}

output "repository_url" {
  # need to verify this format
  value = format("%s-docker.pkg.dev/%s/%s", var.region, var.project_id, google_artifact_registry_repository.docker.repository_id)
}
