# Terraform Provider: Huddle01 Cloud

This provider manages Huddle01 Cloud IaaS resources via the public API.

## Provider source

```hcl
terraform {
  required_providers {
    huddle = {
      source  = "huddle01/cloud"
      version = ">= 0.1.0"
    }
  }
}
```

## Authentication

```hcl
provider "huddle" {
  api_key  = var.huddle_api_key
  region   = "eu2"
  base_url = "https://cloud.huddleapis.com/api/v1"
}
```

Environment variables:

- `HUDDLE_API_KEY`
- `HUDDLE_REGION`

## Resources

- `huddle_cloud_network`
- `huddle_cloud_security_group`
- `huddle_cloud_security_group_rule`
- `huddle_cloud_keypair`
- `huddle_cloud_instance`

## Data sources

- `huddle_cloud_regions`
- `huddle_cloud_flavors`
- `huddle_cloud_images`
- `huddle_cloud_networks`
- `huddle_cloud_instance`

## Examples

- `examples/minimal-vm`
- `examples/custom-security-rules`
- `examples/existing-network-instance`

## Local development (before registry publish)

If `terraform init` fails with:

`registry.terraform.io does not have a provider named registry.terraform.io/huddle01/cloud`

the provider is not published in the public Terraform Registry yet. Use a local filesystem mirror:

1. Install locally from `cloud/terraform/provider`:

```bash
make install-local
```

Use another version label if needed:

```bash
make install-local VERSION=0.1.1
```

2. Add provider installation config to `~/.terraformrc` (or a file referenced by `TF_CLI_CONFIG_FILE`):

```hcl
provider_installation {
  filesystem_mirror {
    path    = "$HOME/.terraform.d/plugins"
    include = ["huddle01/cloud"]
  }
  direct {
    exclude = ["huddle01/cloud"]
  }
}
```

3. Run `terraform init` in an example directory again.
