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
