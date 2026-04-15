# Provider examples

- `minimal-vm`: create keypair, security group, and VM
- `custom-security-rules`: create custom SG rule set
- `existing-network-instance`: create VM on an existing network

Set required variables (API key, flavor/image IDs, SSH key) with `.tfvars` or environment.

If `terraform init` reports that `registry.terraform.io/huddle01/cloud` cannot be found, follow the
local mirror setup in the provider README: [../README.md](../README.md#local-development-before-registry-publish).
