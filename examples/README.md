# Provider examples

- `minimal-vm`: create keypair, security group, and VM
- `custom-security-rules`: create custom SG rule set
- `existing-network-instance`: create VM on an existing network
- `vm-with-volume`: create VM and attach an additional standalone volume
- `volume-lifecycle`: create a standalone volume and attach/detach it from an instance

## Volume lifecycle notes

Use explicit volume resources for data disks:

- `huddle_cloud_volume` manages disk lifecycle
- `huddle_cloud_volume_attachment` manages attach/detach lifecycle

`huddle_cloud_instance.additional_volume_size` has been removed. Use explicit volume + attachment resources.

Set required variables (API key, flavor/image IDs, SSH key) with `.tfvars` or environment.

If `terraform init` reports that `registry.terraform.io/huddle01/cloud` cannot be found, follow the
local mirror setup in the provider README: [../README.md](../README.md#local-development-before-registry-publish).
