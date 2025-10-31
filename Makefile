.PHONY: help install build test lint clean dev up down logs restart scale deploy-prod deploy-dev rollback-prod rollback-dev

help:
	@echo "════════════════════════════════════════════════════════════════════════"
	@echo "  UAV Heights Map - Project Makefile"
	@echo "════════════════════════════════════════════════════════════════════════"
	@echo ""
	@echo "Development Commands:"
	@echo "  make install         - Install all dependencies"
	@echo "  make dev             - Start development environment"
	@echo "  make build           - Build all services"
	@echo "  make test            - Run all tests"
	@echo "  make lint            - Run linters"
	@echo "  make clean           - Clean all generated files"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make up              - Start all services"
	@echo "  make down            - Stop all services"
	@echo "  make restart svc=X   - Restart specific service"
	@echo "  make logs svc=X      - Show logs for service"
	@echo "  make scale svc=X n=N - Scale service to N instances"
	@echo "  make ps              - Show running containers"
	@echo ""
	@echo "Deployment Commands:"
	@echo "  make deploy-prod COMMIT=<sha>    - Deploy to production"
	@echo "  make deploy-dev COMMIT=<sha>     - Deploy to development"
	@echo "  make rollback-prod               - Rollback production"
	@echo "  make rollback-dev                - Rollback development"
	@echo ""
	@echo "Service-Specific Commands:"
	@echo "  make go-<target>     - Run target in go-orchestrator"
	@echo "  make py-<target>     - Run target in python-heightmap"
	@echo "  make fe-<target>     - Run target in frontend"
	@echo ""
	@echo "Available services: postgres, rabbitmq, minio, nodeodm, go-orchestrator,"
	@echo "                    python-heightmap, frontend, nginx, prometheus, grafana"
	@echo "════════════════════════════════════════════════════════════════════════"

install:
	@echo "Installing all dependencies..."
	@cd backend/go-orchestrator && make deps
	@cd backend/python-services/heightmap-service && make install
	@cd frontend && make install
	@echo "All dependencies installed successfully"

build:
	@echo "Building all services..."
	@cd backend/go-orchestrator && make build
	@cd frontend && make build
	@echo "All services built successfully"

test:
	@echo "Running all tests..."
	@cd backend/go-orchestrator && make test
	@cd backend/python-services/heightmap-service && make test
	@cd frontend && make test || true
	@echo "All tests completed"

lint:
	@echo "Running linters..."
	@cd backend/go-orchestrator && go fmt ./... && go vet ./...
	@cd backend/python-services/heightmap-service && make lint
	@cd frontend && make lint || true
	@echo "Linting completed"

clean:
	@echo "Cleaning all generated files..."
	@cd backend/go-orchestrator && make clean
	@cd backend/python-services/heightmap-service && make clean
	@cd frontend && make clean
	@echo "Clean completed"

dev:
	@echo "Starting development environment..."
	docker compose -f deployment/docker/docker-compose.dev.yml --env-file .env.dev up -d
	@echo "Development environment started"
	@echo "Frontend: http://localhost:80"
	@echo "Backend API: http://localhost:8080"
	@echo "MinIO Console: http://localhost:9001"
	@echo "RabbitMQ Management: http://localhost:15672"
	@echo "Grafana: http://localhost:3000"

up:
	@echo "Starting all services..."
	docker compose -f deployment/docker/docker-compose.yml --env-file .env.prod up -d
	@echo "All services started"

down:
	@echo "Stopping all services..."
	docker compose -f deployment/docker/docker-compose.yml down
	@echo "All services stopped"

logs:
	@cd deployment && make logs svc=$(svc)

restart:
	@cd deployment && make restart svc=$(svc)

scale:
	@cd deployment && make scale svc=$(svc) n=$(n)

ps:
	@cd deployment && make ps

deploy-prod:
	@cd deployment && make deploy-prod COMMIT=$(COMMIT)

deploy-dev:
	@cd deployment && make deploy-dev COMMIT=$(COMMIT)

rollback-prod:
	@cd deployment && make rollback-prod

rollback-dev:
	@cd deployment && make rollback-dev

go-%:
	@cd backend/go-orchestrator && make $*

py-%:
	@cd backend/python-services/heightmap-service && make $*

fe-%:
	@cd frontend && make $*

