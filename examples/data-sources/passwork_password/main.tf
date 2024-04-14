terraform {
  required_version = ">= 1.6"

  required_providers {
    passwork = {
      source = "registry.terraform.io/lupa95/passwork"
    }
  }
}

provider "passwork" {}

data "passwork_password" "example" {
  id       = var.password_id
  vault_id = var.vault_id
}
