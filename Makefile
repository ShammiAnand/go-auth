ifneq (,$(wildcard ./.env))
    include .env
    export
endif

PROJECT_ROOT := $(shell pwd)
BINARY_NAME=go-auth
DOCKER_IMAGE=go-auth:latest

.PHONY: build test run gen-ent init create-superuser jwks-refresh clean docker-build docker-up docker-down


build:
	@go mod tidy
	@go mod vendor
	@go build -o ./bin/$(BINARY_NAME) .

run: build
	@./bin/$(BINARY_NAME) server

init: build
	@./bin/$(BINARY_NAME) init --config ./configs/rbac-config.yaml

create-superuser: build
	@./bin/$(BINARY_NAME) admin create-superuser \
		--email test@go-auth.local \
		--first-name Admin \
		--last-name User \
		--password test123

jwks-refresh: build
	@./bin/$(BINARY_NAME) jobs jwks-refresh --interval 12h

test:
	@go test -v ./...

gen-ent:
	@go generate ./ent

clean:
	@rm -f ./bin/$(BINARY_NAME)
	@rm -rf vendor/

docker-build:
	@docker build -t $(DOCKER_IMAGE) .

docker-up:
	@docker-compose up -d

docker-down:
	@docker-compose down


