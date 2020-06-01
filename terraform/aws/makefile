S3_TF_BUCKET_TAGS ?= "TagSet=[{Key=Name,Value=KeyConjurerS3Bucket}]"


ifndef S3_TF_BUCKET_NAME
$(error S3_TF_BUCKET_NAME is not set)
endif

ifndef S3_FRONTEND_BUCKET_NAME
$(error S3_FRONTEND_BUCKET_NAME is not set)
endif

STATE_BUCKET_EXISTS=$(shell aws s3api list-buckets --query "Buckets[].Name" | grep $(S3_TF_BUCKET_NAME))
FRONTEND_BUCKET_EXISTS=$(shell aws s3api list-buckets --query "Buckets[].Name" | grep $(S3_FRONTEND_BUCKET_NAME)-$(TF_WORKSPACE))

deploy: upload terraform_apply

plan_deploy: upload terraform_plan

setup_buckets:
ifeq ($(STATE_BUCKET_EXISTS),)
	aws s3api create-bucket --bucket ${S3_TF_BUCKET_NAME} --region us-west-2 --create-bucket-configuration LocationConstraint=us-west-2 
	aws s3api put-bucket-tagging --bucket ${S3_TF_BUCKET_NAME} --tagging '${S3_TF_BUCKET_TAGS}' 
else
	$(info state bucket is already setup)
endif 
ifeq ($(FRONTEND_BUCKET_EXISTS),)
	aws s3api create-bucket --bucket $(S3_FRONTEND_BUCKET_NAME)-$(TF_WORKSPACE) --region us-west-2 --create-bucket-configuration LocationConstraint=us-west-2 
	aws s3api put-bucket-tagging --bucket $(S3_FRONTEND_BUCKET_NAME)-$(TF_WORKSPACE) --tagging '${S3_TF_BUCKET_TAGS}'
else
	$(info frontend bucket is already setup)
endif

upload: api_upload cli_upload frontend_upload

cli_upload: 
	cd ../../builds/$(TF_WORKSPACE)/cli \
	&& aws s3 cp . s3://$(S3_FRONTEND_BUCKET_NAME)-$(TF_WORKSPACE) --exclude "*" --include "keyconjurer*" --recursive

frontend_upload: 
	cd ../../builds/$(TF_WORKSPACE)/frontend \
	&& aws s3 cp . s3://$(S3_FRONTEND_BUCKET_NAME)-$(TF_WORKSPACE) --include "*" --recursive

api_upload: api_zip 
	cd ../../builds/$(TF_WORKSPACE)/aws \
	&& aws s3 cp . s3://$(S3_TF_BUCKET_NAME)/$(TF_WORKSPACE) --exclude "*" --include "*.zip" --recursive

api_zip:
	cd ../../builds/$(TF_WORKSPACE)/aws && \
	zip get_user_data.zip get_user_data \
	&& zip get_cloud_creds.zip get_cloud_creds

terraform_apply:
	sed -i'.bak' -e "s/<S3_TF_BUCKET_NAME>/${S3_TF_BUCKET_NAME}/" main.tf \
	&& terraform init \
	&& terraform apply ${TERRAFORM_FLAGS}

terraform_plan:
	sed -i'.bak' -e "s/<S3_TF_BUCKET_NAME>/${S3_TF_BUCKET_NAME}/" main.tf \
	&& terraform init \
	&& terraform plan ${TERRAFORM_FLAGS}
