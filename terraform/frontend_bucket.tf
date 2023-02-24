resource "aws_s3_bucket" "keyconjurer_frontend" {
  bucket = var.frontend_bucket_name
}

data "aws_iam_policy_document" "frontend_bucket_policy" {
  statement {
    sid       = "Cloudfront Access"
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.keyconjurer_frontend.arn}/*"]
    principals {
      type        = "CanonicalUser"
      identifiers = [aws_cloudfront_origin_access_identity.keyconjurer_identity.s3_canonical_user_id]
    }
  }

  statement {
    sid       = "CI Upload"
    actions   = ["s3:PutObject"]
    resources = ["${aws_s3_bucket.keyconjurer_frontend.arn}/*"]
    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${var.account_number}:role/infosec_ci"]
    }
  }
}

resource "aws_s3_bucket_policy" "frontend_bucket" {
  bucket = aws_s3_bucket.keyconjurer_frontend.bucket
  policy = data.aws_iam_policy_document.frontend_bucket_policy.json
}


resource "aws_s3_bucket_acl" "frontend_bucket" {
  bucket = aws_s3_bucket.keyconjurer_frontend.bucket
  acl    = "private"
}

resource "aws_s3_bucket_server_side_encryption_configuration" "frontend_bucket" {
  bucket = aws_s3_bucket.keyconjurer_frontend.bucket
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
