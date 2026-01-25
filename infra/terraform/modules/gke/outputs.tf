output "service_account" {
  description = "Default service account used by the GKE nodes"
  value       = module.gke.service_account
}
