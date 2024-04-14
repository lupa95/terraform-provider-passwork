output "password_value" {
  value     = data.passwork_password.example.password
  sensitive = true
}