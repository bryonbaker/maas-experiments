# Makefile for External IDP Portal
# Build and push container images using podman

# Image configuration
IMAGE_REGISTRY := quay.io
IMAGE_NAMESPACE := bryonbaker
IMAGE_NAME := external-idp-portal
IMAGE_TAG := latest
IMAGE := $(IMAGE_REGISTRY)/$(IMAGE_NAMESPACE)/$(IMAGE_NAME):$(IMAGE_TAG)

# Container build tool
CONTAINER_TOOL := podman

# Build arguments
BUILD_ARGS := --format docker

.PHONY: help
help: ## Display this help message
	@echo "External IDP Portal - Container Build"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the container image
	@echo "Building container image: $(IMAGE)"
	$(CONTAINER_TOOL) build $(BUILD_ARGS) -t $(IMAGE) -f Containerfile .
	@echo "Image built successfully: $(IMAGE)"

.PHONY: push
push: ## Push the container image to registry
	@echo "Pushing image to registry: $(IMAGE)"
	$(CONTAINER_TOOL) push $(IMAGE)
	@echo "Image pushed successfully to $(IMAGE_REGISTRY)/$(IMAGE_NAMESPACE)/$(IMAGE_NAME):$(IMAGE_TAG)"

.PHONY: build-push
build-push: build push ## Build and push the container image

.PHONY: login
login: ## Login to container registry
	@echo "Logging into $(IMAGE_REGISTRY)"
	@$(CONTAINER_TOOL) login $(IMAGE_REGISTRY)

.PHONY: run
run: ## Run the container locally
	@echo "Running container locally on port 8080"
	$(CONTAINER_TOOL) run --rm -it \
		-p 8080:8080 \
		-v $(HOME)/secrets:/home/appuser/secrets:ro \
		$(IMAGE)

.PHONY: run-dev
run-dev: ## Run locally with Flask dev server
	python3 app.py

.PHONY: shell
shell: ## Open a shell in the container
	@echo "Opening shell in container"
	$(CONTAINER_TOOL) run --rm -it \
		--entrypoint /bin/bash \
		$(IMAGE)

.PHONY: test
test: build ## Build and test the container locally
	@echo "Testing container"
	$(CONTAINER_TOOL) run --rm \
		$(IMAGE) python3 -c "import app; print('App imports successfully')"

.PHONY: clean
clean: ## Remove local container images
	@echo "Removing local image: $(IMAGE)"
	$(CONTAINER_TOOL) rmi $(IMAGE) 2>/dev/null || true
	@echo "Cleanup complete"

.PHONY: tag
tag: ## Tag the image with a custom tag (usage: make tag TAG=v1.0.0)
	@if [ -z "$(TAG)" ]; then \
		echo "Error: TAG variable is required. Usage: make tag TAG=v1.0.0"; \
		exit 1; \
	fi
	@echo "Tagging image $(IMAGE) as $(IMAGE_REGISTRY)/$(IMAGE_NAMESPACE)/$(IMAGE_NAME):$(TAG)"
	$(CONTAINER_TOOL) tag $(IMAGE) $(IMAGE_REGISTRY)/$(IMAGE_NAMESPACE)/$(IMAGE_NAME):$(TAG)
	@echo "Tagged successfully"

.PHONY: push-tag
push-tag: ## Push a specific tag (usage: make push-tag TAG=v1.0.0)
	@if [ -z "$(TAG)" ]; then \
		echo "Error: TAG variable is required. Usage: make push-tag TAG=v1.0.0"; \
		exit 1; \
	fi
	@echo "Pushing $(IMAGE_REGISTRY)/$(IMAGE_NAMESPACE)/$(IMAGE_NAME):$(TAG)"
	$(CONTAINER_TOOL) push $(IMAGE_REGISTRY)/$(IMAGE_NAMESPACE)/$(IMAGE_NAME):$(TAG)
	@echo "Pushed successfully"

.PHONY: info
info: ## Display image information
	@echo "Container Image Information:"
	@echo "  Registry:   $(IMAGE_REGISTRY)"
	@echo "  Namespace:  $(IMAGE_NAMESPACE)"
	@echo "  Name:       $(IMAGE_NAME)"
	@echo "  Tag:        $(IMAGE_TAG)"
	@echo "  Full Image: $(IMAGE)"
	@echo ""
	@echo "Build Tool:   $(CONTAINER_TOOL)"

# Default target
.DEFAULT_GOAL := help
