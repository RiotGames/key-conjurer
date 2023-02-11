S3_TF_BUCKET_TAGS ?= "TagSet=[{Key=Name,Value=KeyConjurerS3Bucket}]"

ifndef RELEASE
$(error RELEASE is not set)
endif

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

api_upload: api_zip
	cd ../builds/$(RELEASE)/aws \
	&& aws s3 cp . s3://$(S3_TF_BUCKET_NAME)/$(RELEASE) --exclude "*" --include "*.zip" --recursive

api_zip:
	cd ../builds/$(RELEASE)/aws && \
	zip get_user_data.zip get_user_data \
	&& zip get_cloud_creds.zip get_cloud_creds \
	&& zip list_providers.zip list_providers
