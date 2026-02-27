# define GCP region
variable "gcp_region" {
  type        = string
  description = "GCP region"
}
# define GCP project name
variable "gcp_project" {
  type        = string
  description = "GCP project name"
}

variable "tfstate_bucket" {
  type        = string
  description = "GCS bucket name for Terraform state"
}

variable "project_name" {
  type        = string
  description = "Name prefix used across all resources for consistent naming"
  default     = "home"
}

variable "cloudflare_account_id" {
  type        = string
  description = "Cloudflare account ID"
  sensitive   = true
}

variable "cloudflare_api_token" {
  type        = string
  description = "Cloudflare API token"
  sensitive   = true
}

variable "cloudflare_zone_id" {
  type        = string
  description = "Cloudflare zone ID for the domain (e.g. example.com)"
  sensitive   = true
}

variable "cloudflare_domain" {
  type        = string
  description = "Cloudflare domain name (e.g. example.com)"
}

variable "ingress_lb_ip" {
  type        = string
  description = "Internal Load Balancer IP of the k8s ingress controller"
  default     = "10.100.0.8"
}

variable "tunnel_route_cidr" {
  type        = string
  description = "CIDR range to route through the Cloudflare tunnel (covers nodes, pods, services, and GKE master)"
  default     = "10.100.0.0/18"
}
