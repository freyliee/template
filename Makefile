SHELL := /bin/bash

.PHONY: run migrate migrate-down generate build test

run:
	set -a; source .env && set +a && go run ./cmd/trip-service

migrate:
	set -a; source .env && set +a && go tool goose -dir migrations postgres "$$DATABASE_URL" up

migrate-down:
	set -a; source .env; set +a; go tool goose -dir migrations postgres "$$DATABASE_URL" down

generate:
	go tool oapi-codegen \
		-config oapi-codegen.yaml \
		contracts/openapi/trip-service.openapi.yaml

build:
	go build ./...

test:
	go test -race ./...