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
