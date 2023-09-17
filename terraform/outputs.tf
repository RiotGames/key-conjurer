output "url" {
  value = module.loadbalancer.dns_name
}

output "cloudfront_distribution_url" {
  value = module.frontend.domain_name
}
