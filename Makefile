lint:
	gofumpt -w .
	go mod tidy
	golangci-lint run

up:
	docker compose up -d db

run:
	docker compose up -d

down:
	docker compose down

get:
	curl 'http://0.0.0.0:8080/api/v1/users'

build:
	docker compose build --no-cache

test: up
	go test -v ./tests/users_test.go

integration: test
	docker compose down

.PHONY: build