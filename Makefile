.PHONY: help
help: ## This help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
.DEFAULT_GOAL := help

build: frontend backend ## Build the whole datasource

frontend: fe-deps fe-build ## Install frontend dependencies and build frontend

backend: be-deps be-build ## Install backend dependencies and build backend

fe-deps: ## Install frontend dependencies
	docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:16-alpine yarn install

fe-build: ## Build frontend
	docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:16-alpine yarn build

be-deps: ## Install backend dependencies
	docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp/backend golang:1-alpine go mod vendor

be-build: ## Build backend (builds linux-amd64 version)
	docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp/backend golang:1-alpine go build -o ../dist/cassandra-plugin_linux_amd64 .


