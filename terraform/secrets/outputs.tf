output "cloudflared_service_account_email" {
  description = "Email of the cloudflared VM service account"
  value       = google_service_account.cloudflared.email
}

output "cloudflared_secret_name" {
  description = "Full resource name of the cloudflared tunnel token secret"
  value       = google_secret_manager_secret.secrets["cloudflared-tunnel-token"].name
}

output "cloudflared_secret_id" {
  description = "Short secret ID of the cloudflared tunnel token secret"
  value       = google_secret_manager_secret.secrets["cloudflared-tunnel-token"].secret_id
}

output "external_secrets_service_account_email" {
  description = "Email of the External Secrets Operator GCP service account"
  value       = google_service_account.external_secrets.email
}

output "cloudflare_api_token_secret_id" {
  description = "Short secret ID of the Cloudflare API token secret"
  value       = google_secret_manager_secret.secrets["cloudflare-api-token"].secret_id
}

output "external_dns_service_account_email" {
  description = "Email of the external-dns GCP service account"
  value       = google_service_account.external_dns.email
}
