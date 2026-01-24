output "project_id" {
  value = module.project.project_id
}

output "cluster_name" {
  value = module.gke.name
}

output "cluster_endpoint" {
  value     = module.gke.endpoint
  sensitive = true
}

output "cluster_ca_certificate" {
  value     = module.gke.ca_certificate
  sensitive = true
}

output "artifact_registry_url" {
  value = module.artifact_registry.repository_url
}
