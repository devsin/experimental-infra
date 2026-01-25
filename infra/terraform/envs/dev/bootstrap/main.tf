terraform {
  required_version = ">= 1.5.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.6"
    }
  }
}

resource "random_id" "suffix" {
  byte_length = 3 # 6 hex chars
}

locals {
  suffix               = random_id.suffix.hex
  project_id_unique    = "${var.project_id}-${local.suffix}"
  project_name_unique  = "${var.project_name}-${local.suffix}"
  cluster_name_unique  = "${var.cluster_name}-${local.suffix}"
  network_name_unique  = "${var.network_name}-${local.suffix}"
  artifact_repo_unique = "${var.artifact_repository_id}-${local.suffix}"
}

provider "google" {
  project = local.project_id_unique
  region  = var.region
}

module "project" {
  source          = "../../../modules/project"
  project_id      = local.project_id_unique
  name            = local.project_name_unique
  billing_account = var.billing_account
  folder_id       = var.folder_id
  org_id          = var.org_id
  labels          = var.project_labels
}

module "apis" {
  source     = "../../../modules/apis"
  project_id = module.project.project_id

  depends_on = [module.project]
}

module "network" {
  source          = "../../../modules/network"
  project_id      = module.project.project_id
  region          = var.region
  network_name    = local.network_name_unique
  subnetwork_cidr = var.subnetwork_cidr
  pods_cidr       = var.pods_cidr
  services_cidr   = var.services_cidr

  depends_on = [module.apis]
}

module "gke" {
  source                 = "../../../modules/gke"
  project_id             = module.project.project_id
  region                 = var.region
  cluster_name           = local.cluster_name_unique
  network                = module.network.network_name
  subnetwork             = module.network.subnetwork_name
  ip_range_pods_name     = module.network.pods_range_name
  ip_range_services_name = module.network.services_range_name
  node_pools             = var.node_pools
  deletion_protection    = var.deletion_protection

  depends_on = [module.apis]
}

module "artifact_registry" {
  source        = "../../../modules/artifact_registry"
  project_id    = module.project.project_id
  region        = var.region
  repository_id = local.artifact_repo_unique

  depends_on = [module.apis]
}

# Allow GKE node service account to pull images from Artifact Registry.
resource "google_project_iam_member" "gke_nodes_artifact_registry_reader" {
  project = module.project.project_id
  role    = "roles/artifactregistry.reader"
  member  = "serviceAccount:${module.gke.service_account}"
}
