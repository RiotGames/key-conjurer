resource "aws_iam_role" "keyconjurer-lambda" {
  name               = var.lambda_execution_role_name
  description        = "Used by KeyConjurer Lambda functions to access protected resources"
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

data "aws_iam_policy_document" "keyconjurer_lambda_permissions" {
  statement {
    sid       = "AllowAssumeRoleIntoFederatedUser"
    actions   = ["sts:AssumeRole"]
    resources = ["arn:aws:sts::*:federated-user/*"]
  }
}

resource "aws_iam_role_policy" "keyconjurer-lamdba" {
  name   = "${var.lambda_execution_role_name}-policy"
  role   = aws_iam_role.keyconjurer-lambda.id
  policy = data.aws_iam_policy_document.keyconjurer_lambda_permissions.json
}

resource "aws_iam_role_policy_attachment" "keyconjurer-lambda-basic-execution" {
  role       = aws_iam_role.keyconjurer-lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "keyconjurer-lambda-vpc-access-execution" {
  role       = aws_iam_role.keyconjurer-lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}
