resource "aws_security_group" "keyconjurer-default" {
  name_prefix = "keyconjurer"
  description = "default security group to allow most expected protocols"
  vpc_id      = var.vpc_id

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}

resource "aws_lambda_function" "keyconjurer-get_aws_creds" {
  function_name    = "keyconjurer-${var.environment}-get_aws_creds"
  description      = "[${var.environment}] Retrieves STS tokens from AWS after validating the user via OneLogin and MFA"
  s3_bucket        = var.s3_tf_bucket
  s3_key           = "${var.environment}/get_cloud_creds.zip"
  source_code_hash = "true"
  role             = aws_iam_role.keyconjurer-lambda.arn
  handler          = "get_cloud_creds"
  runtime          = "go1.x"
  timeout          = 300

  environment {
    variables = var.lambda_env
  }

  vpc_config {
    subnet_ids         = var.subnets
    security_group_ids = [aws_security_group.keyconjurer-default.id]
  }
}

resource "aws_lambda_function" "keyconjurer-get_user_data" {
  function_name    = "keyconjurer-${var.environment}-get_user_data"
  description      = "[${var.environment}] Retrieves user access information from OneLogin"
  s3_bucket        = var.s3_tf_bucket
  s3_key           = "${var.environment}/get_user_data.zip"
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
}


resource "aws_lambda_function" "keyconjurer-list_providers" {
  function_name    = "keyconjurer-${var.environment}-list_providers"
  description      = "[${var.environment}] List the providers a user can use"
  s3_bucket        = var.s3_tf_bucket
  s3_key           = "${var.environment}/list_providers.zip"
  source_code_hash = "true"
  role             = aws_iam_role.keyconjurer-lambda.arn
  handler          = "list_providers"
  runtime          = "go1.x"
  timeout          = 300

  environment {
    variables = var.lambda_env
  }

  vpc_config {
    subnet_ids         = var.subnets
    security_group_ids = [aws_security_group.keyconjurer-default.id]
  }
}

resource "aws_lambda_function" "keyconjurer-list_applications_v2" {
  function_name    = "keyconjurer-${var.environment}-list_applications_v2"
  description      = "[${var.environment}] List the providers a user can use"
  s3_bucket        = var.s3_tf_bucket
  s3_key           = "${var.environment}/list_applications_v2.zip"
  source_code_hash = "true"
  role             = aws_iam_role.keyconjurer-lambda.arn
  handler          = "bootstrap"
  runtime          = "provider.al2"
  timeout          = 300

  environment {
    variables = var.lambda_env
  }

  vpc_config {
    subnet_ids         = var.subnets
    security_group_ids = [aws_security_group.keyconjurer-default.id]
  }
}
