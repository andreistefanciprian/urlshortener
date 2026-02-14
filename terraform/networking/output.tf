# VPC Outputs
output "vpc_name" {
  description = "The name of the VPC"
  value       = google_compute_network.vpc.name
}

output "vpc_id" {
  description = "The ID of the VPC"
  value       = google_compute_network.vpc.id
}

output "vpc_self_link" {
  description = "The self link of the VPC"
  value       = google_compute_network.vpc.self_link
}

# Subnet Outputs
output "subnet_name" {
  description = "The name of the subnet"
  value       = google_compute_subnetwork.subnet.name
}

output "subnet_id" {
  description = "The ID of the subnet"
  value       = google_compute_subnetwork.subnet.id
}

output "subnet_self_link" {
  description = "The self link of the subnet"
  value       = google_compute_subnetwork.subnet.self_link
}

output "subnet_ip_cidr_range" {
  description = "The IP CIDR range of the subnet"
  value       = google_compute_subnetwork.subnet.ip_cidr_range
}

output "subnet_region" {
  description = "The region of the subnet"
  value       = google_compute_subnetwork.subnet.region
}

# NAT Router Outputs
output "nat_router_name" {
  description = "The name of the Cloud Router"
  value       = google_compute_router.router.name
}

output "nat_router_id" {
  description = "The ID of the Cloud Router"
  value       = google_compute_router.router.id
}

output "nat_name" {
  description = "The name of the Cloud NAT"
  value       = google_compute_router_nat.nat.name
}
