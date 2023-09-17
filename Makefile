.PHONY: cli_upload frontend_upload api_upload upload clean

RELEASE ?= dev
VERSION ?= $(shell git rev-parse --short HEAD)

all: build

clean:
	rm -r builds

# Multiple targets are used because Make can parallelize them
# If we had test commands in a single target, Make would serially run it instead.
frontend_test:
	cd frontend && CI=true npm test

go_test:
	go test ./...

test: frontend_test go_test

builds/$(RELEASE)/:
	mkdir -p $@

build: builds/$(RELEASE)/
	make cli_build \
	&& make api_build \
	&& make frontend_build

api_build: builds/$(RELEASE)/list_applications.zip

frontend/node_modules:
	cd frontend && npm install

frontend_build: frontend/node_modules
	mkdir -p builds/$(RELEASE)/frontend/
	@test $${FRONTEND_URL?is not set}
	@test $${API_URL?is not set}
	cd frontend && \
	REACT_APP_VERSION='$$(git rev-parse --short HEAD)-$(RELEASE)' REACT_APP_API_URL=${API_URL} REACT_APP_BINARY_NAME=${BINARY_NAME} REACT_APP_DOCUMENTATION_URL=${REACT_APP_DOCUMENTATION_URL} REACT_APP_CLIENT=webUI npm run-script build
	cp -R frontend/build/* builds/$(RELEASE)/frontend/

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
