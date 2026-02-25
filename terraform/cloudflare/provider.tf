terraform {
  required_version = ">= 1.14.5"
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 7.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}
provider "cloudflare" {
  api_token = var.cloudflare_api_token
}
provider "google" {
  project = var.gcp_project
  region  = var.gcp_region
}
