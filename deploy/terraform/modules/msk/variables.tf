variable "environment" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "private_subnet_ids" {
  type = list(string)
}

variable "eks_security_group_id" {
  type = string
}

variable "broker_instance_type" {
  type = string
}

variable "broker_count" {
  type = number
}

variable "ebs_volume_size" {
  type = number
}

variable "alert_email" {
  type    = string
  default = ""
}
