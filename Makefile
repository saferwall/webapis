# Needed SHELL since I'm using zsh
SHELL := /bin/bash

# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

# Retrieve the root directory of the project.
ROOT_DIR	:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

# Define standard colors
BLACK        := $(shell tput -Txterm setaf 0)
RED          := $(shell tput -Txterm setaf 1)
GREEN        := $(shell tput -Txterm setaf 2)
YELLOW       := $(shell tput -Txterm setaf 3)
LIGHTPURPLE  := $(shell tput -Txterm setaf 4)
PURPLE       := $(shell tput -Txterm setaf 5)
BLUE         := $(shell tput -Txterm setaf 6)
WHITE        := $(shell tput -Txterm setaf 7)

RESET := $(shell tput -Txterm sgr0)

# Our config file.
include $(ROOT_DIR)/.env

# Include our internals makefiles.
include build/docker.mk


dk-build: ## Build frontend in a docker container
	@echo "${GREEN} [*] =============== Build Saferwall ML Pipeline =============== ${RESET}"
	sudo make docker-build IMG=$(DOCKER_HUB_IMG) \
		DOCKER_FILE=Dockerfile DOCKER_DIR=. ;
	@EXIT_CODE=$$?
	@if test $$EXIT_CODE ! 0; then \
		sudo make docker-build IMG=$(DOCKER_HUB_IMG) \
			DOCKER_FILE=Dockerfile DOCKER_DIR=. ; \
	fi

dk-release: ## Build and release frontend in a docker container.
	@echo "${GREEN} [*] =============== Build and Release Frontend =============== ${RESET}"
	sudo make docker-release IMG=$(DOCKER_HUB_IMG) VERSION=$(SAFERWALL_VER) \
		DOCKER_FILE=Dockerfile DOCKER_DIR=. ;
	@EXIT_CODE=$$?
	@if test $$EXIT_CODE ! 0; then \
		sudo make docker-release IMG=$(DOCKER_HUB_IMG) VERSION=$(SAFERWALL_VER) \
			DOCKER_FILE=Dockerfile DOCKER_DIR=. ; \
	fi

COUCHBASE_CONTAINER_NAME = couchbase
COUCHBASE_CONTAINER_VER = couchbase:enterprise-6.6.0
couchbase-start:	## Start Couchbase Server docker container.
	$(eval COUCHBASE_CONTAINER_STATUS := $(shell docker container inspect -f '{{.State.Status}}' $(COUCHBASE_CONTAINER_NAME)))
ifeq ($(COUCHBASE_CONTAINER_STATUS),running)
	@echo "All good, couchabse server is already running."
else ifeq ($(COUCHBASE_CONTAINER_STATUS),exited)
	@echo "Starting Couchbase Server ..."
	docker start $(COUCHBASE_CONTAINER_NAME)
else
	@echo "Creating Couchbase Server ..."
	docker run -d --name $(COUCHBASE_CONTAINER_NAME) -p 8091-8094:8091-8094 -p 11210:11210 $(COUCHBASE_CONTAINER_VER)
endif


COUCHBASE_BUCKETS_LIST = sfw
COUCHBASE_ADMIN_USER = Administrator
COUCHBASE_ADMIN_PWD = password
couchbase-init:		## Init couchbase database by creating the cluster and required buckets.
	# Init the cluster.
	echo "${GREEN} [*] =============== Creating Cluster =============== ${RESET}"
	docker exec couchbase \
		couchbase-cli cluster-init \
		--cluster localhost \
		--cluster-username $(COUCHBASE_ADMIN_USER) \
		--cluster-password $(COUCHBASE_ADMIN_PWD) \
		--cluster-port 8091 \
		--cluster-name saferwall \
		--services data,index,query \
		--cluster-ramsize 512 \
		--cluster-index-ramsize 256

	# Create require buckets.
	for bucket in $(COUCHBASE_BUCKETS_LIST) ; do \
		echo "${GREEN} [*] =============== Creating $$bucket =============== ${RESET}" ; \
		docker exec couchbase \
			couchbase-cli bucket-create \
			--cluster localhost \
			--username Administrator \
			--password password \
			--bucket sfw \
			--bucket-type couchbase \
			--bucket-ramsize 128 \
			--max-ttl 500000000 \
			--enable-flush 0 ; \
	done
