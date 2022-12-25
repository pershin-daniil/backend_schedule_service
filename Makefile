lint:
	gofumpt -w .
	go mod tidy
	golangci-lint run ./...
