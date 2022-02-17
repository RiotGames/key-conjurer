provider "aws" {
  region = var.region
}

terraform {
  required_version = ">= 0.13.1"
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "~> 3.0"
    }
  }

  backend "s3" {
    // The bucket needs to be the same as S3_TF_BUCKET_NAME in the .env file
    //  This cannot be set by a variable
    bucket  = "<S3_TF_BUCKET_NAME>"
    key     = "state.tfstate"
    region  = "us-west-2"
    encrypt = "true"
  }
}
