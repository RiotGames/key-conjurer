resource "aws_security_group" "keyconjurer-default" {
	name = "Key Conjurer ${terraform.workspace}"
	description = "default security group to allow most expected protocols"
	vpc_id = "${var.settings["vpc_id"]}"

	ingress {
		from_port = 8080
		to_port = 8080
		protocol = "tcp"
		cidr_blocks = "${var.vpc_config["cidrs"]}"
	}

	ingress {
		from_port = 8443
		to_port = 8443
		protocol = "tcp"
		cidr_blocks = "${var.vpc_config["cidrs"]}"
	}

	ingress {
		from_port = 22
		to_port = 22
		protocol = "tcp"
		cidr_blocks = "${var.vpc_config["cidrs"]}"
	}

	egress {
		from_port = 0
		to_port = 0
		protocol = "-1"
		cidr_blocks = ["0.0.0.0/0"]
	}

	tags = "${var.tags}"
}

resource "aws_lambda_function" "keyconjurer-get_aws_creds" {
  function_name = "keyconjurer-${terraform.workspace}-get_aws_creds"
  description = "[${terraform.workspace}] Retrieves STS tokens from AWS after validating the user via OneLogin and MFA"
  s3_bucket = "${var.settings["s3_bucket"]}"
  s3_key = "${terraform.workspace}/get_aws_creds.zip"
  source_code_hash = "true"
  role = "${aws_iam_role.keyconjurer-lambda.arn}"
  handler = "get_aws_creds"
  runtime = "go1.x"
  timeout = 300

  environment {
    variables = "${var.lambda_env}"
  }

  vpc_config {
    subnet_ids = "${var.vpc_config["subnets"]}"
    security_group_ids = ["${aws_security_group.keyconjurer-default.id}"]
  }

  tags = "${var.tags}"
}

resource "aws_lambda_function" "keyconjurer-get_user_data" {
  function_name = "keyconjurer-${terraform.workspace}-get_user_data"
  description = "[${terraform.workspace}] Retrieves user access information from OneLogin"
  s3_bucket = "${var.settings["s3_bucket"]}"
  s3_key = "${terraform.workspace}/get_user_data.zip"
  source_code_hash = "true"
  role = "${aws_iam_role.keyconjurer-lambda.arn}"
  handler = "get_user_data"
  runtime = "go1.x"
  timeout = 300

  environment {
    variables = "${var.lambda_env}"
  }

  vpc_config {
    subnet_ids = "${var.vpc_config["subnets"]}"
    security_group_ids = ["${aws_security_group.keyconjurer-default.id}"]
  }

  tags = "${var.tags}"
}
