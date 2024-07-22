# Needed SHELL since I'm using zsh
SHELL := /bin/bash

# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[0-9a-zA-Z\/_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
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
include .env
-include private.env
export

# Include our internals makefiles.
include build/docker.mk

init: compose/pull install/swag couchbase/init	## Init project.

install/swag:	## Install Swag
	go install github.com/swaggo/swag/cmd/swag@latest

compose/pull:	## Docker compose pull.
	docker compose pull

compose/up:	## Start docker-compose (args: SVC: name of the service to exclude)
	@echo "${GREEN} [*] =============== Docker Compose Up =============== ${RESET}"
	docker compose config --services | grep -v '${SVC}' | xargs docker compose up

compose/up/min:	## Start docker-compose (args: SVC: name of the service to exclude)
	@echo "${YELLOW} [*] =============== Docker Compose Up Minimum =============== ${RESET}"
	docker compose config --services | grep -v 'clamav\|sandbox\|meta\|orchestrator\|postprocessor\|aggregator\|pe\|ui' \
		| xargs docker compose up

docker/build:	## Build in a docker container.
	@echo "${GREEN} [*] =============== Docker Build  =============== ${RESET}"
	make docker-build IMG=$(DOCKER_HUB_IMG) DOCKER_FILE=Dockerfile DOCKER_DIR=. ;
	@EXIT_CODE=$$?
	@if test $$EXIT_CODE ! 0; then \
		make docker-build IMG=$(DOCKER_HUB_IMG) DOCKER_FILE=Dockerfile DOCKER_DIR=. ; \
	fi

docker/release:	## Build and release in a docker container.
	@echo "${GREEN} [*] =============== Docker Build and Release  =============== ${RESET}"
	make docker-release IMG=$(DOCKER_HUB_IMG) VERSION=$(SAFERWALL_VER) \
		DOCKER_FILE=Dockerfile DOCKER_DIR=. ;
	@EXIT_CODE=$$?
	@if test $$EXIT_CODE ! 0; then \
		make docker-release IMG=$(DOCKER_HUB_IMG) VERSION=$(SAFERWALL_VER) \
			DOCKER_FILE=Dockerfile DOCKER_DIR=. ; \
	fi

couchbase/start:	## Start Couchbase Server docker container.
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

couchbase/init:	## Init couchbase database by creating the cluster and required buckets.
	# Init the cluster.
	echo "${GREEN} [*] =============== Creating Cluster =============== ${RESET}"
	docker compose start couchbase
	docker exec $(COUCHBASE_CONTAINER_NAME) \
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
		docker exec $(COUCHBASE_CONTAINER_NAME) \
			couchbase-cli bucket-create \
			--cluster localhost \
			--username $(COUCHBASE_ADMIN_USER) \
			--password $(COUCHBASE_ADMIN_PWD) \
			--bucket sfw \
			--bucket-type couchbase \
			--bucket-ramsize 128 \
			--max-ttl 500000000 \
			--enable-flush 0 ; \
	done

generate/doc:	## Generate OpenAPI spec.
	swag init --parseDepth 2 -g cmd/main.go

	old="- '{}': \[\]" && new="- {}" \
		&& sed -i "s|$$old|$$new|g" ${ROOT_DIR}/docs/swagger.yaml
	old='  Bearer: \[\]' && new='- Bearer: []' \
		&& sed -i "s|$$old|$$new|g" ${ROOT_DIR}/docs/swagger.yaml

	tr -d '\n' < ${ROOT_DIR}/docs/swagger.json > ${ROOT_DIR}/docs/swagger-tmp.json
	mv ${ROOT_DIR}/docs/swagger-tmp.json ${ROOT_DIR}/docs/swagger.json

	old='"Bearer": \[\],' && new='"Bearer": \[\]},' \
		&& sed -i "s|$$old|$$new|g" ${ROOT_DIR}/docs/swagger.json
	old='"{}": \[\]                    }' && new="{}" \
		&& sed -i "s|$$old|$$new|g" ${ROOT_DIR}/docs/swagger.json

	old='"{}":' && new="- {}:" \
		&& sed -i "s|$$old|$$new|g" ${ROOT_DIR}/docs/docs.go
