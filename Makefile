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


dk-run:		## Run the docker container
	sudo docker run -it -p 80:80 --name ml saferwall/$(DOCKER_HUB_IMG)

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

dc-up: 	## Run the docker-compose up.
	docker-compose up

BUCKET_LIST = files users
dc-init-db:		## Init couchbase database
	# Init the cluster.
	echo "${GREEN} [*] =============== Creating Cluster =============== ${RESET}"
	docker-compose exec couchbase \
		couchbase-cli cluster-init \
		--cluster localhost \
		--cluster-username Administrator \
		--cluster-password password \
		--cluster-port 8091 \
		--cluster-name saferwall \
		--services data,index,query \
		--cluster-ramsize 512 \
		--cluster-index-ramsize 256

	echo "${GREEN} [*] =============== Add the node =============== ${RESET}"
	docker-compose exec couchbase \
		couchbase-cli server-add \
		--cluster localhost \
		--username Administrator \
		--password password \
		--server-add localhost \
		--server-add-username Administrator \
		--server-add-password password \
		--services data

	echo "${GREEN} [*] =============== Rebalance the node =============== ${RESET}"
	docker-compose exec couchbase \
		couchbase-cli rebalance \
		--cluster localhost \
		--username Administrator \
		--password password

	# Create require buckets.
	for bucket in $(BUCKET_LIST) ; do \
		echo "${GREEN} [*] =============== Creating $$bucket =============== ${RESET}" ; \
		docker-compose exec couchbase \
			couchbase-cli bucket-create \
			--cluster localhost \
			--username Administrator \
			--password password \
			--bucket $$bucket \
			--bucket-type couchbase \
			--bucket-ramsize 128 \
			--max-ttl 500000000 \
			--durability-min-level persistToMajority \
			--enable-flush 0 ; \
	done
