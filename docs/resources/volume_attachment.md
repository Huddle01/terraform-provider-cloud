---
page_title: "huddle_cloud_volume_attachment Resource - Huddle01 Cloud"
description: |-
  Attaches a block storage volume to a virtual machine instance.
---

# huddle_cloud_volume_attachment (Resource)

Attaches a block storage volume to a virtual machine instance.

## Example Usage

```hcl
resource "huddle_cloud_volume" "data" {
  name   = "my-data-volume"
  size   = 50
  region = "eu2"
}

resource "huddle_cloud_instance" "app" {
  name                 = "my-app-server"
  region               = "eu2"
  flavor_name          = var.flavor_name
  image_name           = var.image_name
  boot_disk_size       = 30
  key_names            = [huddle_cloud_keypair.example.name]
  security_group_names = [huddle_cloud_security_group.example.name]
}

resource "huddle_cloud_volume_attachment" "data" {
  volume_id   = huddle_cloud_volume.data.id
  instance_id = huddle_cloud_instance.app.id
  region      = "eu2"
}

output "device_path" {
  value = huddle_cloud_volume_attachment.data.device
}
```

## Schema

### Required

- `instance_id` (String) ID of the instance to attach the volume to.
- `volume_id` (String) ID of the volume to attach.

### Optional

- `region` (String) Region of the instance and volume. Defaults to the provider-level region.

### Read-Only

- `device` (String) Device path on the instance where the volume is exposed (e.g. `/dev/vdb`). Assigned by the API.
- `id` (String) Unique identifier of the attachment.

## Import

Import an existing attachment using `<volume_id>/<instance_id>`:

```shell
terraform import huddle_cloud_volume_attachment.example <volume-id>/<instance-id>
```
