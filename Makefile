.PHONY: cli_upload frontend_upload api_upload upload clean

RELEASE ?= dev
VERSION ?= $(shell git rev-parse --short HEAD)

all: build

clean:
	rm -r builds

builds/$(RELEASE)/:
	mkdir -p $@

build: builds/$(RELEASE)/
	make cli_build \
	&& make api_build \
	&& make frontend_build

api_build: builds/$(RELEASE)/list_applications.zip

frontend_build:
	@test $${FRONTEND_URL?is not set}
	@test $${API_URL?is not set}
	@test $${BINARY_NAME?is not set}
	mkdir -p builds/$(RELEASE)/frontend
	cd frontend \
	&& $(MAKE) -f makefile build

cli_build:
	mkdir -p builds/$(RELEASE)/cli
	cd cli \
	&& $(MAKE) -f makefile all

builds/$(RELEASE)/list_applications.zip: builds/$(RELEASE)/
# A temporary destination is used because we don't want multiple targets run at the same time to conflict - they all have to be named 'bootstrap'
	TMP_DST=$$(mktemp -d) ;\
	GOOS=linux GOARCH=amd64 go build \
		-tags lambda.norpc \
		-o $$TMP_DST/bootstrap lambda/$(subst .zip,,$(notdir $@))/main.go && \
	(cd $$TMP_DST && zip - bootstrap) > $@

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

api_upload: builds/$(RELEASE)/list_applications.zip
	@test $${S3_TF_BUCKET_NAME?is not set}
# tr is used to split $^, which is a space separated string, into newlines so that they can be
# passed to aws s3 cp one at a time.
#
# aws s3 cp doesnt support multiple targets, so we have to do one per line.
	echo $^ | tr " " "\n" | xargs -I{} -n1 aws s3 cp "{}" s3://$(S3_TF_BUCKET_NAME)/$(RELEASE)/
