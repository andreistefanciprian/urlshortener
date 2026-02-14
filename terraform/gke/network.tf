
# Allow ssh into GKE nodes for debug purposes
# not recommended in production clusters
resource "google_compute_firewall" "allow_ssh" {
  name    = "${var.project_name}-allow-ssh"
  network = data.terraform_remote_state.networking.outputs.vpc_name

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  target_tags   = ["gke-node"]
  source_ranges = ["0.0.0.0/0"]
}

# Allow istio pilot to inject sidecars
# needed by the Pilot discovery validation webhook
resource "google_compute_firewall" "allow_istio_auto_inject" {
  name    = "${var.project_name}-allow-istio-auto-inject"
  network = data.terraform_remote_state.networking.outputs.vpc_name

  allow {
    protocol = "tcp"
    ports    = ["15017"]
  }

  source_ranges = [var.gke_master_cidr]

  target_tags = ["gke-node"]
}
