# load env variables from .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

PROJECT_ROOT := $(shell pwd)

.PHONY: build test run gen-grpc gen-http gen-ent gen swagger-ui stop-swagger-ui start

compose-up:
	@echo "Starting Postgres and Redis ..."
	sudo docker-compose up -d
	@echo "DONE"

compose-down:
	@echo "Stopping Postgres and Redis ..."
	sudo docker-compose down
	@echo "DONE"

build:
	@echo "Building the application..."
	@go build -o bin/go-auth cmd/main.go
	@echo "Build complete."

test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "Tests complete."


run: build
	@echo "Starting the application..."
	@./bin/go-auth


gen-ent:
	@echo "Generating Ent code..."
	@go generate ./ent
	@echo "Ent code generation complete."

start: gen-ent build run
