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
