.PHONY: cli_upload frontend_upload api_upload upload clean

RELEASE ?= dev
VERSION ?= $(shell git rev-parse --short HEAD)

## Standard targets for all Makefiles in our team

all: build

clean:
	rm -r build

test: frontend_test go_test

build: api_build build/frontend/index.html build/cli/keyconjurer-darwin build/cli/keyconjurer-darwin-amd64 build/cli/keyconjurer-darwin-arm64 build/cli/keyconjurer-linux build/cli/keyconjurer-linux-amd64 build/cli/keyconjurer-linux-arm64 build/cli/keyconjurer-windows.exe

# Multiple targets are used because Make can parallelize them
# If we had test commands in a single target, Make would serially run it instead.
frontend_test:
	cd frontend && CI=true npm test

go_test:
	go test ./...

api_build: build/list_applications.zip

frontend/node_modules:
	cd frontend && npm install

build/frontend/index.html: frontend/node_modules
	mkdir -p build/frontend/
	@test $${FRONTEND_URL?is not set}
	@test $${API_URL?is not set}
	cd frontend && \
	REACT_APP_VERSION='$$(git rev-parse --short HEAD)-$(RELEASE)' \
	REACT_APP_API_URL=${API_URL} \
	REACT_APP_BINARY_NAME=${BINARY_NAME} \
	REACT_APP_DOCUMENTATION_URL=${REACT_APP_DOCUMENTATION_URL} \
	REACT_APP_CLIENT=webUI npm run-script build
	cp -R frontend/build/* build/frontend/

build/cli/keyconjurer-darwin build/cli/keyconjurer-darwin-amd64 build/cli/keyconjurer-darwin-arm64 build/cli/keyconjurer-linux build/cli/keyconjurer-linux-amd64 build/cli/keyconjurer-linux-arm64 build/cli/keyconjurer-windows.exe:
	mkdir -p build/cli
	cd cli && $(MAKE) -f makefile all

build/list_applications.zip:
	mkdir -p build/cli
# A temporary destination is used because we don't want multiple targets run at the same time to conflict - they all have to be named 'bootstrap'
	TMP_DST=$$(mktemp -d) ;\
	GOOS=linux GOARCH=amd64 go build \
		-tags lambda.norpc \
		-o $$TMP_DST/bootstrap lambda/$(subst .zip,,$(notdir $@))/main.go && \
	(cd $$TMP_DST && zip - bootstrap) > $@

upload: api_upload cli_upload frontend_upload

cli_upload: $(CLI_TARGETS)
	@test $${S3_FRONTEND_BUCKET_NAME?is not set}
	@test $${RELEASE?is not set}
	cd build/cli && \
	aws s3 cp . s3://$(S3_FRONTEND_BUCKET_NAME)-$(RELEASE) --exclude "*" --include "keyconjurer*" --recursive

frontend_upload: build/frontend/index.html
	@test $${S3_FRONTEND_BUCKET_NAME?is not set}
	@test $${RELEASE?is not set}
	cd build/frontend && \
	aws s3 cp . s3://$(S3_FRONTEND_BUCKET_NAME)-$(RELEASE) --include "*" --recursive

api_upload: build/list_applications.zip
	@test $${S3_TF_BUCKET_NAME?is not set}
# tr is used to split $^, which is a space separated string, into newlines so that they can be
# passed to aws s3 cp one at a time.
#
# aws s3 cp doesnt support multiple targets, so we have to do one per line.
	echo $^ | tr " " "\n" | xargs -I{} -n1 aws s3 cp "{}" s3://$(S3_TF_BUCKET_NAME)/$(RELEASE)/
