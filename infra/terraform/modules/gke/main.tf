variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "cluster_name" {
  type = string
}

variable "network" {
  type = string
}

variable "subnetwork" {
  type = string
}

variable "ip_range_pods_name" {
  type = string
}

variable "ip_range_services_name" {
  type = string
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
}

variable "deletion_protection" {
  type    = bool
  default = true
}

module "gke" {
  source  = "terraform-google-modules/kubernetes-engine/google"
  version = "43.0.0"

  project_id = var.project_id
  name       = var.cluster_name
  region     = var.region

  network    = var.network
  subnetwork = var.subnetwork

  ip_range_pods     = var.ip_range_pods_name
  ip_range_services = var.ip_range_services_name

  remove_default_node_pool = true
  initial_node_count       = 1

  node_pools = var.node_pools

  deletion_protection = var.deletion_protection
}

output "name" {
  value = module.gke.name
}

output "endpoint" {
  value = module.gke.endpoint
}

output "ca_certificate" {
  value     = module.gke.ca_certificate
  sensitive = true
}
