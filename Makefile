.PHONY: run test lint format clean build

run:
	go run ./cmd/server

test:
	go test ./... -count=1 -race -timeout=60s

lint:
	golangci-lint run

format:
	gofumpt -l -w .
	goimports -w .

build:
	go build -o bin/server ./cmd/server

clean:
	rm -rf bin/

deps:
	go mod download
	go mod tidy

install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest
	go install golang.org/x/tools/cmd/goimports@latest
