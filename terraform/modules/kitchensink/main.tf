# This main.tf file can be used to deploy a version of KeyConjurer.
# For more advanced deployments, you should use the constituent modules separately.

module "frontend" {
  source          = "../frontend"
  create_waf_acl  = var.create_waf_acl
  waf_acl_id      = var.waf_acl_id
  bucket_name     = var.s3_tf_bucket
  certificate_arn = var.frontend_cert
  domain          = var.frontend_domain
  account_number  = var.account_number
}

module "loadbalancer" {
  source          = "../loadbalancer"
  subnets         = var.subnets
  certificate_arn = var.api_cert
  security_group_ids = [
    aws_security_group.keyconjurer-lb.id
  ]
}

module "list_applications" {
  source                = "../list_applications"
  listener_arn          = module.loadbalancer.https_listener_arn
  bucket_name           = var.s3_tf_bucket
  environment           = var.environment
  environment_variables = var.lambda_env
  subnets               = var.subnets
  execution_role_arn    = aws_iam_role.keyconjurer-lambda.arn
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

resource "aws_security_group" "keyconjurer-lb" {
  name_prefix = "keyconjurer-lb"
  vpc_id      = var.vpc_id

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}
