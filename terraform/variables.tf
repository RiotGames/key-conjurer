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

variable "region" {
  type        = string
  description = "AWS region"
}

variable "s3_tf_bucket" {
  type        = string
  description = "Bucket to store terraform information; it should be same as the one in your AWS provider"
}

variable "subnets" {
  type        = list(string)
  description = "Subnets to use for deployment"
}

variable "tags" {
  type        = map(string)
  description = "Tags to use for AWS resources"
}

variable "vpc_id" {
  type        = string
  description = "VPC to use for deployment"
}
