---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "passwork_vault Resource - terraform-provider-passwork"
subcategory: ""
description: |-
  Vault resource
---

# passwork_vault (Resource)

Vault resource



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of Vault

### Optional

- `is_private` (Boolean) Create a private vault.
- `master_password` (String) Master password of the Vault

### Read-Only

- `access` (String) Access of the Vault
- `id` (String) ID of the Vault
- `scope` (String) Scope of the Vault