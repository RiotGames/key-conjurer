resource "aws_lb" "keyconjurer" {
  name_prefix = "keycon"
  internal    = true
  subnets     = var.subnets
  security_groups = var.security_group_ids
}

resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.keyconjurer.arn
  certificate_arn   = var.api_certificate_arn

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
