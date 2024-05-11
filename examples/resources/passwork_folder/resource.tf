resource "passwork_vault" "example" {
  name       = "example-vault"
  is_private = true
}

resource "passwork_folder" "example" {
  name     = "example-folder"
  vault_id = passwork_vault.example.id
}

resource "passwork_folder" "example_nested" {
  name      = "nested-example-folder"
  vault_id  = passwork_vault.example.id
  parent_id = passwork_folder.example.id
}
