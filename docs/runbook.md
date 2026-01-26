# CyberArk GitOps Platform Runbook

## Prereqs
- gcloud CLI authenticated with access to the target project
- kubectl installed
- Terraform installed (>=1.6)
- ArgoCD CLI (optional for troubleshooting)


## Access / Day 0 commands
- Login to GCP: `gcloud auth login`
- Configure kube context: `gcloud container clusters get-credentials cyberark-gke-dev-c322b4 --zone europe-west2 --project cyberark-gitops-c322b4`
- Get ArgoCD server IP: `kubectl get svc -n argocd argocd-server -o jsonpath='{.status.loadBalancer.ingress[0].ip}'`
- Access the UI at `https://<IP_ADDRESS>` with username `admin` and password from: `kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d; echo`

## Bootstrapping / Day 1
1) Configure Terraform variables:
	- Infra stack: `infra/terraform/envs/dev/bootstrap/terraform.tfvars` (project, network, GKE, artifact registry).
	- Platform stack: `infra/terraform/envs/dev/platform/terraform.tfvars` (Git repo URL, region if needed).
2) Bootstrap apply: `terraform -chdir=infra/terraform/envs/dev/bootstrap init` then `terraform -chdir=infra/terraform/envs/dev/bootstrap apply`.
3) Platform apply: `terraform -chdir=infra/terraform/envs/dev/platform init` then `terraform -chdir=infra/terraform/envs/dev/platform apply`.
	- Installs ArgoCD via Helm in namespace `argocd` and renders the root Application pointing to `gitops/root`.
4) Wait for ArgoCD to sync; namespaces and add-on Applications will appear.
5) `db-bootstrap` jobs run (waves: SA/ClusterRole −2, ClusterRoleBinding −1, secrets job 0, DB init jobs 1) to seed app secrets and DB users before app deployments (wave 5).
6) Observability stack (`kube-prometheus-stack`, `tempo`, `otel-collector`) deploys into `observability` via ArgoCD.

## GitOps Sync Model
- Root kustomization: `gitops/root` includes platform namespaces and add-ons.
- Add-ons: ArgoCD Applications in `gitops/platform/addons` (Postgres, Redis) target namespace `db`.
- Apps: future business services live under `gitops/apps` (base/overlays) and are referenced by ArgoCD Applications.

## Common Operations
- **Force ArgoCD sync (UI/CLI):**
	- UI: select app → Sync → Apply.
	- CLI: `argocd app sync <app-name>` (requires ArgoCD login/context).
- **Check app health:** ArgoCD UI/CLI; look for OutOfSync/Degraded.
- **Roll back add-on:** revert the Git commit (e.g., chart version/values) and allow ArgoCD to reconcile.
- **Upgrade add-on:** edit chart `targetRevision` or `helm.values` in `gitops/platform/addons/*.yaml`; commit and ArgoCD will apply on next sync.

## Accessing Postgres / Redis (dev)
- Ensure port-forwarding from local to cluster service:
	- Postgres: `kubectl port-forward -n db svc/postgres-postgresql 5432:5432`
	- Redis: `kubectl port-forward -n db svc/redis-master 6379:6379`
- Default credentials come from chart-generated secrets in `db` namespace (e.g., `postgres-postgresql` and `redis` releases).

## App data bootstrap (accounts/transfers)
- Secrets: `gitops/apps/db-bootstrap/job-secrets.yaml` creates `accounts-db`, `transfers-db`, and mirrors `redis` into `apps` namespace.
- DB users/schemas: `job-accounts.yaml` and `job-transfers.yaml` create or update roles and databases with passwords from the secrets.
- Jobs are immutable; on spec changes delete and reapply (or ArgoCD Sync with Replace).
- Passwords are URL-safe (hex) to avoid breaking `DATABASE_URL`.

## Observability
- Metrics: kube-prometheus-stack exposes Prometheus/Grafana/Alertmanager in `observability` (ClusterIP). Port-forward Grafana: `kubectl -n observability port-forward svc/kube-prom-stack-grafana 3000:80`.
- Traces: Tempo in `observability`, receiving from Otel Collector.
- Otel Collector endpoint for apps: `http://otel-collector.observability.svc.cluster.local:4317` (gRPC) or `:4318` (HTTP).
- Recommended app envs:
  - `OTEL_SERVICE_NAME=<service>`
  - `OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector.observability.svc.cluster.local:4317`
  - `OTEL_TRACES_SAMPLER=parentbased_traceidratio`
  - `OTEL_TRACES_SAMPLER_ARG=0.1`

## Troubleshooting
- **ArgoCD cannot fetch Bitnami charts:** using HTTPS repo, no OCI auth required. If switching to OCI, create a Docker registry secret in `argocd` and configure repository credentials in ArgoCD.
- **App stuck OutOfSync:** check events; ensure `CreateNamespace=true` and `ServerSideApply=true` are set (already configured in add-on apps).
- **Terraform apply fails on namespace drift:** the `argocd` namespace is managed; avoid manual deletes. Re-run `terraform apply` to recreate.
- **GKE access:** `gcloud container clusters get-credentials <cluster> --zone <zone> --project <project>`.
- **State drift:** rerun `terraform apply`; for persistent errors, inspect `terraform state list` and provider logs.
- **CreateContainerConfigError (secret not found):** ensure `bootstrap-db-secrets` job succeeded; if not, `kubectl -n db delete job bootstrap-db-secrets && kubectl -n db apply -f gitops/apps/db-bootstrap/job-secrets.yaml`, then restart the app deployment.
- **Postgres auth failed after password rotation:** rerun `bootstrap-accounts-db` / `bootstrap-transfers-db` to `ALTER USER` with the current secret, then restart deployments.
- **Job spec changed (immutable template error):** delete the Job before reapplying, or set `argocd.argoproj.io/sync-options: Replace=true`.
- **Traces missing:** confirm services point OTLP to `otel-collector.observability.svc.cluster.local:4317` and sampling env vars are set; check collector logs: `kubectl -n observability logs deploy/otel-collector`.

## Secrets & Credentials
- Charts are fetched over HTTPS; no registry creds needed by default.
- Database and Redis passwords are stored in Kubernetes secrets generated by the Helm charts; rotate by `helm`/ArgoCD values changes or by resetting secrets and re-syncing.

## Backups / Recovery (future)
- Add backup jobs (e.g., Velero or pgBackRest) via `gitops/platform/addons` when required.
