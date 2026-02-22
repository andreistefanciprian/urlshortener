output "cloudflared_internal_ip" {
  description = "Internal IP address of the cloudflared VM"
  value       = google_compute_instance.cloudflared.network_interface[0].network_ip
}
