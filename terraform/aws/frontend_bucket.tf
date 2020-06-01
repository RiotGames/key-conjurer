resource "aws_s3_bucket" "keyconjurer_frontend" {
  bucket = "keyconjurer-frontend-${terraform.workspace}"
  acl    = "private"
  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Id": "KeyConjurerAccess",
    "Statement": [
        {
            "Sid": "Grant a CloudFront Origin Identity access",
            "Effect": "Allow",
            "Principal": {
                "CanonicalUser": "${aws_cloudfront_origin_access_identity.keyconjurer_identity.s3_canonical_user_id}"
            },
            "Action": "s3:GetObject",
            "Resource": "arn:aws:s3:::keyconjurer-frontend-${terraform.workspace}/*"
        },
        {
            "Sid": "CI Upload",
            "Effect": "Allow",
            "Principal": { "AWS": "arn:aws:iam::${var.account_number}:role/infosec_ci" },
            "Action": "s3:PutObject",
            "Resource": "arn:aws:s3:::keyconjurer-frontend-${terraform.workspace}/*"
        }
    ]
}
POLICY

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }

  tags = var.tags
}
