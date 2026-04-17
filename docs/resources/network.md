---
page_title: "huddle_cloud_network Resource - Huddle01 Cloud"
description: |-
  Manages a private network and its primary subnet.
---

# huddle_cloud_network (Resource)

Manages a private network and its primary subnet.

## Example Usage

```hcl
resource "huddle_cloud_network" "example" {
  name                = "my-private-net"
  description         = "Private network for app servers"
  primary_subnet_cidr = "192.168.1.0/24"
  enable_dhcp         = true
  region              = "eu2"
}
```

## Schema

### Required

- `name` (String) Human-readable name for the network.

### Optional

- `description` (String) Optional description for the network.
- `enable_dhcp` (Boolean) Whether DHCP is enabled on the primary subnet. Defaults to `true`.
- `no_gateway` (Boolean) When `true`, no gateway is created for the subnet. Defaults to `false`.
- `pool_cidr` (String) CIDR block of the allocation pool for the subnet (e.g. `10.0.0.0/8`).
- `primary_subnet_cidr` (String) CIDR block for the primary subnet (e.g. `192.168.1.0/24`). Takes precedence over `primary_subnet_size`.
- `primary_subnet_size` (Number) Prefix length used to auto-allocate the primary subnet from `pool_cidr` (e.g. `24` allocates a `/24`).
- `region` (String) Region in which to create the network. Defaults to the provider-level region.

### Read-Only

- `admin_state_up` (Boolean) Whether the network is administratively up.
- `id` (String) Unique identifier of the network.
- `status` (String) Current status of the network (`ACTIVE`, `DOWN`, etc.).
- `subnets` (List of String) List of subnet IDs associated with this network.

## Import

Import an existing network by its ID:

```shell
terraform import huddle_cloud_network.example <network-id>
```
