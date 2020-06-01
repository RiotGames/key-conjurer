output "workspace" {
  value = terraform.workspace
}

output "rest_api_endpoint" {
  value = "${aws_api_gateway_rest_api.keyconjurer.id}.execute-api.${var.region}.amazonaws.com"
}

output "deployment_id" {
  value = aws_api_gateway_deployment.live.id
}


