REPO = saferwall

docker-build:
	@docker build $(ARGS) -t $(REPO)/$(IMG) \
		-f $(DOCKER_FILE) $(DOCKER_DIR)

docker-build-nc:
	@docker build ${ARGS} --no-cache -t $(REPO)/$(IMG) \
		-f $(DOCKER_FILE) $(DOCKER_DIR)

docker-run:
	docker run -d -p 50051:50051 $(REPO)/$(IMG)

docker-up: build run

docker-stop:
	docker stop $(IMG); docker rm $(REPO)/$(IMG)

docker-release: docker-repo-login docker-build-nc docker-publish

docker-publish: docker-repo-login docker-publish-latest docker-publish-version

docker-publish-latest: docker-tag-latest
	@echo 'publish latest to $(REPO)/$(IMG)'
	docker push $(REPO)/$(IMG):latest

docker-publish-version: docker-tag-version
	@echo 'publish $(VERSION) to $(IMG)'
	docker push $(REPO)/$(IMG):$(VERSION)

docker-tag: docker-tag-latest docker-tag-version

docker-tag-latest:
	@echo 'create tag latest'
	docker tag $(REPO)/$(IMG) $(REPO)/$(IMG):latest

docker-tag-version:
	@echo 'create tag $(VERSION)'
	docker tag $(REPO)/$(IMG) $(REPO)/$(IMG):$(VERSION)

docker-repo-login:
	@echo '$(DOCKER_HUB_PWD)' | docker login -u $(DOCKER_HUB_USR) --password-stdin
