RELEASE ?= dev
VERSION ?= $(shell git rev-parse --short HEAD)

all: build

list_applications_v2.zip:
	GOOS=linux GOARCH=amd64 go build \
		-tags lambda.norpc \
		-o bootstrap serverless_functions/$(subst .zip,,$@)/main.go
	zip $@ bootstrap
	rm bootstrap

../builds/$(RELEASE)/list_applications_v2.zip: list_applications_v2.zip
	mkdir -p ../builds/$(RELEASE)
	mv $^ $@

build: ../builds/$(RELEASE)/list_applications_v2.zip

test:
	go test ./...

clean:
	-rm -f -r ../builds
