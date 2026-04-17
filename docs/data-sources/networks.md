---
page_title: "huddle_cloud_networks Data Source - Huddle01 Cloud"
description: |-
  Lists networks accessible to the workspace in a region.
---

# huddle_cloud_networks (Data Source)

Lists networks accessible to the workspace in a region.

## Example Usage

```hcl
# List all networks (including shared ones) in eu2
data "huddle_cloud_networks" "eu2" {
  region = "eu2"
}

# Look up only workspace-owned networks
data "huddle_cloud_networks" "owned" {
  region = "eu2"
  owned  = true
}

# Use the first available network when creating an instance
locals {
  network_id = data.huddle_cloud_networks.eu2.networks[0].id
}

resource "huddle_cloud_instance" "example" {
  name                 = "my-vm"
  region               = "eu2"
  flavor_id            = var.flavor_id
  image_id             = var.image_id
  boot_disk_size       = 30
  network_id           = local.network_id
  key_names            = [huddle_cloud_keypair.example.name]
  security_group_names = [huddle_cloud_security_group.example.name]
}
```

## Schema

### Optional

- `owned` (Boolean) When `true`, only returns networks owned by the workspace. When `false` (default), shared networks are included.
- `region` (String) Region to query. Defaults to the provider-level region.

### Read-Only

- `networks` (List of Object) List of accessible networks. Each object has the following attributes:
  - `admin_state_up` (Boolean) Whether the network is administratively up.
  - `id` (String) Unique identifier of the network.
  - `name` (String) Name of the network.
  - `status` (String) Current network status (`ACTIVE`, `DOWN`, etc.).
  - `subnets` (List of String) List of subnet IDs associated with the network.
