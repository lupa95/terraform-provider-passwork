resource "passwork_vault" "example" {
  name = "example-vault"
}

resource "random_password" "example" {
  length = 16
}

resource "passwork_password" "example" {
  name        = "example-password"
  vault_id    = passwork_vault.example.id
  login       = "example-username"
  url         = "https://example.com"
  description = "These are example credentials."
  password    = random_password.example.result
}
