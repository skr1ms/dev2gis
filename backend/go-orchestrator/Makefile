.PHONY: help swagger test test-coverage build run clean deps sqlc

help:
	@echo "Available targets:"
	@echo "  swagger          - Generate Swagger documentation"
	@echo "  sqlc             - Generate SQLC code from queries"
	@echo "  test             - Run unit tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  build            - Build the application"
	@echo "  run              - Run the application"
	@echo "  clean            - Clean generated files"
	@echo "  deps             - Install dependencies"

sqlc:
	@echo "Generating SQLC code..."
	@if ! command -v sqlc >/dev/null 2>&1; then \
		echo "Installing sqlc..."; \
		go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest; \
	fi
	sqlc generate
	@echo "SQLC code generated successfully"

swagger:
	@echo "Generating Swagger documentation..."
	@if ! command -v swag >/dev/null 2>&1; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	swag init -g cmd/main.go -o docs --parseDependency --parseInternal
	@echo "Swagger documentation generated successfully"

test:
	@echo "Running unit tests..."
	go test ./... -v -race

test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -v -race -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

build:
	@echo "Building application..."
	go build -o bin/app cmd/main.go
	@echo "Build complete: bin/app"

run:
	@echo "Running application..."
	go run cmd/main.go

clean:
	@echo "Cleaning generated files..."
	rm -rf docs/
	rm -f coverage.out coverage.html
	rm -rf bin/
	@echo "Clean complete"

deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed"

