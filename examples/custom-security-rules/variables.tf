variable "huddle_api_key" {
  type      = string
  sensitive = true
}

variable "region" {
  type    = string
  default = "eu2"
}

variable "rules" {
  type = list(object({
    protocol = string
    port     = number
    cidr     = string
  }))
  default = [
    { protocol = "tcp", port = 22, cidr = "0.0.0.0/0" },
    { protocol = "tcp", port = 80, cidr = "0.0.0.0/0" },
    { protocol = "tcp", port = 443, cidr = "0.0.0.0/0" }
  ]
}
