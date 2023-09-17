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

resource "aws_lambda_function" "keyconjurer-list_applications_v2" {
  function_name    = "keyconjurer-${var.environment}-list_applications_v2"
  description      = "[${var.environment}] List the providers a user can use"
  s3_bucket        = var.s3_tf_bucket
  s3_key           = "${var.environment}/list_applications_v2.zip"
  source_code_hash = "true"
  role             = aws_iam_role.keyconjurer-lambda.arn
  handler          = "bootstrap"
  runtime          = "provided.al2"
  timeout          = 300

  environment {
    variables = var.lambda_env
  }

  vpc_config {
    subnet_ids         = var.subnets
    security_group_ids = [aws_security_group.keyconjurer-default.id]
  }
}
