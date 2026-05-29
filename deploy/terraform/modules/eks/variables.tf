variable "environment" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "private_subnet_ids" {
  type = list(string)
}

variable "cluster_version" {
  type = string
}

variable "node_instance_types" {
  type = list(string)
}

variable "min_nodes" {
  type = number
}

variable "max_nodes" {
  type = number
}

variable "desired_nodes" {
  type = number
}

variable "oidc_thumbprint" {
  type        = string
  description = "OIDC provider thumbprint (use '9e99a48a9960b14926bb7f3b02e22da2b0ab7280' for prod)"
  default     = "9e99a48a9960b14926bb7f3b02e22da2b0ab7280"
}
