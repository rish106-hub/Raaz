.PHONY: help build run test clean docker-up docker-down docker-logs fmt lint

help:
	@echo "Raaz Development Commands"
	@echo ""
	@echo "Setup & Build:"
	@echo "  make build        - Build the Go server binary"
	@echo "  make install-deps - Install/update Go dependencies"
	@echo ""
	@echo "Development:"
	@echo "  make run          - Run server in-memory mode (no DB needed)"
	@echo "  make test         - Run all tests"
	@echo "  make test-race    - Run tests with race detector"
	@echo "  make fmt          - Format code (gofmt)"
	@echo "  make lint         - Run linter (go vet)"
	@echo ""
	@echo "Docker Stack:"
	@echo "  make docker-up    - Start PostgreSQL + Redis + Server"
	@echo "  make docker-down  - Stop all containers"
	@echo "  make docker-logs  - View container logs"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean        - Remove build artifacts"

build:
	cd server && go build -o ../raaz-server .
	@echo "✅ Binary built: ./raaz-server"

install-deps:
	cd server && go mod tidy
	@echo "✅ Dependencies updated"

run: build
	./raaz-server

test:
	cd server && go test ./... -v -timeout 30s

test-race:
	cd server && go test -race ./... -timeout 30s

fmt:
	cd server && go fmt ./...
	@echo "✅ Code formatted"

lint:
	cd server && go vet ./...
	@echo "✅ Linting complete"

docker-up:
	docker-compose up -d
	@echo "✅ Docker stack started"
	@echo "  Server: http://localhost:8080"
	@echo "  PostgreSQL: localhost:5432"
	@echo "  Redis: localhost:6379"

docker-down:
	docker-compose down
	@echo "✅ Docker stack stopped"

docker-logs:
	docker-compose logs -f server

docker-db-logs:
	docker-compose logs -f postgres redis

clean:
	rm -f raaz-server
	@echo "✅ Cleaned up"

all: fmt lint test build
	@echo "✅ All checks passed!"
