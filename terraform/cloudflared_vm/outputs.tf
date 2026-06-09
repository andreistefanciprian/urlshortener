output "cloudflared_internal_ip" {
  description = "Static internal IP address of the cloudflared VM"
  value       = google_compute_address.cloudflared.address
}
