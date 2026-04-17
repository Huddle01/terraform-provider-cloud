---
page_title: "huddle_cloud_instance Data Source - Huddle01 Cloud"
description: |-
  Fetches details of an existing virtual machine instance by ID.
---

# huddle_cloud_instance (Data Source)

Fetches details of an existing virtual machine instance by ID.

## Example Usage

```hcl
data "huddle_cloud_instance" "existing" {
  id     = "abc123de-f456-7890-abcd-ef1234567890"
  region = "eu2"
}

output "instance_ip" {
  value = data.huddle_cloud_instance.existing.public_ipv4
}

output "instance_status" {
  value = data.huddle_cloud_instance.existing.status
}
```

## Schema

### Required

- `id` (String) ID of the instance to look up.

### Optional

- `region` (String) Region where the instance lives. Defaults to the provider-level region.

### Read-Only

- `created_at` (String) Timestamp when the instance was created (RFC 3339).
- `name` (String) Name of the instance.
- `private_ipv4` (String) Private IPv4 address.
- `public_ipv4` (String) Public IPv4 address, if assigned.
- `ram` (Number) RAM in MB.
- `status` (String) Current status of the instance.
- `vcpus` (Number) Number of virtual CPUs.
