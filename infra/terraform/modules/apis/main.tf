variable "project_id" {
  type = string
}

variable "services" {
  description = "APIs to enable in the project"
  type        = list(string)
  default = [
    "container.googleapis.com",
    "artifactregistry.googleapis.com",
    "compute.googleapis.com",
    "iam.googleapis.com",
  ]
}

resource "google_project_service" "services" {
  for_each = toset(var.services)

  project            = var.project_id
  service            = each.value
  disable_on_destroy = false
}

output "enabled_services" {
  value = var.services
}
