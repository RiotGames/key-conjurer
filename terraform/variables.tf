variable "account_number" {
  type        = string
  description = "Account number for AWS account"
}

variable "api_cert" {
  type        = string
  description = "ACM ARN for api"
}

variable "api_domain" {
  type        = string
  description = "Domain name to use for the api"
}

variable "frontend_cert" {
  type        = string
  description = "ACM ARN for frontend"
}

variable "frontend_domain" {
  type        = string
  description = "Domain name to use for the frontend"
}

variable "lambda_env" {
  type        = map(string)
  description = "Enviroment variables for lambdas"
}

variable "s3_tf_bucket" {
  type        = string
  description = "Bucket to store terraform information; it should be same as the one in your AWS provider"
}

variable "subnets" {
  type        = list(string)
  description = "Subnets to use for deployment"
}

variable "vpc_id" {
  type        = string
  description = "VPC to use for deployment"
}

variable "create_waf_acl" {
  type    = bool
  default = false
}

variable "waf_acl_id" {
  type        = string
  default     = ""
  description = "The ACL to use with the Cloudfront distribution that is created. if not specified, an ACL which blocks all public access will be created"
}

variable "kms_key_arn" {
  type        = string
  description = "The KMS encryption key that is used to encrypt and decrypt credentials so that they are not stored on the users drive in plaintext"
}

variable "lb_security_group_ids" {
  type = list(string)
}

variable frontend_bucket_name {
  type = string
}

variable lambda_execution_role_name {
  type = string
}

variable environment {
  type = string
}