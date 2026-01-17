.PHONY: help
help: ## This help
	@echo "\033[1;34mAvailable targets:\033[0m"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "\033[1;34mConfigurable Variables:\033[0m"
	@echo "\033[36mOS\033[0m                            Target operating system (default: linux)"
	@echo "\033[36mARCH\033[0m                          Target architecture (default: amd64)"
	@echo "\033[36mGOLANG\033[0m                        Golang version for build (default: 1.24.3)"
	@echo "\033[36mNODE\033[0m                          Node version for build (default: 22)"
	@echo ""
	@echo "\033[1;34mUsage Examples:\033[0m"
	@echo "  make build                    # Build with default settings"
	@echo "  make build OS=darwin ARCH=arm64"
	@echo "                                # Build for macOS ARM64"
	@echo "  make be-build OS=windows ARCH=amd64"
	@echo "                                # Build backend for Windows AMD64"
	@echo "  make frontend NODE=20         # Build frontend with Node 20"
.DEFAULT_GOAL := help

# Build configuration variables - override these on the command line
# Usage: make build OS=darwin ARCH=arm64
OS=linux
ARCH=amd64
GOLANG=1.24.3
NODE=22

install: fe-deps be-deps ## install datasource dependencies (front and back)

build: clean fe-build be-build ## Build the whole datasource

frontend: fe-deps fe-build ## Install frontend dependencies and build frontend

backend: be-tidy be-deps be-build ## Install backend dependencies and build backend

test: be-test fe-test ## Run tests

clean: ## Cleans destination folder
	rm -rf ./dist/*

start: ## Launches dev environment
	docker-compose up -d

stop: ## Stops dev environment
	docker-compose stop

fe-deps: ## Install frontend dependencies
	docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:${NODE}-alpine sh -c "corepack enable && yarn install"

fe-build: ## Build frontend
	docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:${NODE}-alpine sh -c "corepack enable && yarn build"

fe-watch: ## Watch frontend
	docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:${NODE}-alpine sh -c "corepack enable && yarn dev"

fe-test: ## Test frontend
	docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:${NODE}-alpine sh -c "corepack enable && yarn test:ci"

be-deps: ## Install backend dependencies
	docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp/pkg golang:${GOLANG}-alpine go mod vendor

be-tidy: ## Go mod tidy
	docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp/pkg golang:${GOLANG}-alpine go mod tidy

be-build: ## Build backend 
	docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp golang:${GOLANG}-alpine sh -c "go build -o mage mage.go && GOOS=$(OS) GOARCH=$(ARCH) ./mage"

be-test: ## Run backend unit tests
	docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp/pkg golang:${GOLANG}-alpine go test ./...
# backend tests in CI required `-vet=off`
# docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp/pkg golang:1-alpine go test -buildvcs=false -v -vet=off ./...

update-versions: ## Update version in plugin.json to match package.json
	docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:${NODE}-alpine sh -c "corepack enable && node scripts/update-versions.js"

sign: ## Sign the plugin before release
	docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds -e GRAFANA_ACCESS_POLICY_TOKEN=${TOKEN} node:${NODE}-alpine sh -c "corepack enable && yarn sign"