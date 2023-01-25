resource "aws_lb" "keyconjurer" {
  name_prefix = "keycon"
  internal    = true
  subnets = var.subnets
  security_groups = concat(var.lb_security_group_ids, [
    aws_security_group.keyconjurer-lb.id
  ])
}

resource "aws_lb_listener" "keyconjurer" {
  load_balancer_arn = aws_lb.keyconjurer.arn
  # TODO: HTTPS
  # TODO: redirect all http to https
  port     = "80"
  protocol = "HTTP"

  default_action {
    type = "fixed-response"
    fixed_response {
      content_type = "text/plain"
      message_body = "Not found\n"
      status_code  = "404"
    }
  }
}

resource "aws_lb_listener_rule" "get_aws_creds" {
  listener_arn = aws_lb_listener.keyconjurer.arn

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.get_aws_creds.arn
  }

  condition {
    path_pattern {
      values = ["/get_aws_creds"]
    }
  }
}

resource "aws_lb_listener_rule" "list_providers" {
  listener_arn = aws_lb_listener.keyconjurer.arn

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.list_providers.arn
  }

  condition {
    path_pattern {
      values = ["/list_providers"]
    }
  }
}

resource "aws_lb_listener_rule" "get_user_data" {
  listener_arn = aws_lb_listener.keyconjurer.arn

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.get_user_data.arn
  }

  condition {
    path_pattern {
      values = ["/get_user_data"]
    }
  }
}

resource "aws_lb_target_group" "get_aws_creds" {
  name_prefix = "keycon"
  target_type = "lambda"
}

resource "aws_lb_target_group_attachment" "get_aws_creds" {
  target_group_arn = aws_lb_target_group.get_aws_creds.arn
  target_id        = aws_lambda_function.keyconjurer-get_aws_creds.arn
}

resource "aws_lambda_permission" "lb-get_aws_creds" {
  statement_id  = "LoadBalancer"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.keyconjurer-get_aws_creds.arn
  principal     = "elasticloadbalancing.amazonaws.com"
  source_arn    = aws_lb_target_group.get_aws_creds.arn
}

resource "aws_lb_target_group" "get_user_data" {
  name_prefix = "keycon"
  target_type = "lambda"
}

resource "aws_lb_target_group_attachment" "get_user_data" {
  target_group_arn = aws_lb_target_group.get_user_data.arn
  target_id        = aws_lambda_function.keyconjurer-get_user_data.arn
}

resource "aws_lambda_permission" "lb-get_user_data" {
  statement_id  = "LoadBalancer"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.keyconjurer-get_user_data.arn
  principal     = "elasticloadbalancing.amazonaws.com"
  source_arn    = aws_lb_target_group.get_user_data.arn
}

resource "aws_lb_target_group" "list_providers" {
  name_prefix = "keycon"
  target_type = "lambda"
}

resource "aws_lb_target_group_attachment" "list_providers" {
  target_group_arn = aws_lb_target_group.list_providers.arn
  target_id        = aws_lambda_function.keyconjurer-list_providers.arn
}

resource "aws_lambda_permission" "lb-list_providers" {
  statement_id  = "LoadBalancer"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.keyconjurer-list_providers.arn
  principal     = "elasticloadbalancing.amazonaws.com"
  source_arn    = aws_lb_target_group.list_providers.arn
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
