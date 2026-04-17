# VPC
resource "google_compute_network" "vpc" {
  name                    = "${var.project_name}-vpc"
  auto_create_subnetworks = "false"
}

# Subnet
resource "google_compute_subnetwork" "subnet" {
  name          = "${var.project_name}-subnet"
  region        = var.gcp_region
  network       = google_compute_network.vpc.self_link
  ip_cidr_range = "10.100.0.0/24" # Nodes

  secondary_ip_range {
    range_name    = "gke-pods"
    ip_cidr_range = "10.100.16.0/20" # 4096 pod IPs
  }

  secondary_ip_range {
    range_name    = "gke-services"
    ip_cidr_range = "10.100.32.0/24" # 256 service IPs
  }
}


# Allow internet connectivity from inside the cluster so we can pull images from public registries
# Not recommended in production clusters
resource "google_compute_router" "router" {
  name    = "${var.project_name}-nat-router"
  region  = google_compute_subnetwork.subnet.region
  network = google_compute_network.vpc.id

  bgp {
    asn = 64514
  }
}

resource "google_compute_router_nat" "nat" {
  name                               = "${var.project_name}-nat"
  router                             = google_compute_router.router.name
  region                             = google_compute_router.router.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
  depends_on = [google_compute_router.router]
}

# Static internal IP for the GKE Gateway API (Traefik) internal load balancer
# Used by HTTPRoute/Gateway resources via the traefik-gateway namespace
resource "google_compute_address" "gateway_lb" {
  name         = "${var.project_name}-gateway-lb"
  region       = var.gcp_region
  subnetwork   = google_compute_subnetwork.subnet.self_link
  address_type = "INTERNAL"
}

# Static internal IP for the standard Traefik ingress controller
# Used by Ingress resources (e.g. flux-operator)
resource "google_compute_address" "ingress_lb" {
  name         = "${var.project_name}-ingress-lb"
  region       = var.gcp_region
  subnetwork   = google_compute_subnetwork.subnet.self_link
  address_type = "INTERNAL"
}

# Enable Cloud DNS API
resource "google_project_service" "dns" {
  service            = "dns.googleapis.com"
  disable_on_destroy = false
}

# Private DNS zone — resolvable only within this VPC
# Used by GKE pods, nodes, and the cloudflared VM for internal service discovery
resource "google_dns_managed_zone" "private" {
  name        = "${var.project_name}-private"
  dns_name    = "home.internal."
  description = "Private DNS zone for GKE workloads and internal GCP resources"
  visibility  = "private"

  private_visibility_config {
    networks {
      network_url = google_compute_network.vpc.self_link
    }
  }

  depends_on = [google_project_service.dns]
}

# DNS A records for internal service hostnames — all resolve to the Gateway LB
# Static records ensure cloudflared can resolve these before external-dns is running
resource "google_dns_record_set" "gateway" {
  name         = "gateway.home.internal."
  managed_zone = google_dns_managed_zone.private.name
  type         = "A"
  ttl          = 300
  rrdatas      = [google_compute_address.gateway_lb.address]
}

resource "google_dns_record_set" "api" {
  name         = "api.home.internal."
  managed_zone = google_dns_managed_zone.private.name
  type         = "A"
  ttl          = 300
  rrdatas      = [google_compute_address.gateway_lb.address]
}

resource "google_dns_record_set" "frontend" {
  name         = "frontend.home.internal."
  managed_zone = google_dns_managed_zone.private.name
  type         = "A"
  ttl          = 300
  rrdatas      = [google_compute_address.gateway_lb.address]
}
