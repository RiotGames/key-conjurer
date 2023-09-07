.DEFAULT_GOAL = all

version = $(shell git rev-parse --short HEAD)

build_date =  $(shell date -u +'%Y-%m-%d')
build_time =  $(shell date -u +'%H:%M:%S')
build_timezone = UTC

RELEASE ?= dev

ifndef CLIENT_ID
$(error CLIENT_ID is not set)
endif
ifndef OIDC_DOMAIN
$(error OIDC_DOMAIN is not set)
endif

.PHONY: all build dir test build_requirements

all: \
	../builds/$(RELEASE)/cli/keyconjurer-linux-arm64 \
	../builds/$(RELEASE)/cli/keyconjurer-linux-amd64 \
	../builds/$(RELEASE)/cli/keyconjurer-darwin-amd64 \
	../builds/$(RELEASE)/cli/keyconjurer-darwin-arm64 \
	../builds/$(RELEASE)/cli/keyconjurer-windows.exe

../builds/$(RELEASE)/cli/keyconjurer-linux:
../builds/$(RELEASE)/cli/keyconjurer-linux-arm64:
../builds/$(RELEASE)/cli/keyconjurer-linux-amd64:
	GOOS=linux GOARCH=amd64 BUILD_TARGET=keyconjurer-linux $(MAKE) build
	GOOS=linux GOARCH=arm64 BUILD_TARGET=keyconjurer-linux-arm64 $(MAKE) build
	cp ../builds/$(RELEASE)/cli/keyconjurer-linux ../builds/$(RELEASE)/cli/keyconjurer-linux-amd64

../builds/$(RELEASE)/cli/keyconjurer-darwin:
../builds/$(RELEASE)/cli/keyconjurer-darwin-amd64:
../builds/$(RELEASE)/cli/keyconjurer-darwin-arm64:
	GOOS=darwin GOARCH=amd64 BUILD_TARGET=keyconjurer-darwin $(MAKE) build
	GOOS=darwin GOARCH=arm64 BUILD_TARGET=keyconjurer-darwin-arm64 $(MAKE) build
	cp ../builds/$(RELEASE)/cli/keyconjurer-darwin ../builds/$(RELEASE)/cli/keyconjurer-darwin-amd64

../builds/$(RELEASE)/cli/keyconjurer-windows.exe:
	GOOS=windows GOARCH=amd64 BUILD_TARGET=keyconjurer-windows.exe $(MAKE) build

dir:
	mkdir -p ../builds/$(RELEASE)/cli

build: dir
	go build \
		-ldflags "-X main.Version=$(version)-$(RELEASE) -X main.ClientName=keyconjurer-$(GOOS) -X main.ClientID=$(CLIENT_ID) -X main.OIDCDomain=$(OIDC_DOMAIN) -X main.BuildDate=$(build_date) -X main.BuildTime=$(build_time) -X main.BuildTimeZone=$(build_timezone)" \
		-o ../builds/$(RELEASE)/cli/$(BUILD_TARGET)

test: dir
	mkdir -p ~/.aws
	touch ~/.aws/config
	touch ~/.aws/credentials
	touch ~/.keyconjurerrc
	go test -v ./...

build_requirements:
	go version
	gcc --version
	make --version
