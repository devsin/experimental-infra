data "google_client_config" "default" {}

locals {
  kube_host  = "https://${module.gke.endpoint}"
  kube_ca    = base64decode(module.gke.ca_certificate)
  kube_token = data.google_client_config.default.access_token
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

resource "kubernetes_namespace_v1" "argocd" {
  metadata {
    name = "argocd"
    labels = {
      "app.kubernetes.io/name"       = "argocd"
      "app.kubernetes.io/part-of"    = "argocd"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }

  depends_on = [module.gke]
}

resource "helm_release" "argocd" {
  name       = "argocd"
  namespace  = kubernetes_namespace_v1.argocd.metadata[0].name
  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-cd"
  version    = "9.3.5"

  values = [yamlencode({
    server = {
      service = {
        type = "LoadBalancer" # ClusterIP for port forward
      }
    }
  })]

  depends_on = [kubernetes_namespace_v1.argocd]
}
