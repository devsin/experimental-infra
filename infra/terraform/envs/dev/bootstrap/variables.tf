variable "project_id" {
  type = string
}

variable "project_name" {
  type = string
}

variable "billing_account" {
  description = "Billing account ID to attach (e.g. 000000-000000-000000)"
  type        = string
}

variable "folder_id" {
  description = "Optional folder ID to place the project under. Use numeric ID, not full path."
  type        = string
  default     = null
}

variable "org_id" {
  description = "Optional org ID to place the project under when folder_id is not set."
  type        = string
  default     = null
}

variable "project_labels" {
  type    = map(string)
  default = {}
}

variable "region" {
  type    = string
  default = "europe-west2"
}

variable "cluster_name" {
  type    = string
  default = "gitops-demo"
}

variable "deletion_protection" {
  description = "Whether to enable deletion protection on the GKE cluster. Set to false to allow terraform destroy."
  type        = bool
  default     = true
}

variable "network_name" {
  type    = string
  default = "demo-network"
}

variable "subnetwork_cidr" {
  type    = string
  default = "10.10.0.0/20"
}

variable "pods_cidr" {
  type    = string
  default = "10.20.0.0/16"
}

variable "services_cidr" {
  type    = string
  default = "10.30.0.0/20"
}

variable "artifact_repository_id" {
  type    = string
  default = "demo-images"
}

variable "node_pools" {
  type = list(object({
    name         = string
    machine_type = string
    min_count    = number
    max_count    = number
    disk_size_gb = number
    auto_upgrade = optional(bool, true)
    auto_repair  = optional(bool, true)
  }))

  default = [
    {
      name         = "primary"
      machine_type = "e2-standard-2"
      min_count    = 1
      max_count    = 2
      disk_size_gb = 50
      auto_upgrade = true
      auto_repair  = true
    }
  ]
}
