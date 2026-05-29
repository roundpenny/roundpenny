variable "environment" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "public_subnet_ids" {
  type = list(string)
}

variable "certificate_arn" {
  type = string
}

variable "kong_private_ips" {
  type        = list(string)
  description = "Private IPs of Kong pods for ALB target group"
  default     = []
}
