resource "aws_iam_role" "keyconjurer-lambda" {
  name               = "keyconjurer-lambda-${terraform.workspace}"
  description        = "Used by keyconjurer-lambda-${terraform.workspace} to allow lambda execution in a VPC"
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

  statement {
    sid = "AllowDecryptAndEncryptCredentials"
    actions = [
      "kms:Encrypt",
      "kms:Decrypt"
    ]
    resources = [var.kms_key_arn]
  }
}

resource "aws_iam_role_policy" "keyconjurer-lamdba" {
  name   = "keyconjurer-lambda-policy-${terraform.workspace}"
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
