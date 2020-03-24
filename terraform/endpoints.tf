resource "aws_api_gateway_resource" "get_user_data" {
  rest_api_id = aws_api_gateway_rest_api.keyconjurer.id
  parent_id   = aws_api_gateway_rest_api.keyconjurer.root_resource_id
  path_part   = "get_user_data"
}

resource "aws_api_gateway_resource" "get_aws_creds" {
  rest_api_id = aws_api_gateway_rest_api.keyconjurer.id
  parent_id   = aws_api_gateway_rest_api.keyconjurer.root_resource_id
  path_part   = "get_aws_creds"
}

// METHODS
module "post_get_user_data" {
  source         = "./gateway_post_endpoint"
  account_number = var.account_number
  region         = var.region
  rest_api_id    = aws_api_gateway_rest_api.keyconjurer.id
  resource_id    = aws_api_gateway_resource.get_user_data.id
  uri_arn        = aws_lambda_function.keyconjurer-get_user_data.invoke_arn
  lambda_arn     = aws_lambda_function.keyconjurer-get_user_data.arn
}

module "post_get_aws_creds" {
  source         = "./gateway_post_endpoint"
  account_number = var.account_number
  region         = var.region
  rest_api_id    = aws_api_gateway_rest_api.keyconjurer.id
  resource_id    = aws_api_gateway_resource.get_aws_creds.id
  uri_arn        = aws_lambda_function.keyconjurer-get_aws_creds.invoke_arn
  lambda_arn     = aws_lambda_function.keyconjurer-get_aws_creds.arn
}

module "options_get_user_data" {
  source      = "./gateway_options_endpoint"
  rest_api_id = aws_api_gateway_rest_api.keyconjurer.id
  resource_id = aws_api_gateway_resource.get_user_data.id
}

module "options_get_aws_creds" {
  source      = "./gateway_options_endpoint"
  rest_api_id = aws_api_gateway_rest_api.keyconjurer.id
  resource_id = aws_api_gateway_resource.get_aws_creds.id
}

