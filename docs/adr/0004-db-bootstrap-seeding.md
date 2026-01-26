# ADR 0004: Database Credential Seeding via Kustomize Jobs

## Status
Accepted

## Context
Application pods need database credentials and ready-made roles/databases before they start. Secrets are namespace-scoped and passwords may rotate. We also mirror Redis credentials into the `apps` namespace for the transfers service.

## Decision
- Use Kustomize-managed Jobs under `gitops/apps/db-bootstrap` to:
  - Generate URL-safe random passwords.
  - Create `accounts-db` and `transfers-db` secrets in both `apps` and `db` namespaces.
  - Mirror the Redis password from `db` into `apps` as `secret/redis`.
  - Create or update PostgreSQL roles and databases for accounts/transfers (ALTER USER if the role exists).
- Order with ArgoCD sync waves: SA/ClusterRole (-2), ClusterRoleBinding (-1), secrets job (0), DB init jobs (1), app deployments (5).

## Consequences
- Jobs must be deleted/recreated on pod template changes (immutable), or synced with `argocd.argoproj.io/sync-options: Replace=true`.
- Password rotations driven by secret regen are reconciled automatically via the ALTER USER step; app restarts still needed to pick up new env vars.
- Secrets are duplicated between namespaces; External Secrets/SOPS could replace this later to avoid mirroring.
