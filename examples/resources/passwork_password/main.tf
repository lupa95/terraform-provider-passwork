terraform {
  required_version = ">= 1.6"

  required_providers {
    passwork = {
      source = "registry.terraform.io/lupa95/passwork"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.6.0"
    }
  }
}

provider "passwork" {}

resource "passwork_vault" "example" {
  name       = var.vault_name
  is_private = true
}

resource "random_password" "example" {
  length = 16
}

resource "passwork_password" "example" {
  name        = var.password_name
  vault_id    = passwork_vault.example.id
  login       = var.password_login
  url         = var.password_url
  description = var.password_description
  password    = random_password.example.result
}
