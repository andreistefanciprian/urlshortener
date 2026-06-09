locals {
  labels = merge({
    managed_by = "terraform"
    env        = var.project_name
  }, var.labels)
}

# Static internal IP — keeps the DNS record stable across VM recreations
resource "google_compute_address" "cloudflared" {
  name         = "${var.project_name}-cloudflared"
  region       = var.gcp_region
  subnetwork   = data.terraform_remote_state.networking.outputs.subnet_self_link
  address_type = "INTERNAL"
}

# cloudflared tunnel connector VM
resource "google_compute_instance" "cloudflared" {
  name         = "${var.project_name}-cloudflared"
  machine_type = "e2-small"
  zone         = var.gcp_zone

  boot_disk {
    initialize_params {
      image = data.google_compute_image.ubuntu.self_link
      size  = 10
      type  = "pd-standard"
    }
  }

  network_interface {
    subnetwork = data.terraform_remote_state.networking.outputs.subnet_self_link
    network_ip = google_compute_address.cloudflared.address
    # No access_config — internal only, egress via Cloud NAT
  }

  service_account {
    email  = data.terraform_remote_state.secrets.outputs.cloudflared_service_account_email
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }

  metadata_startup_script = <<-EOT
    #!/bin/bash
    set -e

    # Install cloudflared from official Cloudflare .deb package
    curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg | gpg --dearmor -o /usr/share/keyrings/cloudflare-main.gpg
    echo "deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared $(lsb_release -cs) main" > /etc/apt/sources.list.d/cloudflared.list
    apt-get update && apt-get install -y cloudflared

    # Fetch tunnel token from GCP Secret Manager
    TOKEN=$(gcloud secrets versions access latest --secret="${data.terraform_remote_state.secrets.outputs.cloudflared_secret_id}")

    # Install and start cloudflared as a systemd service
    cloudflared service install $TOKEN
    systemctl enable cloudflared
    systemctl start cloudflared
  EOT

  tags   = ["cloudflared-tunnel"]
  labels = local.labels

  allow_stopping_for_update = true
}

# Allow egress to Cloudflare edge (HTTPS and QUIC)
resource "google_compute_firewall" "cloudflared_egress" {
  name      = "${var.project_name}-cloudflared-egress"
  network   = data.terraform_remote_state.networking.outputs.vpc_name
  direction = "EGRESS"

  allow {
    protocol = "tcp"
    ports    = ["443"]
  }

  allow {
    protocol = "udp"
    ports    = ["7844"]
  }

  target_tags        = ["cloudflared-tunnel"]
  destination_ranges = ["0.0.0.0/0"]
}
