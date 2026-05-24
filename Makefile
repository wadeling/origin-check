.PHONY: dev up down migrate seed test build

up:
	docker compose up -d postgres redis

down:
	docker compose down

dev-api:
	go run ./cmd/api

dev-worker:
	go run ./cmd/worker

dev-scheduler:
	go run ./cmd/scheduler

seed:
	go run ./cmd/seed

test:
	go test ./...

build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker
	go build -o bin/scheduler ./cmd/scheduler
