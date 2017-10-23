.PHONY: cloudops-api

all: binaries
binaries: cloudops-api-binary
test: cloudops-api-unit-test
deploy: cloudops-api-deploy

###############################################################################
# Default value: directory with Makefile
BASE_DIR?=$(dir $(lastword $(MAKEFILE_LIST)))
SOURCE_DIR:=$(abspath $(BASE_DIR))
###############################################################################

CONTAINER_ARCHIVE?=rtuttle/$(BINARY_IMAGE_NAME):latest
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

cloudops-api-unit-test: cloudops-api-base-if
	@echo "\033[92m\n\nUnit testing $(BINARY_IMAGE_NAME)\033[0m"
	@echo "-----------------------------------------------------------------"
	docker run --rm -v $(SOURCE_DIR):$(REPO_LOCATION) -w $(REPO_LOCATION) \
		$(BASE_BUILD_CONTAINER) \
		/bin/bash -c 'go test ./...'

cloudops-api-deploy: cloudops-api-binary
	@echo "\033[92m\n\nCopying binary $(BINARY_IMAGE_NAME) to archive \033[0m"
	@echo "-----------------------------------------------------------------"
	cp $(SOURCE_DIR)/dist/$(BINARY_IMAGE_NAME) $(SOURCE_DIR)/build
	docker build -t $(CONTAINER_ARCHIVE) $(SOURCE_DIR)/build
	docker push $(CONTAINER_ARCHIVE)
	-rm $(SOURCE_DIR)/build/$(BINARY_IMAGE_NAME)
