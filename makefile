RELEASE ?= dev

build:
	make cli_build \
	&& make api_build \
	&& make frontend_build

api_build:
	mkdir -p builds/$(RELEASE)/aws
	cd api \
	&& $(MAKE) -f makefile build

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
