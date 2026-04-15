variable "huddle_api_key" {
  type      = string
  sensitive = true
}

variable "region" {
  type    = string
  default = "eu2"
}

variable "flavor_id" {
  type = string
}

variable "image_id" {
  type = string
}

variable "ssh_public_key" {
  type = string
}
