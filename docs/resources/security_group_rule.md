---
page_title: "huddle_cloud_security_group_rule Resource - Huddle01 Cloud"
description: |-
  Adds an ingress or egress rule to an existing security group.
---

# huddle_cloud_security_group_rule (Resource)

Adds an ingress or egress rule to an existing security group.

## Example Usage

```hcl
resource "huddle_cloud_security_group" "web" {
  name   = "tf-web-sg"
  region = var.region
}

resource "huddle_cloud_security_group_rule" "rules" {
  for_each          = { for r in var.rules : "${r.protocol}-${r.port}" => r }
  security_group_id = huddle_cloud_security_group.web.id
  direction         = "ingress"
  ether_type        = "IPv4"
  protocol          = each.value.protocol
  port_range_min    = each.value.port
  port_range_max    = each.value.port
  remote_ip_prefix  = each.value.cidr
  region            = var.region
}
```

## Schema

### Required

- `direction` (String) Traffic direction: `ingress` or `egress`.
- `ether_type` (String) IP version the rule applies to: `IPv4` or `IPv6`. Defaults to `IPv4`.
- `security_group_id` (String) ID of the security group to add this rule to.

### Optional

- `port_range_max` (Number) End of the permitted port range (inclusive).
- `port_range_min` (Number) Start of the permitted port range (inclusive).
- `protocol` (String) IP protocol for the rule (e.g. `tcp`, `udp`, `icmp`). Omit to match all protocols.
- `region` (String) Region of the security group. Defaults to the provider-level region.
- `remote_group_id` (String) ID of a remote security group that this rule permits traffic from/to.
- `remote_ip_prefix` (String) CIDR block that this rule permits traffic from/to.

### Read-Only

- `created_at` (String) Timestamp when the rule was created (RFC 3339).
- `id` (String) Unique identifier of the security group rule.

## Import

Import an existing rule using `<security_group_id>/<rule_id>`:

```shell
terraform import huddle_cloud_security_group_rule.example <security-group-id>/<rule-id>
```
