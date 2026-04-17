---
page_title: "huddle_cloud_instance Resource - Huddle01 Cloud"
description: |-
  Manages a Huddle01 Cloud virtual machine instance.
---

# huddle_cloud_instance (Resource)

Manages a Huddle01 Cloud virtual machine instance.

## Example Usage

```hcl
terraform {
  required_providers {
    huddle = {
      source = "huddle01/cloud"
    }
  }
}

provider "huddle" {
  api_key = var.huddle_api_key
  region  = var.region
}

resource "huddle_cloud_keypair" "example" {
  name       = "tf-minimal-key"
  public_key = var.ssh_public_key
}

resource "huddle_cloud_security_group" "example" {
  name   = "tf-minimal-sg"
  region = var.region
}

resource "huddle_cloud_security_group_rule" "ssh" {
  security_group_id = huddle_cloud_security_group.example.id
  direction         = "ingress"
  ether_type        = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  region            = var.region
}

resource "huddle_cloud_instance" "example" {
  name                 = "tf-minimal-vm"
  region               = var.region
  flavor_id            = var.flavor_id
  image_id             = var.image_id
  boot_disk_size       = 30
  key_names            = [huddle_cloud_keypair.example.name]
  security_group_names = [huddle_cloud_security_group.example.name]
  assign_public_ip     = true
}
```

## Schema

### Required

- `boot_disk_size` (Number) Size of the boot disk in GB. Defaults to `30`.
- `flavor_id` (String) ID of the flavor (hardware profile) to use. Use the `huddle_cloud_flavors` data source to list available flavors.
- `image_id` (String) ID of the OS image to boot from. Use the `huddle_cloud_images` data source to list available images.
- `key_names` (List of String) List of keypair names to inject into the instance for SSH access.
- `name` (String) Human-readable name for the instance.
- `security_group_names` (List of String) List of security group names to attach to the instance.

### Optional

- `assign_public_ip` (Boolean) Whether to assign a public IPv4 address. Defaults to `true`.
- `network_id` (String) ID of the network to attach the instance to. If omitted, the workspace default network is used.
- `power_state` (String) Desired power state of the instance. One of `active`, `stopped`, `paused`, `suspended`. Defaults to `active`.
- `region` (String) Region in which to create the instance. Defaults to the provider-level region.
- `tags` (List of String) Map of key/value tags to apply to the instance.

### Read-Only

- `created_at` (String) Timestamp when the instance was created (RFC 3339).
- `id` (String) Unique identifier of the instance.
- `private_ipv4` (String) Private IPv4 address of the instance.
- `public_ipv4` (String) Public IPv4 address of the instance, if assigned.
- `ram` (Number) Amount of RAM allocated to the instance in MB.
- `status` (String) Current status of the instance as reported by the API.
- `vcpus` (Number) Number of virtual CPUs allocated to the instance.

## Import

Import an existing instance by its ID:

```shell
terraform import huddle_cloud_instance.example <instance-id>
```

If the provider `region` is not set, you may also need to set the `region` attribute in the imported resource block.
