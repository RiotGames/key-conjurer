resource "aws_lambda_function" "lambda" {
  function_name    = "keyconjurer-${var.environment}-list_applications_v2"
  description      = "[${var.environment}] List the providers a user can use"
  s3_bucket        = var.bucket_name
  s3_key           = "${var.environment}/list_applications_v2.zip"
  source_code_hash = "true"
  role             = var.execution_role_arn
  handler          = "bootstrap"
  runtime          = "provided.al2"
  timeout          = 300

  environment {
    variables = var.environment_variables
  }

  vpc_config {
    subnet_ids         = var.subnets
    security_group_ids = var.security_group_ids
  }
}
