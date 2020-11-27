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
	&& GOOS=linux go build -ldflags "-X github.com/riotgames/key-conjurer/cli/keyconjurer.Version=$(version)-$(TF_WORKSPACE) -X github.com/riotgames/key-conjurer/cli/keyconjurer.ClientName=keyconjurer-linux -X github.com/riotgames/key-conjurer/cli/cmd.defaultHostname=$(API_URL) -X github.com/riotgames/key-conjurer/cli/keyconjurer.DownloadURL=$(FRONTEND_URL)" -o ../builds/$(TF_WORKSPACE)/cli/keyconjurer-linux

darwin:
	export os=darwin \
	&& GOOS=darwin go build -ldflags "-X github.com/riotgames/key-conjurer/cli/keyconjurer.Version=$(version)-$(TF_WORKSPACE) -X github.com/riotgames/key-conjurer/cli/keyconjurer.ClientName=keyconjurer-darwin -X github.com/riotgames/key-conjurer/cli/cmd.defaultHostname=$(API_URL) -X github.com/riotgames/key-conjurer/cli/keyconjurer.DownloadURL=$(FRONTEND_URL)" -o ../builds/$(TF_WORKSPACE)/cli/keyconjurer-darwin

windows:
	export os=windows \
	&& GOOS=windows go build -ldflags "-X github.com/riotgames/key-conjurer/cli/keyconjurer.Version=$(version)-$(TF_WORKSPACE) -X github.com/riotgames/key-conjurer/cli/keyconjurer.ClientName=keyconjurer-windows -X github.com/riotgames/key-conjurer/cli/cmd.defaultHostname=$(API_URL) -X github.com/riotgames/key-conjurer/cli/keyconjurer.DownloadURL=$(FRONTEND_URL)" -o ../builds/$(TF_WORKSPACE)/cli/keyconjurer-windows.exe

dir:
	mkdir -p ../builds/$(TF_WORKSPACE)/cli

test: dir
	mkdir -p ~/.aws \
	&& touch ~/.aws/config \
	&& touch ~/.aws/credentials \
	&& touch ~/.keyconjurerrc \
	&& go test -v ./...
