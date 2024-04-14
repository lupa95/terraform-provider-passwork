terraform {
  required_providers {
    passwork = {
      source = "registry.terraform.io/lupa95/passwork"
    }
  }
}

provider "passwork" {
  host    = "https://my-passwork-instance.com/api/v4" # Can also be passed by the environment variable PASSWORK_HOST
  api_key = "my-api-key" # Can also be passed by the environment variable PASSWORK_API_KEY
}
