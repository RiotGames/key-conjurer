.DEFAULT_GOAL = all

version = $(shell git rev-parse --short HEAD)

build_date =  $(shell date -u +'%Y-%m-%d')
build_time =  $(shell date -u +'%H:%M:%S')
build_timezone = UTC

ifndef API_URL
$(error API_URL is not set)
endif
ifndef FRONTEND_URL
$(err FRONTEND_URL is not set)
endif
ifndef TF_WORKSPACE
$(error TF_WORKSPACE is not set)
endif

.PHONY: all build dir test build_requirements

all: \
	../builds/$(TF_WORKSPACE)/cli/keyconjurer-linux \
	../builds/$(TF_WORKSPACE)/cli/keyconjurer-darwin-amd64 \
	../builds/$(TF_WORKSPACE)/cli/keyconjurer-darwin-arm64 \
	../builds/$(TF_WORKSPACE)/cli/keyconjurer-windows.exe

../builds/$(TF_WORKSPACE)/cli/keyconjurer-linux:
	GOOS=linux GOARCH=amd64 BUILD_TARGET=keyconjurer-linux $(MAKE) build

../builds/$(TF_WORKSPACE)/cli/keyconjurer-darwin:
../builds/$(TF_WORKSPACE)/cli/keyconjurer-darwin-amd64:
../builds/$(TF_WORKSPACE)/cli/keyconjurer-darwin-arm64:
	GOOS=linux GOARCH=amd64 BUILD_TARGET=keyconjurer-darwin $(MAKE) build
	GOOS=linux GOARCH=arm64 BUILD_TARGET=keyconjurer-darwin-arm64 $(MAKE) build
	cp ../builds/$(TF_WORKSPACE)/cli/keyconjurer-darwin ../builds/$(TF_WORKSPACE)/cli/keyconjurer-darwin-amd64

../builds/$(TF_WORKSPACE)/cli/keyconjurer-windows.exe:
	GOOS=windows GOARCH=amd64 BUILD_TARGET=keyconjurer-windows.exe $(MAKE) build

dir:
	mkdir -p ../builds/$(TF_WORKSPACE)/cli

build: dir
	go build \
		-ldflags "-X main.Version=$(version)-$(TF_WORKSPACE) -X main.ClientName=keyconjurer-$(GOOS) -X main.defaultHost=$(API_URL) -X main.DownloadURL=$(FRONTEND_URL) -X main.BuildDate=$(build_date) -X main.BuildTime=$(build_time) -X main.BuildTimeZone=$(build_timezone)" \
		-o ../builds/$(TF_WORKSPACE)/cli/$(BUILD_TARGET)

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
