.PHONY: cli_upload frontend_upload upload clean

RELEASE ?= dev
VERSION ?= $(shell git rev-parse --short HEAD)
# This runs on Linux machines. Mac users should override TIMESTAMP.
TIMESTAMP ?= $(-shell date --iso-8601=minutes)

## Standard targets for all Makefiles in our team

all: build

clean:
	rm -rf bin/keyconjurer*
	rm -r frontend/dist

test: frontend_test go_test

CLI_TARGETS = bin/keyconjurer-darwin bin/keyconjurer-darwin-amd64 bin/keyconjurer-darwin-arm64 bin/keyconjurer-linux bin/keyconjurer-linux-amd64 bin/keyconjurer-linux-arm64 bin/keyconjurer-windows.exe
build: api_build frontend/dist/index.html $(CLI_TARGETS)

go_test:
	go test ./...

## Frontend Build Targets
frontend_test:
	cd frontend && CI=true npm test

frontend/node_modules:
	cd frontend && npm install

frontend/dist/index.html: frontend/node_modules
	VITE_APP_VERSION='$(shell git rev-parse --short HEAD)-$(RELEASE)' cd frontend && npm run-script build

### CLI Build Targets
bin/keyconjurer-linux-arm64 bin/keyconjurer-linux:
	GOOS=linux GOARCH=amd64 BUILD_TARGET=keyconjurer-linux $(MAKE) bin/keyconjurer
	GOOS=linux GOARCH=arm64 BUILD_TARGET=keyconjurer-linux-arm64 $(MAKE) bin/keyconjurer

bin/keyconjurer-linux-amd64: bin/keyconjurer-linux
	cp bin/keyconjurer-linux bin/keyconjurer-linux-amd64

bin/keyconjurer-darwin-arm64 bin/keyconjurer-darwin:
	GOOS=darwin GOARCH=amd64 BUILD_TARGET=keyconjurer-darwin $(MAKE) bin/keyconjurer
	GOOS=darwin GOARCH=arm64 BUILD_TARGET=keyconjurer-darwin-arm64 $(MAKE) bin/keyconjurer

bin/keyconjurer-darwin-amd64: bin/keyconjurer-darwin
	cp bin/keyconjurer-darwin bin/keyconjurer-darwin-amd64

bin/keyconjurer-windows.exe:
	GOOS=windows GOARCH=amd64 BUILD_TARGET=keyconjurer-windows.exe $(MAKE) bin/keyconjurer

bin/keyconjurer: bin/
	@test $${CLIENT_ID?is not set}
	@test $${OIDC_DOMAIN?is not set}
	@test $${SERVER_ADDRESS?is not set}
	@CGO_ENABLED=0 go build \
		-ldflags "\
			-s -w \
			-X github.com/riotgames/key-conjurer/command.Version=$(shell git rev-parse --short HEAD)-$(RELEASE) \
			-X github.com/riotgames/key-conjurer/command.ClientID=$(CLIENT_ID) \
			-X github.com/riotgames/key-conjurer/command.OIDCDomain=$(OIDC_DOMAIN) \
			-X github.com/riotgames/key-conjurer/command.BuildTimestamp='$(TIMESTAMP)' \
			-X github.com/riotgames/key-conjurer/command.ServerAddress=$(SERVER_ADDRESS)" \
		-o bin/$(BUILD_TARGET)

bin/:
	mkdir -p bin/

## Upload Targets
upload: cli_upload frontend_upload

cli_upload: $(CLI_TARGETS)
	@test $${S3_FRONTEND_BUCKET_NAME?is not set}
	@test $${RELEASE?is not set}
	cd bin/ && \
	aws s3 cp . s3://$(S3_FRONTEND_BUCKET_NAME)-$(RELEASE) --exclude "*" --include "keyconjurer*" --recursive

frontend_upload: frontend/dist/index.html
	@test $${S3_FRONTEND_BUCKET_NAME?is not set}
	@test $${RELEASE?is not set}
	cd frontend/dist && \
	aws s3 cp . s3://$(S3_FRONTEND_BUCKET_NAME)-$(RELEASE) --include "*" --recursive
