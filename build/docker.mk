REPO = saferwall

docker-build: ## Build the container
	@docker build $(ARGS) -t $(REPO)/$(IMG) \
		-f $(DOCKER_FILE) $(DOCKER_DIR)

docker-build-nc: ## Build the container without caching
	@docker build ${ARGS} --no-cache -t $(REPO)/$(IMG) \
		-f $(DOCKER_FILE) $(DOCKER_DIR)

docker-run: ## Run container on port configured in `config.env`
	docker run -d -p 50051:50051 $(REPO)/$(IMG)

docker-up: build run ## Run container

docker-stop: ## Stop and remove a running container
	docker stop $(IMG); docker rm $(REPO)/$(IMG)

docker-release: docker-repo-login docker-build-nc docker-publish ## Make a release by building and publishing the `{version}` and `latest` tagged containers to ECR

docker-publish: docker-repo-login docker-publish-latest docker-publish-version ## Publish the `{version}` and `latest` tagged containers to ECR

docker-publish-latest: docker-tag-latest ## Publish the `latest` taged container to ECR
	@echo 'publish latest to $(REPO)/$(IMG)'
	docker push $(REPO)/$(IMG):latest

docker-publish-version: docker-tag-version ## Publish the `{version}` taged container to ECR
	@echo 'publish $(VERSION) to $(IMG)'
	docker push $(REPO)/$(IMG):$(VERSION)

docker-tag: docker-tag-latest docker-tag-version ## Generate container tags for the `{version}` and `latest` tags

docker-tag-latest: 	## Generate container `{version}` tag
	@echo 'create tag latest'
	docker tag $(REPO)/$(IMG) $(REPO)/$(IMG):latest

docker-tag-version: 	## Generate container `latest` tag
	@echo 'create tag $(VERSION)'
	docker tag $(REPO)/$(IMG) $(REPO)/$(IMG):$(VERSION)

docker-repo-login: 	## Login to Docker Hub
	@echo '$(DOCKER_HUB_PWD)' | docker login --username=$(DOCKER_HUB_USR) --password-stdin
