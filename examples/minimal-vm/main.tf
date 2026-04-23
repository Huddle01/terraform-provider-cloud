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
  flavor_name          = var.flavor_name
  image_name           = var.image_name
  boot_disk_size       = 30
  key_names            = [huddle_cloud_keypair.example.name]
  security_group_names = [huddle_cloud_security_group.example.name]
  assign_public_ip     = true
}
