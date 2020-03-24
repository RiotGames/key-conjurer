resource "aws_security_group" "keyconjurer-default" {
  name        = "Key Conjurer ${terraform.workspace}"
  description = "default security group to allow most expected protocols"
  vpc_id      = var.vpc_id

  tags = var.tags
}

resource "aws_lambda_function" "keyconjurer-get_aws_creds" {
  function_name    = "keyconjurer-${terraform.workspace}-get_aws_creds"
  description      = "[${terraform.workspace}] Retrieves STS tokens from AWS after validating the user via OneLogin and MFA"
  s3_bucket        = var.s3_tf_bucket
  s3_key           = "${terraform.workspace}/get_aws_creds.zip"
  source_code_hash = "true"
  role             = aws_iam_role.keyconjurer-lambda.arn
  handler          = "get_aws_creds"
  runtime          = "go1.x"
  timeout          = 300

  environment {
    variables = var.lambda_env
  }

  vpc_config {
    subnet_ids         = var.subnets
    security_group_ids = [aws_security_group.keyconjurer-default.id]
  }

  tags = var.tags
}

resource "aws_lambda_function" "keyconjurer-get_user_data" {
  function_name    = "keyconjurer-${terraform.workspace}-get_user_data"
  description      = "[${terraform.workspace}] Retrieves user access information from OneLogin"
  s3_bucket        = var.s3_tf_bucket
  s3_key           = "${terraform.workspace}/get_user_data.zip"
  source_code_hash = "true"
  role             = aws_iam_role.keyconjurer-lambda.arn
  handler          = "get_user_data"
  runtime          = "go1.x"
  timeout          = 300

  environment {
    variables = var.lambda_env
  }

  vpc_config {
    subnet_ids         = var.subnets
    security_group_ids = [aws_security_group.keyconjurer-default.id]
  }

  tags = var.tags
}
