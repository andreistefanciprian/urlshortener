terraform {
  required_version = ">= 1.14.5"
}
provider "google" {
  project = var.gcp_project
  region  = var.gcp_region
}