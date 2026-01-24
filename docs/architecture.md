# CyberArk GitOps Platform Architecture

## Overview
- Terraform provisions Google Cloud project resources, networking, and GKE cluster (module stack: project → apis → network → gke → artifact registry).
- ArgoCD is installed via Terraform (Helm provider) into `argocd` namespace and bootstraps the GitOps repo.
- GitOps repo (`gitops/`) is the single source of truth for platform namespaces, add-ons, and application overlays.
- Cluster add-ons include Bitnami PostgreSQL and Redis deployed to `db` namespace and managed by ArgoCD Applications.

## Control Plane
- **ArgoCD installation**: Terraform `helm_release` installs ArgoCD in-cluster; namespace managed by `kubernetes_namespace_v1`.
- **Root Application**: Terraform `kubernetes_manifest` defines the ArgoCD root app pointing at `gitops/root`, which pulls platform and app kustomizations.
- **Sync policy**: Automated with prune and self-heal enabled; server-side apply used for safer diffs.

## GitOps Structure
- `gitops/root`: entrypoint kustomization including platform namespaces and add-ons.
- `gitops/platform/namespaces`: declares shared namespaces (`payments`, `db`).
- `gitops/platform/addons`: ArgoCD Applications for Bitnami PostgreSQL (`postgresql` chart v18.2.3) and Redis (`redis` chart v24.1.2) targeting `db`.
- `gitops/apps`: placeholder for business apps (base + overlays) managed via ArgoCD.

## Data Plane
- **GKE**: Provisioned with terraform-google-kubernetes-engine v43; private control plane endpoint output marked sensitive; cluster credentials consumed by providers for ArgoCD install.
- **Networking**: VPC/subnets via `network` module; default network retained unless disabled; random suffix used for uniqueness.
- **Artifact Registry**: Regional repo for images; URL output constructed in module.

## Workflow
1) `terraform apply` in `infra/terraform/envs/dev` creates/updates GCP resources, installs ArgoCD, and creates the ArgoCD root application.
2) ArgoCD syncs `gitops/root`, which applies namespaces and add-on Applications.
3) Add-ons reconcile Helm charts (Postgres, Redis) into `db` with automated sync/prune.
4) Application teams add overlays under `gitops/apps`; ArgoCD tracks drift and enables Git-driven rollbacks.

## Operations
- **Upgrades**: Bump chart versions or Helm values in `gitops/platform/addons`; ArgoCD applies on next sync.
- **Rollback**: Revert commits in Git; ArgoCD reconciles back to the prior revision.
- **Secrets**: OCI auth not required since charts are pulled from Bitnami HTTPS repo; if OCI is desired, add a pull secret to `argocd` and configure repo creds.
- **Access**: Use ArgoCD UI/CLI for app health; use Terraform for infra drift correction.

## Future Enhancements
- Add SOPS or External Secrets for secret management.
- Add metrics/logging stack (e.g., kube-prometheus-stack + Loki) via GitOps add-ons.
- Harden cluster with network policies and pod security standards.
