// SETTINGS
provider "aws" {
  region = var.region
}

locals {
  bucket = var.s3_tf_bucket
  region = var.region
}

terraform {
  backend "s3" {
    bucket  = "keyconjurer-tf"
    key     = "state.tfstate"
    region  = "us-west-2"
    encrypt = "true"
  }
}
