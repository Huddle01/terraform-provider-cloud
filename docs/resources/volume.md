---
page_title: "huddle_cloud_volume Resource - Huddle01 Cloud"
description: |-
  Manages a block storage volume.
---

# huddle_cloud_volume (Resource)

Manages a block storage volume.

-> **Important:** By default, `terraform destroy` removes the volume from Terraform state but **does not delete the underlying volume** in the cloud. Set `delete_on_destroy = true` if you want Terraform to permanently delete the volume on destroy.

## Example Usage

```hcl
resource "huddle_cloud_instance" "example" {
  name                 = "tf-vm-with-volume"
  region               = var.region
  flavor_id            = var.flavor_id
  image_id             = var.image_id
  boot_disk_size       = 30
  key_names            = [huddle_cloud_keypair.example.name]
  security_group_names = [huddle_cloud_security_group.example.name]
  assign_public_ip     = true
}

resource "huddle_cloud_volume" "data" {
  name        = "tf-vm-data-volume"
  description = "Data disk for tf-vm-with-volume"
  size        = var.volume_size
  volume_type = var.volume_type
  region      = var.region

  # By default (delete_on_destroy = false) the volume is NOT deleted when
  # `terraform destroy` runs — it is only removed from Terraform state.
  # Set to true if you want Terraform to permanently delete the volume on destroy.
  delete_on_destroy = true
}

resource "huddle_cloud_volume_attachment" "data" {
  volume_id   = huddle_cloud_volume.data.id
  instance_id = huddle_cloud_instance.example.id
  region      = var.region
}
```

## Schema

### Required

- `name` (String) Human-readable name for the volume.
- `size` (Number) Size of the volume in GB.

### Optional

- `delete_on_destroy` (Boolean) When `true`, the volume is deleted from the cloud when the resource is destroyed. When `false` (default), `terraform destroy` removes the resource from state but **leaves the volume intact** in the cloud. Set this to `true` only if you want Terraform to permanently delete the volume on destroy.
- `description` (String) Optional description for the volume.
- `region` (String) Region in which to create the volume. Defaults to the provider-level region.
- `volume_type` (String) Storage backend type. Defaults to `standard`.

### Read-Only

- `attachments` (List of Object) List of instances this volume is currently attached to. Each object has the following attributes:
  - `device` (String) Device path on the instance (e.g. `/dev/vdb`).
  - `server_id` (String) ID of the instance the volume is attached to.
- `bootable` (Boolean) Whether the volume is bootable.
- `created_at` (String) Timestamp when the volume was created (RFC 3339).
- `id` (String) Unique identifier of the volume.
- `status` (String) Current status of the volume (`available`, `in-use`, `error`, etc.).
- `updated_at` (String) Timestamp when the volume was last updated (RFC 3339).

## Import

Import an existing volume by its ID:

```shell
terraform import huddle_cloud_volume.example <volume-id>
```

Note: when importing, `delete_on_destroy` defaults to `false`. Add the attribute explicitly to your configuration if you want destroy to delete the volume.
