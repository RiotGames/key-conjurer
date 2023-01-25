output "url" {
  value = aws_lb.keyconjurer.dns_name
}

output "cloudfront_distribution_url" {
  value = aws_cloudfront_distribution.keyconjurer_distribution.domain_name
}
