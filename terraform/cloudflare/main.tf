# Random secret for tunnel authentication
resource "random_id" "tunnel_secret" {
  byte_length = 32
}

# Cloudflare Tunnel
resource "cloudflare_zero_trust_tunnel_cloudflared" "gke" {
  account_id    = var.cloudflare_account_id
  name          = "${var.project_name}-gke-tunnel"
  tunnel_secret = random_id.tunnel_secret.b64_std
}

# Route private network CIDR through the tunnel
resource "cloudflare_zero_trust_tunnel_cloudflared_route" "gke" {
  account_id = var.cloudflare_account_id
  tunnel_id  = cloudflare_zero_trust_tunnel_cloudflared.gke.id
  network    = var.tunnel_route_cidr # 10.100.0.0/18
  comment    = "Route to GKE private network (nodes, pods, services, master)"
}

# WARP client split tunnel — route GKE private network through Cloudflare
resource "cloudflare_zero_trust_device_default_profile" "gke" {
  account_id = var.cloudflare_account_id

  # Proxy both Traffic and DNS through WARP
  service_mode_v2 = {
    mode = "warp"
  }

  include = [
    {
      address     = var.tunnel_route_cidr # 10.100.0.0/18
      description = "GKE private network (nodes, pods, services, master)"
    },
  ]
}

# # Expose internal K8s ingress via Cloudflare Tunnel and DNS
# # Public DNS record — route foo.example.com through the tunnel
# resource "cloudflare_dns_record" "foo" {
#   zone_id = var.cloudflare_zone_id
#   name    = "foo"
#   type    = "CNAME"
#   content = "${cloudflare_zero_trust_tunnel_cloudflared.gke.id}.cfargotunnel.com"
#   proxied = true
#   ttl     = 1 # Auto TTL (required when proxied)
# }

# # Tunnel ingress — tell cloudflared where to forward public hostname traffic
# resource "cloudflare_zero_trust_tunnel_cloudflared_config" "gke" {
#   account_id = var.cloudflare_account_id
#   tunnel_id  = cloudflare_zero_trust_tunnel_cloudflared.gke.id

#   config = {
#     ingress = [
#       {
#         hostname = "foo.${var.cloudflare_domain}"
#         path     = "^/webhook"
#         service  = "http://${var.ingress_lb_ip}"
#       },
#       {
#         hostname = "foo.${var.cloudflare_domain}"
#         service  = "http_status:404"
#       },
#       {
#         service = "http_status:404"
#       },
#     ]
#   }
# }

# Fetch tunnel token via data source (tunnel_token attribute removed in v5)
data "cloudflare_zero_trust_tunnel_cloudflared_token" "gke" {
  account_id = var.cloudflare_account_id
  tunnel_id  = cloudflare_zero_trust_tunnel_cloudflared.gke.id
}

# Write tunnel token into the secret shell created by the secrets layer
resource "google_secret_manager_secret_version" "tunnel_token" {
  secret      = data.terraform_remote_state.secrets.outputs.cloudflared_secret_name
  secret_data = data.cloudflare_zero_trust_tunnel_cloudflared_token.gke.token
}
