output "project_number" {
  value = data.google_project.project.number
}

output "external_dns_service_account_email" {
  description = "Email of the external-dns GCP service account"
  value       = google_service_account.external_dns.email
}