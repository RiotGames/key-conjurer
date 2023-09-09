S3_TF_BUCKET_TAGS ?= "TagSet=[{Key=Name,Value=KeyConjurerS3Bucket}]"
RELEASE ?= dev

ifndef S3_TF_BUCKET_NAME
$(error S3_TF_BUCKET_NAME is not set)
endif

ifndef S3_FRONTEND_BUCKET_NAME
$(error S3_FRONTEND_BUCKET_NAME is not set)
endif

upload: api_upload cli_upload frontend_upload

cli_upload:
	cd ../builds/$(RELEASE)/cli \
	&& aws s3 cp . s3://$(S3_FRONTEND_BUCKET_NAME)-$(RELEASE) --exclude "*" --include "keyconjurer*" --recursive

frontend_upload:
	cd ../builds/$(RELEASE)/frontend \
	&& aws s3 cp . s3://$(S3_FRONTEND_BUCKET_NAME)-$(RELEASE) --include "*" --recursive

api_upload: ../builds/$(RELEASE)/list_applications_v2.zip ../builds/$(RELEASE)/cloud/get_user_data.zip ../builds/$(RELEASE)/cloud/list_providers.zip ../builds/$(RELEASE)/cloud/get_cloud_creds.zip
	echo $^ | xargs -I{} -n1 aws s3 cp "{}" s3://$(S3_TF_BUCKET_NAME)/$(RELEASE)
