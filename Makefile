.PHONY: cli_upload frontend_upload api_upload upload clean

RELEASE ?= dev
VERSION ?= $(shell git rev-parse --short HEAD)

## Standard targets for all Makefiles in our team

all: build

clean:
	rm -r build

test: frontend_test go_test

build: api_build build/frontend/index.html build/cli/keyconjurer-darwin build/cli/keyconjurer-darwin-amd64 build/cli/keyconjurer-darwin-arm64 build/cli/keyconjurer-linux build/cli/keyconjurer-linux-amd64 build/cli/keyconjurer-linux-arm64 build/cli/keyconjurer-windows.exe

go_test:
	go test ./...

## Frontend Build Targets
frontend_test:
	cd frontend && CI=true npm test

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

### CLI Build Targets
build/cli/keyconjurer-linux-arm64 build/cli/keyconjurer-linux:
	GOOS=linux GOARCH=amd64 BUILD_TARGET=keyconjurer-linux $(MAKE) cli/keyconjurer
	GOOS=linux GOARCH=arm64 BUILD_TARGET=keyconjurer-linux-arm64 $(MAKE) cli/keyconjurer

build/cli/keyconjurer-linux-amd64: build/cli/keyconjurer-linux
	cp build/cli/keyconjurer-linux build/cli/keyconjurer-linux-amd64

build/cli/keyconjurer-darwin-arm64 build/cli/keyconjurer-darwin:
	GOOS=darwin GOARCH=amd64 BUILD_TARGET=keyconjurer-darwin $(MAKE) cli/keyconjurer
	GOOS=darwin GOARCH=arm64 BUILD_TARGET=keyconjurer-darwin-arm64 $(MAKE) cli/keyconjurer

build/cli/keyconjurer-darwin-amd64: build/cli/keyconjurer-darwin
	cp build/cli/keyconjurer-darwin build/cli/keyconjurer-darwin-amd64

build/cli/keyconjurer-windows.exe:
	GOOS=windows GOARCH=amd64 BUILD_TARGET=keyconjurer-windows.exe $(MAKE) cli/keyconjurer

cli/keyconjurer:
	@test $${CLIENT_ID?is not set}
	@test $${OIDC_DOMAIN?is not set}
	@test $${SERVER_ADDRESS?is not set}
	@mkdir -p build/cli
	cd cli && \
	go build \
		-ldflags "\
			-X main.Version=$(shell git rev-parse --short HEAD)-$(RELEASE) \
			-X main.ClientID=$(CLIENT_ID) \
			-X main.OIDCDomain=$(OIDC_DOMAIN) \
			-X main.BuildTimestamp='$(shell date --iso-8601=minutes)' \
			-X main.ServerAddress=$(SERVER_ADDRESS)" \
		-o ../build/cli/$(BUILD_TARGET)

## API Build Targets
api_build: build/list_applications.zip

build/list_applications.zip:
	mkdir -p build/cli
# A temporary destination is used because we don't want multiple targets run at the same time to conflict - they all have to be named 'bootstrap'
	TMP_DST=$$(mktemp -d) ;\
	GOOS=linux GOARCH=amd64 go build \
		-tags lambda.norpc \
		-o $$TMP_DST/bootstrap lambda/$(subst .zip,,$(notdir $@))/main.go && \
	(cd $$TMP_DST && zip - bootstrap) > $@


## Upload Targets
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
