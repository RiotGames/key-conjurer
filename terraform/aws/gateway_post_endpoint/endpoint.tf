resource "aws_api_gateway_method" "endpoint_method" {
  rest_api_id   = var.rest_api_id
  resource_id   = var.resource_id
  http_method   = "POST"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "endpoint_method_responses" {
  rest_api_id = var.rest_api_id
  resource_id = var.resource_id
  http_method = aws_api_gateway_method.endpoint_method.http_method
  status_code = "200"
  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = true
  }
  depends_on = [aws_api_gateway_method.endpoint_method]
}


resource "aws_api_gateway_integration" "endpoint_proxy_integrations" {
  rest_api_id             = var.rest_api_id
  resource_id             = var.resource_id
  http_method             = aws_api_gateway_method.endpoint_method.http_method
  integration_http_method = aws_api_gateway_method.endpoint_method.http_method
  type                    = "AWS_PROXY"
  uri                     = var.uri_arn
  depends_on              = [aws_api_gateway_method.endpoint_method]
}

resource "aws_api_gateway_integration_response" "endpoint_integration_responses" {
  rest_api_id = var.rest_api_id
  resource_id = var.resource_id
  http_method = aws_api_gateway_method.endpoint_method.http_method
  status_code = "200"
  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = "'*'"
  }
  depends_on = [aws_api_gateway_integration.endpoint_proxy_integrations, aws_api_gateway_method_response.endpoint_method_responses]
}

resource "aws_lambda_permission" "lambda_permissions" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = var.lambda_arn
  principal     = "apigateway.amazonaws.com"

  source_arn = "arn:aws:execute-api:${var.region}:${var.account_number}:${var.rest_api_id}/*"
}
