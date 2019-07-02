.DEFAULT_GOAL = all

version = $(shell git rev-parse --short HEAD)

ifndef API_URL
$(error API_URL is not set)
endif
ifndef FRONTEND_URL
$(err FRONTEND_URL is not set)
endif
ifndef WSNAME
$(error WSNAME is not set)
endif

all: linux darwin windows

upload:
	cd ./builds/$(WSNAME) \
	&& aws s3 cp . s3://keyconjurer-frontend-$(WSNAME) --exclude "*" --include "keyconjurer*" --recursive

linux: dir
	export os=linux \
	&& GOOS=linux go build -ldflags "-X keyconjurer-cli/keyconjurer.Version=$(version) -X keyconjurer-cli/keyconjurer.Client=keyconjurer-linux -X keyconjurer-cli/keyconjurer.ProdAPI=$(API_URL) -X keyconjurer-cli/keyconjurer.DevAPI=$(API_URL) -X keyconjurer-cli/keyconjurer.DownloadURL=$(FRONTEND_URL)" -o ./builds/$(WSNAME)/keyconjurer-linux

darwin:
	export os=darwin \
	&& GOOS=darwin go build -ldflags "-X keyconjurer-cli/keyconjurer.Version=$(version) -X keyconjurer-cli/keyconjurer.Client=keyconjurer-darwin -X keyconjurer-cli/keyconjurer.ProdAPI=$(API_URL) -X keyconjurer-cli/keyconjurer.DevAPI=$(API_URL) -X keyconjurer-cli/keyconjurer.DownloadURL=$(FRONTEND_URL)" -o ./builds/$(WSNAME)/keyconjurer-darwin

windows:
	export os=windows \
	&& GOOS=windows go build -ldflags "-X keyconjurer-cli/keyconjurer.Version=$(version) -X keyconjurer-cli/keyconjurer.Client=keyconjurer-windows -X keyconjurer-cli/keyconjurer.ProdAPI=$(API_URL) -X keyconjurer-cli/keyconjurer.DevAPI=$(API_URL) -X keyconjurer-cli/keyconjurer.DownloadURL=$(FRONTEND_URL)" -o ./builds/$(WSNAME)/keyconjurer-windows.exe

dir:
	mkdir -p ./builds

test: dir 
	mkdir -p ~/.aws \
	&& touch ~/.aws/config \
	&& touch ~/.aws/credentials \
	&& touch ~/.keyconjurerrc \
	&& go test -v ./...
