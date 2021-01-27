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
	&& GOOS=linux go build -ldflags "-X github.com/riotgames/key-conjurer/cli/main.Version=$(version)-$(TF_WORKSPACE) -X github.com/riotgames/key-conjurer/cli/main.ClientName=keyconjurer-linux -X github.com/riotgames/key-conjurer/cli/cmd.defaultHost=$(API_URL) -X github.com/riotgames/key-conjurer/cli/main.DownloadURL=$(FRONTEND_URL)" -o ../builds/$(TF_WORKSPACE)/cli/main-linux

darwin:
	export os=darwin \
	&& GOOS=darwin go build -ldflags "-X github.com/riotgames/key-conjurer/cli/main.Version=$(version)-$(TF_WORKSPACE) -X github.com/riotgames/key-conjurer/cli/main.ClientName=keyconjurer-darwin -X github.com/riotgames/key-conjurer/cli/cmd.defaultHost=$(API_URL) -X github.com/riotgames/key-conjurer/cli/main.DownloadURL=$(FRONTEND_URL)" -o ../builds/$(TF_WORKSPACE)/cli/main-darwin

windows:
	export os=windows \
	&& GOOS=windows go build -ldflags "-X github.com/riotgames/key-conjurer/cli/main.Version=$(version)-$(TF_WORKSPACE) -X github.com/riotgames/key-conjurer/cli/main.ClientName=keyconjurer-windows -X github.com/riotgames/key-conjurer/cli/cmd.defaultHost=$(API_URL) -X github.com/riotgames/key-conjurer/cli/main.DownloadURL=$(FRONTEND_URL)" -o ../builds/$(TF_WORKSPACE)/cli/main-windows.exe

dir:
	mkdir -p ../builds/$(TF_WORKSPACE)/cli

test: dir
	mkdir -p ~/.aws \
	&& touch ~/.aws/config \
	&& touch ~/.aws/credentials \
	&& touch ~/.keyconjurerrc \
	&& go test -v ./...
