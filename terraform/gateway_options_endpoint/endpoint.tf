resource "aws_api_gateway_method" "endpoint_method" {
  rest_api_id = "${var.rest_api_id}"
  resource_id = "${var.resource_id}"
  http_method = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "endpoint_method_responses" {
  rest_api_id = "${var.rest_api_id}"
  resource_id = "${var.resource_id}"
  http_method = "${aws_api_gateway_method.endpoint_method.http_method}"
  status_code = "200"
  response_models = { "application/json" = "Empty" }
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true,
    "method.response.header.Access-Control-Allow-Methods" = true,
    "method.response.header.Access-Control-Allow-Origin" = true
  }
  depends_on = ["aws_api_gateway_method.endpoint_method"]
}

resource "aws_api_gateway_integration" "endpoint_integrations" {
  rest_api_id = "${var.rest_api_id}"
  resource_id = "${var.resource_id}"
  http_method = "${aws_api_gateway_method.endpoint_method.http_method}"
  passthrough_behavior = "WHEN_NO_MATCH"
  type = "MOCK"
  request_templates = { 
    "application/json" = <<TEMPLATE
{ "statusCode": 200 }
TEMPLATE
  }
  depends_on = ["aws_api_gateway_method.endpoint_method"]
}

resource "aws_api_gateway_integration_response" "endpoint_integration_responses" {
  rest_api_id = "${var.rest_api_id}"
  resource_id = "${var.resource_id}"
  http_method = "${aws_api_gateway_method.endpoint_method.http_method}"
  status_code = "200"
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type'",
    "method.response.header.Access-Control-Allow-Methods" = "'POST,OPTIONS,GET,PUT'",
    "method.response.header.Access-Control-Allow-Origin" = "'*'"
  }
  depends_on = ["aws_api_gateway_integration.endpoint_integrations", "aws_api_gateway_method_response.endpoint_method_responses"]
}

