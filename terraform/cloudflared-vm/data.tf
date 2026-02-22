data "terraform_remote_state" "networking" {
  backend = "gcs"
  config = {
    bucket = var.tfstate_bucket
    prefix = "tfstate/networking"
  }
}

data "terraform_remote_state" "secrets" {
  backend = "gcs"
  config = {
    bucket = var.tfstate_bucket
    prefix = "tfstate/secrets"
  }
}

data "google_compute_image" "ubuntu" {
  family  = "ubuntu-2204-lts"
  project = "ubuntu-os-cloud"
}
