data "terraform_remote_state" "secrets" {
  backend = "gcs"
  config = {
    bucket = var.tfstate_bucket
    prefix = "tfstate/secrets"
  }
}
