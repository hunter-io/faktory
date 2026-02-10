NAME=faktory
VERSION=0.15.0

# when fixing packaging bugs but not changing the binary, we increment ITERATION
ITERATION=1

DOCKER_IMAGE=$(NAME)
DOCKER_TEST_IMAGE=$(NAME)-test

.DEFAULT_GOAL := help

all: test

test: ## Run test suite in Docker
	docker build -t $(DOCKER_TEST_IMAGE) -f Dockerfile.test .
	docker run --rm $(DOCKER_TEST_IMAGE)

build: ## Build production Docker image
	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .

run: build ## Run Faktory in Docker
	docker run --rm -it -e "FAKTORY_SKIP_PASSWORD=true" \
		-v faktory-data:/var/lib/faktory \
		-p 127.0.0.1:7419:7419 \
		-p 127.0.0.1:7420:7420 \
		$(DOCKER_IMAGE):latest /faktory -e production

fmt: ## Format code in Docker (modifies local files)
	docker run --rm -v "$$(pwd)":/src -w /src golang:1.25-alpine go fmt ./...

generate: ## Generate webui templates in Docker (copies generated files locally)
	docker build -t $(DOCKER_TEST_IMAGE) -f Dockerfile.test .
	container_id=$$(docker create $(DOCKER_TEST_IMAGE)) && \
		for f in $$(docker run --rm $(DOCKER_TEST_IMAGE) sh -c 'ls /app/webui/*.ego.go' 2>/dev/null); do \
			docker cp $$container_id:$$f webui/$$(basename $$f); \
		done && \
		docker rm $$container_id > /dev/null

cover: ## Generate coverage report in Docker
	docker build -t $(DOCKER_TEST_IMAGE) -f Dockerfile.test .
	docker run --rm -v "$$(pwd)":/out $(DOCKER_TEST_IMAGE) \
		sh -c "redis-server --daemonize yes && sleep 1 && go test -cover -coverprofile /out/cover.out github.com/hunter-io/faktory/server"
	@echo "Coverage written to cover.out"

version_check: ## Verify version strings are in sync
	@grep -q $(VERSION) client/faktory.go || (echo VERSIONS OUT OF SYNC && false)

clean: ## Remove generated files
	@rm -f webui/*.ego.go
	@rm -rf tmp
	@rm -f main faktory templates.go cover.out coverage.html

tag: ## Tag the current version
	git tag v$(VERSION)-$(ITERATION) && git push --tags || :

.PHONY: help all clean test build run fmt generate cover version_check tag

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
