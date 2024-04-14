terraform {
  required_version = ">= 1.6"

  required_providers {
    passwork = {
      source = "registry.terraform.io/lupa95/passwork"
    }
  }
}

provider "passwork" {}

resource "passwork_vault" "example" {
  name       = var.vault_name
  is_private = var.vault_is_private
}
