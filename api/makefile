.PHONY: build
SERVERLESS_BUILDS = $(shell ls -d serverless_functions/**/*)

ifndef SERVERLESS_BUILDS
$(error SERVERLESS_BUILDS not set)
else
$(info ${SERVERLESS_BUILDS} will be built)
endif

ifndef TF_WORKSPACE
$(error TF_WORKSPACE is not set)
endif

version = $(shell git rev-parse --short HEAD)
ldflags = "-X github.com/riotgames/key-conjurer/api/consts.Version=$(version)-${TF_WORKSPACE}"

deploy:
	make build \
	&& make zip \
	&& make upload

clean:
	rm -rf ../builds

build: $(SERVERLESS_BUILDS)

$(SERVERLESS_BUILDS): build_dir
	$(info building ${@})
	GOOS=linux go build -ldflags $(ldflags) -o ../builds/$(TF_WORKSPACE)/$(subst serverless_functions/,,$@) $@/*.go;

build_dir:
	mkdir -p builds

test:
	go test ./...
