ifndef TF_WORKSPACE
$(error TF_WORKSPACE is not set)
endif

ifndef CLOUD_PROVIDER
$(error CLOUD_PROVIDER is not set)
endif

build:
	make cli_build \
	&& make api_build \
	&& make frontend_build
	
api_build:
	mkdir -p builds/$(TF_WORKSPACE)/$(CLOUD_PROVIDER)
	cd api \
	&& $(MAKE) -f makefile build

frontend_build:
	mkdir -p builds/$(TF_WORKSPACE)/frontend
	cd frontend \
	&& $(MAKE) -f makefile build

cli_build:
	mkdir -p builds/$(TF_WORKSPACE)/cli
	cd cli \
	&& $(MAKE) -f makefile all

frontend_file_reset:
	cd frontend \
	&& $(MAKE) -f makefile reset_files

reset_files: frontend_file_reset

deploy:
ifeq ($(CLOUD_PROVIDER),aws)
	make deploy_aws
endif

deploy_aws: 
	cd terraform/aws \
	&& $(MAKE) -f makefile deploy

plan_aws:
	cd terraform/aws \
	&& $(MAKE) -f makefile 
