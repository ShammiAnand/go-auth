# load env variables from .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

PROJECT_ROOT := $(shell pwd)

.PHONY: build test run gen-grpc gen-http gen-ent gen swagger-ui stop-swagger-ui start

compose-up:
	@echo "Starting Postgres and Redis ..."
	docker-compose up -d
	@echo "DONE"

compose-down:
	@echo "Stopping Postgres and Redis ..."
	docker-compose down
	@echo "DONE"

build:
	@go build -o bin/go-auth cmd/main.go

test:
	@go test -v ./...


run: build
	@./bin/go-auth

migrate: build
	@./bin/go-auth --migrate true

gen-ent:
	@echo "Generating Ent code..."
	@go generate ./ent

