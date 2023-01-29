lint:
	gofumpt -w .
	go mod tidy
	golangci-lint run ./...

up:
	docker-compose up -d

start:


get:
	curl 'http://0.0.0.0:8080/api/v1/users'

test: up
	go test -v ./tests/users_test.go
