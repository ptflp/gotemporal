APP_NAME ?= gotemporal

.PHONY: up down logs gen test build fmt

up:
	docker compose up -d

down:
	docker compose down -v

logs:
	docker compose logs -f app temporal temporal-ui

gen:
	go generate ./ent

test:
	go test ./...

build:
	go build ./cmd/app

fmt:
	go fmt ./...

