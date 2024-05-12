terraform {
  required_providers {
    passwork = {
      source = "registry.terraform.io/lupa95/passwork"
    }
  }
}

provider "passwork" {
  host    = "https://my-passwork-instance.com" # Can be sourced from the environment variable PASSWORK_HOST
  api_key = "my-api-key"                              # Can be sourced from the environment variable PASSWORK_API_KEY
}

resource "passwork_vault" "example" {
  name = "example-vault"
}

resource "passwork_password" "example" {
  name     = "example-password"
  vault_id = passwork_vault.example.id
  password = "my-secret-password"
}
