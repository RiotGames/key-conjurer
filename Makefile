.PHONY: cli_upload frontend_upload api_upload upload clean

RELEASE ?= dev
VERSION ?= $(shell git rev-parse --short HEAD)
# This runs on Linux machines. Mac users should override TIMESTAMP.
TIMESTAMP ?= $(shell date --iso-8601=minutes)

## Standard targets for all Makefiles in our team

all: build

clean:
	rm -rf cli/keyconjurer*
	rm -r frontend/build

test: frontend_test go_test

CLI_TARGETS = cli/keyconjurer-darwin cli/keyconjurer-darwin-amd64 cli/keyconjurer-darwin-arm64 cli/keyconjurer-linux cli/keyconjurer-linux-amd64 cli/keyconjurer-linux-arm64 cli/keyconjurer-windows.exe
build: api_build frontend/build/index.html $(CLI_TARGETS)

go_test:
	go test ./...

## Frontend Build Targets
frontend_test:
	cd frontend && CI=true npm test

frontend/node_modules:
	cd frontend && npm install

frontend/build/index.html: frontend/node_modules
	mkdir -p build/frontend/
	@test $${FRONTEND_URL?is not set}
	@test $${API_URL?is not set}
	cd frontend && \
	REACT_APP_VERSION='$(shell git rev-parse --short HEAD)-$(RELEASE)' \
	REACT_APP_API_URL=${API_URL} \
	npm run-script build

### CLI Build Targets
cli/keyconjurer-linux-arm64 cli/keyconjurer-linux:
	GOOS=linux GOARCH=amd64 BUILD_TARGET=keyconjurer-linux $(MAKE) cli/keyconjurer
	GOOS=linux GOARCH=arm64 BUILD_TARGET=keyconjurer-linux-arm64 $(MAKE) cli/keyconjurer

cli/keyconjurer-linux-amd64: cli/keyconjurer-linux
	cp cli/keyconjurer-linux cli/keyconjurer-linux-amd64

cli/keyconjurer-darwin-arm64 cli/keyconjurer-darwin:
	GOOS=darwin GOARCH=amd64 BUILD_TARGET=keyconjurer-darwin $(MAKE) cli/keyconjurer
	GOOS=darwin GOARCH=arm64 BUILD_TARGET=keyconjurer-darwin-arm64 $(MAKE) cli/keyconjurer

cli/keyconjurer-darwin-amd64: cli/keyconjurer-darwin
	cp cli/keyconjurer-darwin cli/keyconjurer-darwin-amd64

cli/keyconjurer-windows.exe:
	GOOS=windows GOARCH=amd64 BUILD_TARGET=keyconjurer-windows.exe $(MAKE) cli/keyconjurer

cli/keyconjurer:
	@test $${CLIENT_ID?is not set}
	@test $${OIDC_DOMAIN?is not set}
	@test $${SERVER_ADDRESS?is not set}
	cd cli && \
	go build \
		-ldflags "\
			-s -w \
			-X main.Version=$(shell git rev-parse --short HEAD)-$(RELEASE) \
			-X main.ClientID=$(CLIENT_ID) \
			-X main.OIDCDomain=$(OIDC_DOMAIN) \
			-X main.BuildTimestamp='$(TIMESTAMP)' \
			-X main.ServerAddress=$(SERVER_ADDRESS)" \
		-o $(BUILD_TARGET)

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
	cd cli/ && \
	aws s3 cp . s3://$(S3_FRONTEND_BUCKET_NAME)-$(RELEASE) --exclude "*" --include "keyconjurer*" --recursive

frontend_upload: frontend/build/index.html
	@test $${S3_FRONTEND_BUCKET_NAME?is not set}
	@test $${RELEASE?is not set}
	cd frontend/build && \
	aws s3 cp . s3://$(S3_FRONTEND_BUCKET_NAME)-$(RELEASE) --include "*" --recursive

api_upload: build/list_applications.zip
	@test $${S3_TF_BUCKET_NAME?is not set}
# tr is used to split $^, which is a space separated string, into newlines so that they can be
# passed to aws s3 cp one at a time.
#
# aws s3 cp doesnt support multiple targets, so we have to do one per line.
	echo $^ | tr " " "\n" | xargs -I{} -n1 aws s3 cp "{}" s3://$(S3_TF_BUCKET_NAME)/$(RELEASE)/
