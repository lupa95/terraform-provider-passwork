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
  is_private = true
}

resource "passwork_folder" "example" {
  name     = var.folder_name
  vault_id = passwork_vault.example.id
}

resource "passwork_folder" "example_nested" {
  name      = "${var.folder_name}-nested"
  vault_id  = passwork_vault.example.id
  parent_id = passwork_folder.example.id
}
