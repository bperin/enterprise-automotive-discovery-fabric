.PHONY: dev test build seed eval migrate-up migrate-down sqlc docker-up docker-down help

DB_URL ?= postgres://postgres:postgres@localhost:5432/automotive_discovery?sslmode=disable

help:
	@echo "GCP Automotive Discovery Fabric - Commands"
	@echo "  make dev          - Run API server locally"
	@echo "  make test         - Run all unit and integration tests"
	@echo "  make build        - Compile all binaries into bin/"
	@echo "  make seed         - Seed synthetic discovery dataset into DB & index"
	@echo "  make eval         - Run 8-question evaluation benchmark against legacy bots"
	@echo "  make migrate-up   - Run Goose migrations UP against PostgreSQL"
	@echo "  make migrate-down - Run Goose migrations DOWN"
	@echo "  make sqlc         - Regenerate sqlc Go database queries"
	@echo "  make docker-up    - Start local PostgreSQL 17 container"
	@echo "  make docker-down  - Stop local PostgreSQL 17 container"

dev:
	DATABASE_URL="$(DB_URL)" go run ./cmd/api

test:
	go test -v ./...

build:
	mkdir -p bin
	go build -o bin/api ./cmd/api
	go build -o bin/enterprise-search ./cmd/enterprise-search
	go build -o bin/workflow ./cmd/workflow
	go build -o bin/seed ./cmd/seed
	go build -o bin/evaluation-runner ./cmd/evaluation-runner

seed:
	DATABASE_URL="$(DB_URL)" go run ./cmd/seed

eval:
	DATABASE_URL="$(DB_URL)" go run ./cmd/evaluation-runner

migrate-up:
	go run github.com/pressly/goose/v3/cmd/goose -dir migrations/postgres postgres "$(DB_URL)" up

migrate-down:
	go run github.com/pressly/goose/v3/cmd/goose -dir migrations/postgres postgres "$(DB_URL)" down

sqlc:
	sqlc generate

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
