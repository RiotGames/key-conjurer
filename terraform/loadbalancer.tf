resource "aws_lb" "keyconjurer" {
  name_prefix = "keycon"
  internal    = true
  subnets     = var.subnets
  security_groups = concat(var.lb_security_group_ids, [
    aws_security_group.keyconjurer-lb.id
  ])
}

resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.keyconjurer.arn
  certificate_arn   = var.api_cert

  port     = "443"
  protocol = "HTTPS"

  default_action {
    type = "fixed-response"
    fixed_response {
      content_type = "text/plain"
      message_body = "Not found\n"
      status_code  = "404"
    }
  }
}

resource "aws_lb_listener" "https_redirect" {
  load_balancer_arn = aws_lb.keyconjurer.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
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
