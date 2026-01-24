variable "project_id" {
  type = string
}

variable "name" {
  type = string
}

variable "billing_account" {
  description = "Billing account ID (e.g. 000000-000000-000000)"
  type        = string
}

variable "folder_id" {
  description = "Optional folder ID (numeric)."
  type        = string
  default     = null
}

variable "org_id" {
  description = "Optional organization ID (numeric). Used when folder_id is not provided."
  type        = string
  default     = null
}

variable "labels" {
  type    = map(string)
  default = {}
}

variable "auto_create_network" {
  description = "Whether to create the default network in the project. Set false only if Compute API is pre-enabled."
  type        = bool
  default     = true
}

variable "deletion_policy" {
  description = "DELETE to allow Terraform to delete the project, PREVENT to block deletes."
  type        = string
  default     = "DELETE"
}

locals {
  parent_org    = var.folder_id == null ? var.org_id : null
  parent_folder = var.folder_id
}

resource "google_project" "project" {
  project_id          = var.project_id
  name                = var.name
  billing_account     = var.billing_account
  org_id              = local.parent_org
  folder_id           = local.parent_folder
  auto_create_network = var.auto_create_network
  deletion_policy     = var.deletion_policy
  labels              = var.labels
}

output "project_id" {
  value = google_project.project.project_id
}

output "project_number" {
  value = google_project.project.number
}
