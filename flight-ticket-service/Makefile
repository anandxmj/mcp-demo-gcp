# Flight Ticket Service Makefile

.PHONY: help build run test clean docs swagger-gen swagger-install deps

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Install dependencies
deps: ## Install Go dependencies
	go mod tidy
	go mod download

# Install swagger CLI tool
swagger-install: ## Install swag CLI tool for generating OpenAPI specs
	go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger documentation
swagger-gen: ## Generate OpenAPI specification and Swagger docs
	$(shell go env GOPATH)/bin/swag init -g src/cmd/server/server.go -o docs
	@echo "Swagger documentation generated in docs/ directory"
	@echo "Access Swagger UI at: http://localhost:8080/swagger/"

# Generate docs (alias for swagger-gen)
docs: swagger-gen ## Generate OpenAPI specification and Swagger docs

# Build the application
build: ## Build the server binary
	go build -o server src/cmd/server/server.go
	@echo "Server binary built: ./server"

# Run the application
run: ## Run the server
	go run src/cmd/server/server.go

# Run with hot reload (requires air)
dev: ## Run with hot reload (install air first: go install github.com/cosmtrek/air@latest)
	air

# Run tests
test: ## Run all tests
	go test -v ./...

# Run tests with coverage
test-coverage: ## Run tests with coverage report
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean: ## Clean build artifacts and generated files
	rm -f server
	rm -f coverage.out coverage.html
	rm -rf docs/

# Format code
fmt: ## Format Go code
	go fmt ./...

# Lint code (requires golangci-lint)
lint: ## Lint Go code
	golangci-lint run

# Build Docker image
docker-build: ## Build Docker image
	docker build -t flight-ticket-service .

# Run Docker container
docker-run: ## Run Docker container
	docker run -p 8080:8080 \
		-e GOOGLE_CLOUD_PROJECT=${GOOGLE_CLOUD_PROJECT} \
		-e GOOGLE_APPLICATION_CREDENTIALS=/app/credentials.json \
		-v ${GOOGLE_APPLICATION_CREDENTIALS}:/app/credentials.json \
		flight-ticket-service

# Full setup for development
setup: deps swagger-install swagger-gen ## Complete setup for development
	@echo "Setup complete!"
	@echo "Run 'make run' to start the server"
	@echo "Access Swagger UI at: http://localhost:8080/swagger/"

# Quick start
start: swagger-gen build run ## Generate docs, build, and run the server
