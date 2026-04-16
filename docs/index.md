---
page_title: "Provider: Huddle01 Cloud"
description: |-
  Use the Huddle01 Cloud provider to manage cloud infrastructure resources — virtual machines, networks, security groups, keypairs, and volumes.
---

# Huddle01 Cloud Provider

The **Huddle01 Cloud** provider lets you manage infrastructure on [Huddle01 Cloud](https://cloud.huddle01.com) using Terraform. It supports virtual machines, block storage, private networks, security groups, and SSH keypairs.

## Authentication

Authenticate using an API key from the [Huddle01 Cloud dashboard](https://cloud.huddle01.com).

The API key can be supplied via:
- The `api_key` argument in the provider block
- The `HUDDLE_API_KEY` environment variable (recommended for CI/CD)

## Example Usage

```hcl
terraform {
  required_providers {
    huddle_cloud = {
      source  = "huddle01/cloud"
      version = "~> 0.1"
    }
  }
}

provider "huddle_cloud" {
  # api_key and region can also be set via HUDDLE_API_KEY / HUDDLE_REGION env vars
  api_key = var.huddle_api_key
  region  = "eu2"
}
```

## Schema

### Optional

- `api_key` (String, Sensitive) Huddle01 Cloud API key used to authenticate all requests. Can also be set via the `HUDDLE_API_KEY` environment variable.
- `region` (String) Default region for all resource operations (e.g. `eu2`). Can also be set via the `HUDDLE_REGION` environment variable.
- `base_url` (String) Base URL of the Huddle01 Cloud API. Defaults to `https://cloud.huddleapis.com/api/v1`.
- `request_timeout_seconds` (Number) HTTP request timeout in seconds. Defaults to `60`.
