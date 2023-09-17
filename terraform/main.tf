module "list_applications" {
  source = "./modules/list_applications"
  listener_arn = aws_lb_listener.https.arn
  bucket_name = var.s3_tf_bucket
  environment = var.environment
  environment_variables = var.lambda_env
  subnets = var.subnets
  execution_role_arn = aws_iam_role.keyconjurer-lambda.arn
  security_group_ids = [
    aws_security_group.keyconjurer-default.id
  ]
}

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
