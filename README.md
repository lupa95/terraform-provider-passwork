# Terraform Provider Passwork
This Terraform provider for the Password Manager (Passwork)[https://passwork.de/] can manage objects like passwords, vaults and folders.

The provider is compatible with Passwork API version `4.0.0`

## Requirements

- Terraform 1.x
- Go 1.21 (for development)

## Development

If you want to develop the provider, the following steps can be done to set up a local development environment.

1. Clone the repository

```bash
git clone git@github.com:lupa95/terraform-provider-passwork
```

2. Make changes and compile the provider:

```bash
go install
```

3. Create the file `~/.terraform.rc` and point to the local sources (GOPATH):

```bash
provider_installation {

  dev_overrides {
      "registry.terraform.io/lupa95/passwork" = "[insert GOPATH]/bin"
  }
  direct {}
}
```

## Run Tests

Running tests requires access to a Passwork instance.

1. Setup Provider configuration: 
```bash
export PASSWORK_API_KEY=<replace-with-api-key>
export PASSWORK_HOST=https://my-passwork-instance.com
export PASSWORK_VAULT_ID=<replace with ID of existing Vault> # Required for data source testing
```

2. Run tests:

```bash
# Run all tests
TF_ACC=1 go test -v

# Run tests for specific resource, e.g. password resource only:
TF_ACC=1 go test -v -run TestPasswordResource
```
