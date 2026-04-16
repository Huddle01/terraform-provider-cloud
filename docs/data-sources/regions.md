---
page_title: "huddle_cloud_regions Data Source - Huddle01 Cloud"
description: |-
  Lists all regions available to the authenticated workspace.
---

# huddle_cloud_regions (Data Source)

Lists all regions available to the authenticated workspace.

## Example Usage

```hcl
data "huddle_cloud_regions" "available" {}

output "regions" {
  value = data.huddle_cloud_regions.available.regions
}

# Use a specific region that is enabled
locals {
  enabled_regions = [
    for region, enabled in data.huddle_cloud_regions.available.regions : region
    if enabled
  ]
}
```

## Schema

### Read-Only

- `regions` (Map of Boolean) Map of region identifiers to a boolean indicating whether the region is enabled for the workspace.
