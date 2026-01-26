# ADR 0005: Observability Stack via ArgoCD (Prometheus/Grafana, Tempo, OTel Collector)

## Status
Accepted

## Context
We need cluster-wide metrics, dashboards, alerting, and distributed tracing for Go services, managed the same way as other platform add-ons (GitOps/ArgoCD).

## Decision
- Add `gitops/platform/observability` with three ArgoCD Applications:
  - **kube-prometheus-stack** (Prometheus Operator, Grafana, Alertmanager) from `prometheus-community` charts; ClusterIP services.
  - **Tempo** tracing backend from `grafana` charts (memory backend for dev).
  - **OpenTelemetry Collector** from `open-telemetry` charts, configured to receive OTLP (4317/4318) and export traces to Tempo (insecure gRPC).
- Create namespace `observability`; include the new kustomization from `gitops/root`.
- Recommend app envs: `OTEL_SERVICE_NAME`, `OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector.observability.svc.cluster.local:4317`, sampler `parentbased_traceidratio` with arg `0.1`.

## Consequences
- Observability components reconcile automatically with ArgoCD and can be rolled back via Git.
- Services get a stable OTLP endpoint; sampling can be tuned per service via env vars.
- Memory Tempo backend suits dev; production should switch to durable object storage (S3/GCS) by updating Helm values.
