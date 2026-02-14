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

variable "domain_name" {
  type        = string
  description = "Primary domain name for certificates"
  default     = "netl1.com"
}

variable "domain_organization" {
  type        = string
  description = "Organization name for certificate authority"
  default     = "netl1"
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
