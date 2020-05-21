provider "aws" {
  region = var.region
}

terraform {
  required_version = ">= 0.12.20"

  backend "s3" {
    // The bucket needs to be the same as S3_TF_BUCKET_NAME in the .env file
    //  This cannot be set by a variable
    bucket  = "keyconjurer-tf"
    key     = "state.tfstate"
    region  = "us-west-2"
    encrypt = "true"
  }
}
