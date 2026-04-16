variable "huddle_api_key" {
  type      = string
  sensitive = true
}

variable "region" {
  type    = string
  default = "eu2"
}

variable "instance_id" {
  type = string
}

variable "volume_name" {
  type    = string
  default = "tf-data-volume"
}

variable "volume_size" {
  type    = number
  default = 20
}

variable "volume_type" {
  type    = string
  default = ""
}
