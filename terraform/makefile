S3_TF_BUCKET_TAGS ?= "TagSet=[{Key=Name,Value=KeyConjurerS3Bucket}]"
RELEASE ?= dev

upload: api_upload cli_upload frontend_upload

cli_upload:
	@test $${S3_FRONTEND_BUCKET_NAME?is not set}
	@test $${RELEASE?is not set}
	cd ../builds/$(RELEASE)/cli \
	&& aws s3 cp . s3://$(S3_FRONTEND_BUCKET_NAME)-$(RELEASE) --exclude "*" --include "keyconjurer*" --recursive

frontend_upload:
	@test $${S3_FRONTEND_BUCKET_NAME?is not set}
	@test $${RELEASE?is not set}
	cd ../builds/$(RELEASE)/frontend \
	&& aws s3 cp . s3://$(S3_FRONTEND_BUCKET_NAME)-$(RELEASE) --include "*" --recursive

api_upload: ../builds/$(RELEASE)/list_applications_v2.zip
	@test $${S3_TF_BUCKET_NAME?is not set}
# tr is used to split $^, which is a space separated string, into newlines so that they can be
# passed to aws s3 cp one at a time.
#
# aws s3 cp doesnt support multiple targets, so we have to do one per line.
	echo $^ | tr " " "\n" | xargs -I{} -n1 aws s3 cp "{}" s3://$(S3_TF_BUCKET_NAME)/$(RELEASE)/

