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
  name       = "tf-vm-vol-key"
  public_key = var.ssh_public_key
}

resource "huddle_cloud_security_group" "example" {
  name   = "tf-vm-vol-sg"
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
  name                 = "tf-vm-with-volume"
  region               = var.region
  flavor_id            = var.flavor_id
  image_id             = var.image_id
  boot_disk_size       = 30
  key_names            = [huddle_cloud_keypair.example.name]
  security_group_names = [huddle_cloud_security_group.example.name]
  assign_public_ip     = true
}

resource "huddle_cloud_volume" "data" {
  name        = "tf-vm-data-volume"
  description = "Data disk for tf-vm-with-volume"
  size        = var.volume_size
  volume_type = var.volume_type
  region      = var.region

  # By default (delete_on_destroy = false) the volume is NOT deleted when
  # `terraform destroy` runs — it is only removed from Terraform state.
  # Set to true if you want Terraform to permanently delete the volume on destroy.
  delete_on_destroy = true
}

resource "huddle_cloud_volume_attachment" "data" {
  volume_id   = huddle_cloud_volume.data.id
  instance_id = huddle_cloud_instance.example.id
  region      = var.region
}

output "instance_id" {
  value = huddle_cloud_instance.example.id
}

output "volume_id" {
  value = huddle_cloud_volume.data.id
}

output "attachment_id" {
  value = huddle_cloud_volume_attachment.data.id
}
