ifndef WSNAME
$(error WSNAME is not set)
endif

ifndef AUTHENTICATOR
$(error AUTHENTICATOR is not set)
endif

ifndef MFA
$(error MFA is not set)
endif

ifndef LOGSTASH_ENDPOINT
$(error LOGSTASH_ENDPOINT is not set)
endif

ifndef S3_TF_BUCKET_NAME
$(error S3_TF_BUCKET_NAME is not set)
endif

version = $(shell git rev-parse --short HEAD)
ldflags = "-X keyconjurer-lambda/consts.Version=$(version)-${WSNAME} -X keyconjurer-lambda/consts.AuthenticatorSelect=$(AUTHENTICATOR) -X keyconjurer-lambda/consts.MFASelect=$(MFA) -X keyconjurer-lambda/consts.LogstashEndpoint=$(LOGSTASH_ENDPOINT)"

deploy:
	make build \
	&& make zip \
	&& make upload

build:
	cd ./lambdas/get_user_data && \
	GOOS=linux go build -ldflags $(ldflags) -o ../../builds/$(WSNAME)/get_user_data;
	cd ./lambdas/get_aws_creds && \
	GOOS=linux go build -ldflags $(ldflags) -o ../../builds/$(WSNAME)/get_aws_creds;

zip:
	cd ./builds/$(WSNAME) && \
	zip get_user_data.zip get_user_data \
	&& zip get_aws_creds.zip get_aws_creds

upload:
	mkdir -p ./builds/$(WSNAME) \
	&& cd ./builds/$(WSNAME) \
	&& aws s3 cp . s3://$(S3_TF_BUCKET_NAME)/$(WSNAME) --exclude "*" --include "*.zip" --recursive

test:
	go test -v ./...
