# ADR 0002: Namespace and Tenancy Layout

## Status
Accepted

## Context
Need simple namespace split for platform data services and the first app team.

## Decision
- Manage namespaces in `gitops/platform/namespaces` (ArgoCD reconciles).
- Shared `db` for Postgres/Redis; `payments` for the payments team.
- ArgoCD apps use `CreateNamespace=true` to avoid ordering issues.

## Consequences
- Namespaces are declarative; manual edits may be reverted.
- RBAC/quotas needed as teams grow; new namespaces added via PRs under `gitops/platform/namespaces`.
