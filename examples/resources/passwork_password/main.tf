resource "passwork_password" "example" {
  name        = var.name
  vault_id    = var.vault_id
  login       = var.login
  password    = var.password
  description = var.description
}
