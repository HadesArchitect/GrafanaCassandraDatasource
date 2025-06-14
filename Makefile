.PHONY: help
help: ## This help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
.DEFAULT_GOAL := help

OS=linux
ARCH=amd64
GOLANG=1.24.3
NODE=20

install: fe-deps be-deps ## install datasource dependencies (front and back)

build: clean fe-build be-build ## Build the whole datasource

frontend: fe-deps fe-build ## Install frontend dependencies and build frontend

backend: be-deps be-build ## Install backend dependencies and build backend

test: be-test ## Run unit tests

clean: ## Cleans destination folder
	rm -rf ./dist/*

start: ## Launches dev environment
	docker-compose up -d

stop: ## Stops dev environment
	docker-compose stop

fe-deps: ## Install frontend dependencies
	docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:${NODE}-alpine yarn install

fe-build: ## Build frontend
	docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:${NODE}-alpine yarn build

fe-watch: ## Watch frontend
	docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:${NODE}-alpine yarn watch

be-deps: ## Install backend dependencies
	docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp/backend golang:${GOLANG}-alpine go mod vendor

be-tidy: ## Go mod tidy
	docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp/backend golang:${GOLANG}-alpine go mod tidy

be-build: ## Build backend (Builds linux-amd64 version by deafult. Run with args to adjust target (make be-build OS=windows ARCH=arm64))
	docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp/backend -e CGO_ENABLED=0 -e GOOS=$(OS) -e GOARCH=$(ARCH) golang:${GOLANG}-alpine go build -buildvcs=false -o ../dist/cassandra-plugin_$(OS)_$(ARCH) .

be-test: ## Run backend unit tests
	docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp/backend golang:${GOLANG}-alpine go test ./...
# backend tests in CI required `-vet=off`
# docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp/backend golang:1-alpine go test -buildvcs=false -v -vet=off ./...

update-versions: ## Update version in plugin.json to match package.json
	docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:${NODE}-alpine node scripts/update-versions.js