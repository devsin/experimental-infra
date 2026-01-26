# CyberArk GitOps Platform Architecture

## Overview
- Terraform provisions Google Cloud project resources, networking, and GKE cluster (module stack: project → apis → network → gke → artifact registry).
- ArgoCD is installed via Terraform (Helm provider) into `argocd` namespace and bootstraps the GitOps repo.
- GitOps repo (`gitops/`) is the single source of truth for platform namespaces, add-ons, and application overlays.
- Cluster add-ons include Bitnami PostgreSQL and Redis deployed to `db` namespace and managed by ArgoCD Applications. Application data bootstrap (DB users/secrets) is handled by Kustomize-managed Jobs in `gitops/apps/db-bootstrap`.
- Observability stack (kube-prometheus-stack, Tempo, OpenTelemetry Collector) runs in `observability` namespace via ArgoCD applications.

## Control Plane
- **ArgoCD installation**: Terraform `helm_release` installs ArgoCD in-cluster; namespace managed by `kubernetes_namespace_v1`.
- **Root Application**: Terraform `kubernetes_manifest` defines the ArgoCD root app pointing at `gitops/root`, which pulls platform and app kustomizations.
- **Sync policy**: Automated with prune and self-heal enabled; server-side apply used for safer diffs.

## GitOps Structure
- `gitops/root`: entrypoint kustomization including platform namespaces, add-ons, observability, and apps.
- `gitops/platform/namespaces`: declares shared namespaces (`apps`, `db`, `payments`, `observability`).
- `gitops/platform/addons`: ArgoCD Applications for Bitnami PostgreSQL and Redis targeting `db`.
- `gitops/platform/observability`: ArgoCD Applications for kube-prometheus-stack, Tempo, and OpenTelemetry Collector targeting `observability`.
- `gitops/apps`: business apps plus `db-bootstrap` jobs that seed DB creds/users and mirror the Redis secret into `apps`.

## Data Plane
- **GKE**: Provisioned with terraform-google-kubernetes-engine v43; private control plane endpoint output marked sensitive; cluster credentials consumed by providers for ArgoCD install.
- **Networking**: VPC/subnets via `network` module; default network retained unless disabled; random suffix used for uniqueness.
- **Artifact Registry**: Regional repo for images; URL output constructed in module. GKE node service account is granted `roles/artifactregistry.reader` so pods can pull images without per-namespace imagePullSecrets.

## Workflow
1) `terraform apply` in `infra/terraform/envs/dev` creates/updates GCP resources, installs ArgoCD, and creates the ArgoCD root application.
2) ArgoCD syncs `gitops/root`, which applies namespaces and add-on Applications.
3) Add-ons reconcile Helm charts (Postgres, Redis) into `db` with automated sync/prune.
4) `db-bootstrap` jobs run (sync waves −2 → 1) to create DB/Redis secrets in `apps`/`db`, create or update DB users/passwords, and ensure schemas.
5) Observability apps deploy (kube-prometheus-stack, Tempo, Otel Collector) providing metrics and traces endpoints.
6) Application deployments roll out (wave 5) using the seeded secrets and export OTLP traces to the collector.

## Operations
- **Upgrades**: Bump chart versions or Helm values in `gitops/platform/addons`; ArgoCD applies on next sync.
- **Rollback**: Revert commits in Git; ArgoCD reconciles back to the prior revision.
- **Secrets**: OCI auth not required since charts are pulled from Bitnami HTTPS repo; if OCI is desired, add a pull secret to `argocd` and configure repo creds.
- **Access**: Use ArgoCD UI/CLI for app health; use Terraform for infra drift correction.

## Future Enhancements
- Add SOPS or External Secrets for secret management.
- Add metrics/logging stack (e.g., kube-prometheus-stack + Loki) via GitOps add-ons.
- Harden cluster with network policies and pod security standards.
