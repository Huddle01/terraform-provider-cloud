# Volume lifecycle example

This example documents the recommended explicit lifecycle for data volumes:

1. Create an instance.
2. Create one or more standalone volumes.
3. Attach each volume using `huddle_cloud_volume_attachment`.
4. Detach by removing the attachment resource.
5. Delete the volume only after it is detached.

## Why this pattern

This keeps storage lifecycle independent from instance lifecycle and avoids conflicts with default instance disk settings.

`huddle_cloud_instance.additional_volume_size` has been removed. Use:

- `huddle_cloud_volume` for volume creation and management
- `huddle_cloud_volume_attachment` for attach/detach operations

## Volume deletion on destroy

`huddle_cloud_volume` defaults to `delete_on_destroy = false`. When this is set (or left at the default), `terraform destroy` removes the volume from Terraform state but **does not delete the actual volume** — your data is preserved.

Set `delete_on_destroy = true` only when you want Terraform to permanently delete the volume:

```hcl
resource "huddle_cloud_volume" "data" {
  name              = "vm-data"
  size              = 100
  region            = var.region
  delete_on_destroy = true  # omit or set false to retain volume on destroy
}
```

> **Warning:** `delete_on_destroy = true` permanently destroys the volume and all data on it. This cannot be undone.

## Minimal snippet

```hcl
resource "huddle_cloud_instance" "vm" {
  name                 = "vm-with-volume"
  region               = var.region
  flavor_id            = var.flavor_id
  image_id             = var.image_id
  key_names            = [huddle_cloud_keypair.example.name]
  security_group_names = ["default"]
}

resource "huddle_cloud_volume" "data" {
  name   = "vm-data"
  size   = 100
  region = var.region
  # delete_on_destroy = false  # default: volume is retained on terraform destroy
}

resource "huddle_cloud_volume_attachment" "data_to_vm" {
  volume_id   = huddle_cloud_volume.data.id
  instance_id = huddle_cloud_instance.vm.id
  region      = var.region
}
```
