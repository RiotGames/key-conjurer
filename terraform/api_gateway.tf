resource "aws_api_gateway_rest_api" "keyconjurer" {
  name = "keyconjurer-${terraform.workspace}"
  description = "Key Conjurer ${terraform.workspace} API"
  endpoint_configuration {
    types = ["REGIONAL"]
  }

  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": "*",
            "Action": "execute-api:Invoke",
            "Resource": "execute-api:/*/OPTIONS/*"
        },
        {
            "Effect": "Allow",
            "Principal": "*",
            "Action": "execute-api:Invoke",
            "Resource": "execute-api:/*/POST/*"
        }
    ]
}
POLICY
}

resource "aws_api_gateway_deployment" "live" {
  depends_on = [
    "module.post_get_user_data",
    "module.post_get_aws_creds",
    "module.options_get_user_data",
    "module.options_get_aws_creds",
  ]
  rest_api_id = "${aws_api_gateway_rest_api.keyconjurer.id}"
  stage_name = "live"
  
  description = "Deployed at ${timestamp()}"
  lifecycle = {
    ignore_changes = ["description"]
    create_before_destroy = true
  }
}

resource "aws_api_gateway_domain_name" "api_domain_name" {
  domain_name = "${var.api_domains["${terraform.workspace}"]}"
  certificate_arn = "${var.api_certs["${terraform.workspace}"]}"
}

resource "aws_api_gateway_base_path_mapping" "live" {
  domain_name = "${aws_api_gateway_domain_name.api_domain_name.domain_name}"
  stage_name = "${aws_api_gateway_deployment.live.stage_name}"
  api_id = "${aws_api_gateway_rest_api.keyconjurer.id}"
}
