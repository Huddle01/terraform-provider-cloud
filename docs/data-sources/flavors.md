---
page_title: "huddle_cloud_flavors Data Source - Huddle01 Cloud"
description: |-
  Lists all compute flavors (hardware profiles) available in a region.
---

# huddle_cloud_flavors (Data Source)

Lists all compute flavors (hardware profiles) available in a region.

## Example Usage

```hcl
data "huddle_cloud_flavors" "eu2" {
  region = "eu2"
}

# Pick the first available flavor
locals {
  flavor_id = data.huddle_cloud_flavors.eu2.flavors[0].id
}

resource "huddle_cloud_instance" "example" {
  name                 = "my-vm"
  region               = "eu2"
  flavor_id            = local.flavor_id
  image_id             = var.image_id
  boot_disk_size       = 30
  key_names            = [huddle_cloud_keypair.example.name]
  security_group_names = [huddle_cloud_security_group.example.name]
}

output "available_flavors" {
  value = data.huddle_cloud_flavors.eu2.flavors
}
```

## Schema

### Optional

- `region` (String) Region to query for available flavors. Defaults to the provider-level region.

### Read-Only

- `flavors` (List of Object) List of available flavors. Each object has the following attributes:
  - `disk` (Number) Root disk size in GB.
  - `id` (String) Unique identifier of the flavor.
  - `name` (String) Human-readable name of the flavor (e.g. `standard-4`).
  - `price_per_hour` (Number) Hourly price in USD.
  - `ram` (Number) RAM in MB.
  - `vcpus` (Number) Number of virtual CPUs.
