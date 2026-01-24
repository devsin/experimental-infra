variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "network_name" {
  type = string
}

variable "subnetwork_cidr" {
  type = string
}

variable "pods_cidr" {
  type = string
}

variable "services_cidr" {
  type = string
}

locals {
  pods_range_name     = "pods"
  services_range_name = "services"
}

resource "google_compute_network" "vpc" {
  name                    = var.network_name
  project                 = var.project_id
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "primary" {
  name          = "${var.network_name}-subnet"
  project       = var.project_id
  ip_cidr_range = var.subnetwork_cidr
  network       = google_compute_network.vpc.id
  region        = var.region

  secondary_ip_range {
    range_name    = local.pods_range_name
    ip_cidr_range = var.pods_cidr
  }

  secondary_ip_range {
    range_name    = local.services_range_name
    ip_cidr_range = var.services_cidr
  }
}

resource "google_compute_router" "router" {
  name    = "${var.network_name}-router"
  project = var.project_id
  region  = var.region
  network = google_compute_network.vpc.id
}

resource "google_compute_router_nat" "nat" {
  name                               = "${var.network_name}-nat"
  project                            = var.project_id
  region                             = var.region
  router                             = google_compute_router.router.name
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}

output "network_name" {
  value = google_compute_network.vpc.name
}

output "subnetwork_name" {
  value = google_compute_subnetwork.primary.name
}

output "pods_range_name" {
  value = local.pods_range_name
}

output "services_range_name" {
  value = local.services_range_name
}
