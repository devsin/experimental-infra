# ADR 0003: Add-ons as ArgoCD Applications

## Status
Accepted

## Context
Need Git-managed add-ons with easy upgrades/rollbacks.

## Decision
- Define each add-on as an ArgoCD Application under `gitops/platform/addons`.
- Store chart version/values in Git; automated sync with prune/self-heal and ServerSideApply.
- Target `db` namespace for data services; tune via Helm values.

## Consequences
- Upgrades/rollbacks are commits; ArgoCD reconciles on sync.
- Add-ons stay out of Terraform state, easing iteration.
- Future add-ons follow the same pattern and are included via `gitops/root`.
