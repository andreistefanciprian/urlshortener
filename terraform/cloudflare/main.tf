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

# WARP local domain fallback — enrolled clients resolve home.internal via the GCP VPC DNS resolver
# so they can reach internal services by name (api.home.internal, app.home.internal, etc.)
resource "cloudflare_zero_trust_device_default_profile_local_domain_fallback" "private" {
  account_id = var.cloudflare_account_id
  domains = [
    {
      suffix      = "home.internal"
      description = "GKE private DNS zone"
      dns_server  = [var.vpc_dns_resolver] # GCP VPC resolver: subnet base + 2
    },
  ]
}

# Public DNS records — Terraform-managed, static CNAMEs to the tunnel
# (not managed by external-dns, which is pointed at the private zone)
resource "cloudflare_dns_record" "api_gateway" {
  zone_id = var.cloudflare_zone_id
  name    = "@" # apex: 9tzy.xyz
  type    = "CNAME"
  content = "${cloudflare_zero_trust_tunnel_cloudflared.gke.id}.cfargotunnel.com"
  proxied = true
  ttl     = 1 # Auto TTL (required when proxied)
}

resource "cloudflare_dns_record" "frontend" {
  zone_id = var.cloudflare_zone_id
  name    = "app" # app.9tzy.xyz
  type    = "CNAME"
  content = "${cloudflare_zero_trust_tunnel_cloudflared.gke.id}.cfargotunnel.com"
  proxied = true
  ttl     = 1 # Auto TTL (required when proxied)
}

# Tunnel ingress — tell cloudflared where to forward each public hostname
resource "cloudflare_zero_trust_tunnel_cloudflared_config" "gke" {
  account_id = var.cloudflare_account_id
  tunnel_id  = cloudflare_zero_trust_tunnel_cloudflared.gke.id

  config = {
    ingress = [
      {
        hostname = var.cloudflare_domain # 9tzy.xyz → api-gateway
        service  = "http://gateway.home.internal"
      },
      {
        hostname = "app.${var.cloudflare_domain}" # app.9tzy.xyz → frontend
        service  = "http://gateway.home.internal"
      },
      {
        service = "http_status:404" # catch-all (required by cloudflared)
      },
    ]
  }
}

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
