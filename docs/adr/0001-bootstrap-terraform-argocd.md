# ADR 0001: Bootstrap with Terraform + ArgoCD

## Status
Accepted

## Context
Need repeatable GKE infra and GitOps bootstrap without manual kubectl; Terraform already handles GCP resources.

## Decision
- Provision project → APIs → network → GKE → artifact registry via Terraform modules.
- Install ArgoCD via Terraform Helm provider in namespace `argocd`.
- Create ArgoCD root Application via Terraform `kubernetes_manifest` pointing to `gitops/root`.
- Enable automated sync with prune/self-heal and ServerSideApply on the root app.

## Consequences
- Single `terraform apply` bootstraps infra and ArgoCD; Git owns state afterward.
- Terraform fixes infra drift; ArgoCD fixes app/add-on drift.
- If ArgoCD is down, app changes pause until it recovers.
