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

.PHONY: all linux  darwin windows build dir test build_requirements


all: linux darwin windows


linux: export GOOS=linux
linux: export BUILD_TARGET=keyconjurer-linux
linux: dir
	make build


darwin: export GOOS=darwin
darwin: export BUILD_TARGET=keyconjurer-darwin
darwin: dir
	make build


windows: export GOOS=windows
windows: export BUILD_TARGET=keyconjurer-windows.exe
windows: dir
	make build

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
