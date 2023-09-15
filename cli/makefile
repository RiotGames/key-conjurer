.DEFAULT_GOAL = all

version = $(shell git rev-parse --short HEAD)

build_date = $(shell date -u +'%Y-%m-%d')
build_time = $(shell date -u +'%H:%M:%S')
build_timezone = UTC

RELEASE ?= dev

.PHONY: all build test build_requirements

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

build:
	@test $${CLIENT_ID?is not set}
	@test $${OIDC_DOMAIN?is not set}
	@test $${SERVER_ADDRESS?is not set}
	@mkdir -p ../builds/$(RELEASE)/cli
	go build \
		-ldflags "-X main.Version=$(version)-$(RELEASE) -X main.ClientID=$(CLIENT_ID) -X main.OIDCDomain=$(OIDC_DOMAIN) -X main.BuildDate=$(build_date) -X main.BuildTime=$(build_time) -X main.BuildTimeZone=$(build_timezone) -X main.ServerAddress=$(SERVER_ADDRESS)" \
		-o ../builds/$(RELEASE)/cli/$(BUILD_TARGET)

test:
	go test -v ./...

build_requirements:
	go version
	gcc --version
	make --version
