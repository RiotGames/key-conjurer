variable "listener_arn" {
  type = string
}

variable "subnets" {
  type = list(string)
}

variable "security_group_ids" {
  type = list(string)
}

variable "bucket_name" {
  type = string
}

variable "environment" {
  type = string
}

variable "environment_variables" {
  type    = map(string)
  default = {}
}

variable "execution_role_arn" {
  type = string
}
