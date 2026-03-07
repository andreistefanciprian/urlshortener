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

variable "maintenance_window" {
  description = "Time window specified for daily maintenance operations to START in RFC3339 format"
  type        = string
  default     = "05:00"
}

variable "node_type" {
  type    = string
  default = "n1-standard-2"
}

variable "node_disk_type" {
  type    = string
  default = "pd-standard"
}

variable "node_disk_size" {
  type    = number
  default = 10
}

variable "project_name" {
  type        = string
  description = "Name prefix used across all resources for consistent naming"
  default     = "home"
}

variable "labels" {
  type        = map(string)
  description = "Additional labels to apply to all resources that support them"
  default     = {}
}

variable "gke_master_cidr" {
  type        = string
  description = "Private IP subnet for GKE control plane"
  default     = "10.100.48.0/28"
}

variable "use_default_node_pool" {
  type        = bool
  description = "Use the default node pool instead of a separately managed node pool. Note: autoscaling is not supported for the default node pool."
  default     = false
}
