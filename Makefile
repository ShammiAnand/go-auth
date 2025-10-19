# load env variables from .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

PROJECT_ROOT := $(shell pwd)
BINARY_NAME=go-auth
DOCKER_IMAGE=go-auth:latest

.PHONY: help build test run gen-ent init create-superuser jwks-refresh clean docker-build docker-up docker-down docker-logs dev all

help: ## Display this help message
	@echo "Go-Auth Makefile Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@go mod tidy
	@go mod vendor
	@go build -o ./bin/$(BINARY_NAME) .
	@echo "Build complete: ./bin/$(BINARY_NAME)"

run: build ## Build and run the server
	@echo "Starting server..."
	@./bin/$(BINARY_NAME) server

init: build ## Initialize RBAC (roles & permissions)
	@echo "Initializing RBAC..."
	@./bin/$(BINARY_NAME) init --config ./configs/rbac-config.yaml

create-superuser: build ## Create a super-admin user
	@echo "Creating super-admin..."
	@./bin/$(BINARY_NAME) admin create-superuser \
		--email admin@go-auth.local \
		--first-name Admin \
		--last-name User \
		--password SuperSecure123!

jwks-refresh: build ## Run JWKS key refresh job (24h interval)
	@echo "Starting JWKS refresh job..."
	@./bin/$(BINARY_NAME) jobs jwks-refresh --interval 24h

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

gen-ent: ## Generate Ent code
	@echo "Generating Ent code..."
	@go generate ./ent

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -f ./bin/$(BINARY_NAME)
	@rm -rf vendor/

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

docker-up: ## Start all services with docker-compose
	@echo "Starting services..."
	@docker-compose up -d

docker-down: ## Stop all services
	@echo "Stopping services..."
	@docker-compose down

docker-logs: ## View docker-compose logs
	@docker-compose logs -f

docker-restart: docker-down docker-up ## Restart all services

dev: docker-up ## Start development environment
	@echo ""
	@echo "✅ Development environment started!"
	@echo ""
	@echo "Services:"
	@echo "  PostgreSQL: localhost:5432"
	@echo "  Redis: localhost:6379"
	@echo "  Mailhog SMTP: localhost:1025"
	@echo "  Mailhog UI: http://localhost:8025"
	@echo ""
	@echo "Next steps:"
	@echo "  1. make init           # Initialize RBAC"
	@echo "  2. make create-superuser  # Create admin user"
	@echo "  3. make run            # Start API server"
	@echo ""

all: clean build test ## Clean, build, and test

