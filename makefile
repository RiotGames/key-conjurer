RELEASE ?= dev

build:
	make cli_build \
	&& make api_build \
	&& make frontend_build

api_build: builds/$(RELEASE)/list_applications.zip

frontend_build:
	mkdir -p builds/$(RELEASE)/frontend
	cd frontend \
	&& $(MAKE) -f makefile build

cli_build:
	mkdir -p builds/$(RELEASE)/cli
	cd cli \
	&& $(MAKE) -f makefile all

frontend_file_reset:
	cd frontend \
	&& $(MAKE) -f makefile reset_files

reset_files: frontend_file_reset

deploy: deploy_aws

deploy_aws:
	cd terraform \
	&& $(MAKE) -f makefile deploy

plan_aws:
	cd terraform \
	&& $(MAKE) -f makefile plan_deploy

RELEASE ?= dev
VERSION ?= $(shell git rev-parse --short HEAD)

list_applications.zip:
	GOOS=linux GOARCH=amd64 go build \
		-tags lambda.norpc \
		-o bootstrap lambda/$(subst .zip,,$@)/main.go
	zip $@ bootstrap
	rm bootstrap

builds/$(RELEASE)/list_applications.zip: list_applications.zip
	mkdir -p builds/$(RELEASE)
	mv $^ $@
