data "terraform_remote_state" "networking" {
  backend = "gcs"
  config = {
    bucket = var.tfstate_bucket
    prefix = "tfstate/networking"
  }
}

data "google_project" "project" {
  project_id = var.gcp_project
}