resource "kubernetes_namespace_v1" "argocd" {
  metadata {
    name = "argocd"
    labels = {
      "app.kubernetes.io/name"       = "argocd"
      "app.kubernetes.io/part-of"    = "argocd"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }

  depends_on = [data.terraform_remote_state.bootstrap]
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
        type = "LoadBalancer"
      }
    }
  })]

  depends_on = [kubernetes_namespace_v1.argocd]
}
