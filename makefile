S3_TF_BUCKET_TAGS ?= "TagSet=[{Key=Name,Value=KeyConjurerS3Bucket}]"

ifndef TF_WORKSPACE
$(error TF_WORKSPACE is not set)
endif

ifndef S3_TF_BUCKET_NAME
$(error S3_TF_BUCKET_NAME is not set)
endif

ifndef TF_VAR_FILE
$(error TF_VAR_FILE is not set)
endif

build:
	make cli_build \
	&& make api_build \
	&& make frontend_build

terraform_apply:
	cd terraform \
	&& tfswitch \
	&& terraform init \
	&& (terraform workspace select $(TF_WORKSPACE) || terraform workspace new $(TF_WORKSPACE)) \
	&& terraform apply -var-file=../$(TF_VAR_FILE) -auto-approve

upload:
	make api_upload \
	&& make cli_upload \
	&& make frontend_upload

deploy:
	make build \
	&& make upload \
	&& make terraform_apply

setup_buckets:
	aws s3api create-bucket --bucket $(S3_TF_BUCKET_NAME) --region us-west-2 --create-bucket-configuration LocationConstraint=us-west-2 \
	&& aws s3api put-bucket-tagging --bucket $(S3_TF_BUCKET_NAME) --tagging '$(S3_TF_BUCKET_TAGS)' \
	&& aws s3api create-bucket --bucket $(S3_TF_BUCKET_NAME) --region us-west-2 --create-bucket-configuration LocationConstraint=us-west-2 \
	&& aws s3api put-bucket-tagging --bucket $(S3_TF_BUCKET_NAME) --tagging '$(S3_TF_BUCKET_TAGS)'

api_build:
	cd api \
	&& $(MAKE) -f makefile build

api_upload:
	cd api \
	&& $(MAKE)  -f makefile zip \
	&& $(MAKE) -f makefile upload

frontend_build:
	cd frontend \
	&& $(MAKE) -f makefile build

frontend_upload:
	cd frontend \
	&& $(MAKE) -f makefile upload

frontend_file_reset:
	cd frontend \
	&& $(MAKE) -f makefile reset_files

cli_build:
	cd cli \
	&& $(MAKE)  -f makefile all

cli_upload:
	cd cli \
	&& $(MAKE)  -f makefile upload

reset_files: frontend_file_reset
