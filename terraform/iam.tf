resource "aws_iam_role" "keyconjurer-lambda" {
  name = "keyconjurer-lambda-${terraform.workspace}"
  description = "Used by keyconjurer-lambda-${terraform.workspace} to allow lambda execution in a VPC"
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
  tags = "${var.tags}"
}

resource "aws_iam_role_policy" "keyconjurer-lamdba" {
  name = "keyconjurer-lambda-policy-${terraform.workspace}"
  role = "${aws_iam_role.keyconjurer-lambda.id}"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
          "sts:AssumeRole"
       ],
            "Resource": [
                "arn:aws:sts::*:federated-user/*"
            ]
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "keyconjurer-lambda-basic-execution" {
  role = "${aws_iam_role.keyconjurer-lambda.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "keyconjurer-lambda-vpc-access-execution" {
  role = "${aws_iam_role.keyconjurer-lambda.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}
