RELEASE ?= dev
VERSION ?= $(shell git rev-parse --short HEAD)

builds/$(RELEASE)/:
	mkdir -p $@

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

builds/$(RELEASE)/list_applications.zip: builds/$(RELEASE)/
# A temporary destination is used because we don't want multiple targets run at the same time to conflict - they all have to be named 'bootstrap'
	TMP_DST=$$(mktemp -d) ;\
	GOOS=linux GOARCH=amd64 go build \
		-tags lambda.norpc \
		-o $$TMP_DST/bootstrap lambda/$(subst .zip,,$(notdir $@))/main.go && \
	(cd $$TMP_DST && zip - bootstrap) > $@
