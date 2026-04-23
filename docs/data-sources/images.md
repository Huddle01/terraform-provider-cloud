---
page_title: "huddle_cloud_images Data Source - Huddle01 Cloud"
description: |-
  Lists all OS images available in a region, grouped by Linux distribution.
---

# huddle_cloud_images (Data Source)

Lists all OS images available in a region, grouped by Linux distribution.

## Example Usage

```hcl
data "huddle_cloud_images" "eu2" {
  region = "eu2"
}

# Print all available image groups and their versions
output "image_groups" {
  value = data.huddle_cloud_images.eu2.image_groups
}

# Find a specific Ubuntu image
locals {
  ubuntu_group = [
    for g in data.huddle_cloud_images.eu2.image_groups : g
    if g.distro == "ubuntu"
  ]

  ubuntu_22_04_id = (
    length(local.ubuntu_group) > 0
    ? [
        for v in local.ubuntu_group[0].versions : v.id
        if v.version == "22.04"
      ][0]
    : null
  )
}
```

## Schema

### Optional

- `region` (String) Region to query for available images. Defaults to the provider-level region.

### Read-Only

- `image_groups` (List of Object) OS images grouped by Linux distribution. Each object has the following attributes:
  - `distro` (String) Linux distribution name (e.g. `ubuntu`, `debian`, `centos`).
  - `versions` (List of Object) Available versions of this distribution. Each object has:
    - `id` (String) Unique image identifier (internal UUID).
    - `name` (String) Human-readable image name to use in `image_name` resource/module inputs (e.g. `ubuntu-22.04`).
    - `version` (String) Human-readable version string (e.g. `22.04`).
