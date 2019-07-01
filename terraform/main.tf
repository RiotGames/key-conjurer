// SETTINGS
provider "aws" {
    region = "${var.settings["region"]}"
}

terraform {
    backend "s3" {
	bucket = "keyconjurer-tf"
	key = "state.tfstate"
	region = "us-west-2"
	encrypt = "true"
    }
}
