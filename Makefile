.PHONY: cloudops-api

all: binaries
binaries: cloudops-api-binary

###############################################################################
# Default value: directory with Makefile
BASE_DIR?=$(dir $(lastword $(MAKEFILE_LIST)))
SOURCE_DIR:=$(abspath $(BASE_DIR))
###############################################################################

CONTAINER_BUILD_ARGS?=
BINARY_IMAGE_NAME?=cloudops-api
BASE_BUILD_CONTAINER?=cloudops-build
REPO_LOCATION?=/go/src/github.com/rantuttl/cloudops


cloudops-api-base:
	@echo "\033[92m\n\nBuilding base build container: $(BASE_BUILD_CONTAINER)\033[0m"
	@echo "-----------------------------------------------------------------"
	-docker rm -f $(BASE_BUILD_CONTAINER)
	docker build $(CONTAINER_BUILD_ARGS) -t $(BASE_BUILD_CONTAINER) .

cloudops-api-base-if:
	@echo "\033[92m\n\nChecking base build container existence: $(BASE_BUILD_CONTAINER)\033[0m"
	@echo "-----------------------------------------------------------------"
	docker images | grep "$(BASE_BUILD_CONTAINER)\s*latest" || $(MAKE) cloudops-api-base

cloudops-api-binary: cloudops-api-base-if
	@echo "\033[92m\n\nBuilding binary: $(BINARY_IMAGE_NAME)\033[0m"
	@echo "-----------------------------------------------------------------"
	-mkdir -p $(SOURCE_DIR)/dist
	docker run --rm -v $(SOURCE_DIR):$(REPO_LOCATION) -w $(REPO_LOCATION) \
		$(BASE_BUILD_CONTAINER) \
		/bin/bash -c 'go build -o dist/$(BINARY_IMAGE_NAME) && chown $(shell id -u):$(shell id -g) -R dist/'
