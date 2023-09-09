RELEASE ?= dev

version = $(shell git rev-parse --short HEAD)
ldflags = "-X github.com/riotgames/key-conjurer/api/consts.Version=$(version)-${RELEASE}"

all: build

list_applications_v2.zip:
	GOOS=linux GOARCH=amd64 go build \
		-ldflags $(ldflags) \
		-tags lambda.norpc \
		-o bootstrap serverless_functions/$(subst .zip,,$@)/main.go
	zip $@ bootstrap
	rm bootstrap

list_providers.zip get_cloud_creds.zip get_user_data.zip:
	GOOS=linux GOARCH=amd64 go build -ldflags $(ldflags) -o $(subst .zip,,$@) serverless_functions/cloud/$(subst .zip,,$@)/main.go
	zip $@ $(subst .zip,,$@)
	rm $(subst .zip,,$@)

../builds/$(RELEASE)/cloud/get_cloud_creds.zip: get_cloud_creds.zip
	mkdir -p ../builds/$(RELEASE)/cloud
	mv $^ $@
../builds/$(RELEASE)/cloud/get_user_data.zip: get_user_data.zip
	mkdir -p ../builds/$(RELEASE)/cloud
	mv $^ $@
../builds/$(RELEASE)/cloud/list_providers.zip: list_providers.zip
	mkdir -p ../builds/$(RELEASE)/cloud
	mv $^ $@
../builds/$(RELEASE)/list_applications_v2.zip: list_applications_v2.zip
	mkdir -p ../builds/$(RELEASE)
	mv $^ $@

build: ../builds/$(RELEASE)/cloud/get_cloud_creds.zip ../builds/$(RELEASE)/cloud/get_user_data.zip ../builds/$(RELEASE)/cloud/list_providers.zip ../builds/$(RELEASE)/list_applications_v2.zip

test:
	go test ./...

clean:
	-rm -f *.zip list_providers get_user_data get_cloud_creds
	-rm -f -r ../builds

