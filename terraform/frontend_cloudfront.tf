resource "aws_cloudfront_origin_access_identity" "keyconjurer_identity" {
  comment = "Key Conjurer ${terraform.workspace} bucket access"
}

resource "aws_cloudfront_distribution" "keyconjurer_distribution" {
  enabled             = true
  default_root_object = "index.html"
  // US, Canada, Europe only
  price_class = "PriceClass_100"
  aliases     = [var.frontend_domain]

  origin {
    domain_name = aws_s3_bucket.keyconjurer_frontend.bucket_regional_domain_name
    origin_id   = "keyconjurer-origin"

    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.keyconjurer_identity.cloudfront_access_identity_path
    }
  }

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    default_ttl            = 300 // 5 minutes
    max_ttl                = 300 // 5 minutes
    target_origin_id       = "keyconjurer-origin"
    viewer_protocol_policy = "redirect-to-https"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn = var.frontend_cert
    ssl_support_method  = "sni-only"
  }

  web_acl_id = var.create_waf_acl ? aws_waf_web_acl.keyconjurer_waf_acl[0].id : var.waf_acl_id
}
