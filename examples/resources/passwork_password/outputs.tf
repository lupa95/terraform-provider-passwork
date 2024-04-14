output "password_id" {
  value = resource.passwork_password.example.id
}

output "password_value" {
  value     = resource.passwork_password.example.password
  sensitive = true
}
