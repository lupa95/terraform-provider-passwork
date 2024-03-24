output "password" {
  value     = data.passwork_password.example.password
  sensitive = true
}