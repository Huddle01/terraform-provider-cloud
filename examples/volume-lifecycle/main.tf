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

resource "huddle_cloud_volume" "data" {
  name        = var.volume_name
  description = "Terraform managed data volume"
  size        = var.volume_size
  volume_type = var.volume_type
  region      = var.region
}

resource "huddle_cloud_volume_attachment" "data" {
  volume_id   = huddle_cloud_volume.data.id
  instance_id = var.instance_id
  region      = var.region
}

output "volume_id" {
  value = huddle_cloud_volume.data.id
}

output "attachment_id" {
  value = huddle_cloud_volume_attachment.data.id
}
