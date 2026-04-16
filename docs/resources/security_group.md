---
page_title: "huddle_cloud_security_group Resource - Huddle01 Cloud"
description: |-
  Manages a security group and reads its associated rules.
---

# huddle_cloud_security_group (Resource)

Manages a security group and reads its associated rules. To add rules beyond the defaults, use the [`huddle_cloud_security_group_rule`](security_group_rule.md) resource.

~> **Note:** The `rules` attribute on this resource is **read-only**. It reflects rules that the platform provisions automatically when the security group is created (e.g. default egress rules). To manage additional rules, use separate `huddle_cloud_security_group_rule` resources.

## Example Usage

```hcl
resource "huddle_cloud_security_group" "web" {
  name        = "web-sg"
  description = "Security group for web servers"
  region      = "eu2"
}

resource "huddle_cloud_security_group_rule" "http" {
  security_group_id = huddle_cloud_security_group.web.id
  direction         = "ingress"
  ether_type        = "IPv4"
  protocol          = "tcp"
  port_range_min    = 80
  port_range_max    = 80
  remote_ip_prefix  = "0.0.0.0/0"
  region            = "eu2"
}
```

## Schema

### Required

- `name` (String) Human-readable name for the security group.

### Optional

- `description` (String) Optional description for the security group.
- `region` (String) Region in which to create the security group. Defaults to the provider-level region.

### Read-Only

- `created_at` (String) Timestamp when the security group was created (RFC 3339).
- `id` (String) Unique identifier of the security group.
- `rules` (List of Object) Read-only list of rules currently attached to this security group. Each object has the following attributes:
  - `id` (String) Unique identifier of the rule.
  - `direction` (String) Traffic direction: `ingress` or `egress`.
  - `ether_type` (String) IP version: `IPv4` or `IPv6`.
  - `protocol` (String) IP protocol name (e.g. `tcp`, `udp`, `icmp`). Empty string means all protocols.
  - `port_range_min` (Number) Start of the port range (inclusive). `0` means not applicable.
  - `port_range_max` (Number) End of the port range (inclusive). `0` means not applicable.
  - `remote_ip_prefix` (String) CIDR block that the rule applies to.
  - `remote_group_id` (String) ID of a remote security group the rule applies to.
- `updated_at` (String) Timestamp when the security group was last updated (RFC 3339).

## Import

Import an existing security group by its ID:

```shell
terraform import huddle_cloud_security_group.example <security-group-id>
```
