# Random secret for tunnel authentication
resource "random_id" "tunnel_secret" {
  byte_length = 32
}

# Cloudflare Tunnel
resource "cloudflare_zero_trust_tunnel_cloudflared" "gke" {
  account_id = var.cloudflare_account_id
  name       = "${var.project_name}-gke-tunnel"
  secret     = random_id.tunnel_secret.b64_std
}

# Route private network CIDR through the tunnel
resource "cloudflare_zero_trust_tunnel_route" "gke" {
  account_id = var.cloudflare_account_id
  tunnel_id  = cloudflare_zero_trust_tunnel_cloudflared.gke.id
  network    = var.tunnel_route_cidr # 10.100.0.0/18
  comment    = "Route to GKE private network (nodes, pods, services, master)"
}

# WARP client split tunnel — route GKE private network through Cloudflare
resource "cloudflare_zero_trust_split_tunnel" "gke" {
  account_id = var.cloudflare_account_id
  mode       = "include"
  tunnels {
    address     = var.tunnel_route_cidr # 10.100.0.0/18
    description = "GKE private network (nodes, pods, services, master)"
  }
}

# Write tunnel token into the secret shell created by the secrets layer
resource "google_secret_manager_secret_version" "tunnel_token" {
  secret      = data.terraform_remote_state.secrets.outputs.cloudflared_secret_name
  secret_data = cloudflare_zero_trust_tunnel_cloudflared.gke.tunnel_token
}
