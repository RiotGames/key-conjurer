variable "account_number" {
  description = "Account you're deploying to.  Needed for policy creation."
}

variable "region" {
  description = "AWS region to deploy to"
}

variable "rest_api_id" {
  description = "Rest API ID"
}

variable "resource_id" {
  description = "Resource for the POST request to attach to"
}

variable "uri_arn" {
  description = "Invoke arn for lambda"
}

variable "lambda_arn" {
  description = "ARN for the lambda to be invoked"
}