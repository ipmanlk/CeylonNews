.PHONY: api-build api-dev api-test help

help:
	@echo "Ceylon News - Available Commands"
	@echo "  make api-build      - Build production binary"
	@echo "  make api-dev        - Run development server"
	@echo "  make api-test       - Run all tests"

api-build:
	cd api && go build --tags "fts5" -o build/cnapi ./cmd/server

api-dev:
	cd api && go run --tags "fts5" cmd/server/main.go

api-test:
	cd api && go test -v --tags "fts5" ./...
