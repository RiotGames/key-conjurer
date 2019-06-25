settings = {
  account_number = "<aws account number>"
  region = "us-west-2"
  s3_bucket = "<example-tf-bucket>"
  vpc_id = "<vpc-id>"
  owner = "<email>"
  accounting = "<internal group name>"
  fe_bucket_name = "<bucket for frontend CDN>"
}

frontend_certs = {
  <workspace name> = "<arn of ACM certificate for frontend domain>"
}

api_certs = {
  <workspace name> = "<arn of ACM certificate for api domain>"
}

frontend_domains = {
  <workspace name> = ["keyconjurer.example.com"]
}

api_domains = {
  <workspace name> = "api.keyconjurer.example.com"
}

vpc_config = {
  subnets = [
    "<vpc subnet id>",
    ...]
  cidrs = ["0.0.0.0/0"]
}

tags = {
  Accounting = "<internal group name>"
  Name = "<deployment name>"
  Owner = "<email>"
}

lambda_env = {
  EncryptedSettings = "<kms encrypted json blob to be used in the Go settings struct>"
  AWSRegion = "us-west-2"
}


