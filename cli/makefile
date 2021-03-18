.DEFAULT_GOAL = all

version = $(shell git rev-parse --short HEAD)

ifndef API_URL
$(error API_URL is not set)
endif
ifndef FRONTEND_URL
$(err FRONTEND_URL is not set)
endif
ifndef TF_WORKSPACE
$(error TF_WORKSPACE is not set)
endif

all: linux darwin windows

linux: dir
	export os=linux \
	&& GOOS=linux go build -ldflags "-X main.Version=$(version)-$(TF_WORKSPACE) -X main.ClientName=keyconjurer-linux -X main.defaultHost=$(API_URL) -X main.DownloadURL=$(FRONTEND_URL)" -o ../builds/$(TF_WORKSPACE)/cli/keyconjurer-linux

darwin:
	export os=darwin \
	&& GOOS=darwin go build -ldflags "-X main.Version=$(version)-$(TF_WORKSPACE) -X main.ClientName=keyconjurer-darwin -X main.defaultHost=$(API_URL) -X main.DownloadURL=$(FRONTEND_URL)" -o ../builds/$(TF_WORKSPACE)/cli/keyconjurer-darwin

windows:
	export os=windows \
	&& GOOS=windows go build -ldflags "-X main.Version=$(version)-$(TF_WORKSPACE) -X main.ClientName=keyconjurer-windows -X main.defaultHost=$(API_URL) -X main.DownloadURL=$(FRONTEND_URL)" -o ../builds/$(TF_WORKSPACE)/cli/keyconjurer-windows.exe

dir:
	mkdir -p ../builds/$(TF_WORKSPACE)/cli

test: dir
	mkdir -p ~/.aws \
	&& touch ~/.aws/config \
	&& touch ~/.aws/credentials \
	&& touch ~/.keyconjurerrc \
	&& go test -v ./...
