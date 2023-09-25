resource "aws_lb_target_group" "lambda" {
  name_prefix = "keycon"
  target_type = "lambda"
}

resource "aws_lb_listener_rule" "lambda" {
  listener_arn = var.listener_arn

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.lambda.arn
  }

  condition {
    path_pattern {
      values = ["/v2/applications"]
    }
  }
}

resource "aws_lb_target_group_attachment" "lambda" {
  target_group_arn = aws_lb_target_group.lambda.arn
  target_id        = aws_lambda_function.lambda.arn
}

resource "aws_lambda_permission" "lambda" {
  statement_id  = "LoadBalancer"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda.arn
  principal     = "elasticloadbalancing.amazonaws.com"
  source_arn    = aws_lb_target_group.lambda.arn
}
