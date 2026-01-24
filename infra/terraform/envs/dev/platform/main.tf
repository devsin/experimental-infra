terraform {
  required_version = ">= 1.5.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.32"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.13"
    }
  }
}

data "terraform_remote_state" "bootstrap" {
  backend = "local"
  config = {
    path = "../bootstrap/terraform.tfstate"
  }
}

data "google_client_config" "default" {}

locals {
  kube_host  = "https://${data.terraform_remote_state.bootstrap.outputs.cluster_endpoint}"
  kube_ca    = base64decode(data.terraform_remote_state.bootstrap.outputs.cluster_ca_certificate)
  kube_token = data.google_client_config.default.access_token
}

provider "google" {
  project = data.terraform_remote_state.bootstrap.outputs.project_id
  region  = var.region
}

provider "kubernetes" {
  host                   = local.kube_host
  token                  = local.kube_token
  cluster_ca_certificate = local.kube_ca
}

provider "helm" {
  kubernetes = {
    host                   = local.kube_host
    token                  = local.kube_token
    cluster_ca_certificate = local.kube_ca
  }
}
