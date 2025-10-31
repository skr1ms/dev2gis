.PHONY: help install test lint format clean run dev

help:
	@echo "Available targets:"
	@echo "  install          - Install dependencies"
	@echo "  test             - Run unit tests"
	@echo "  test-cov         - Run tests with coverage"
	@echo "  lint             - Run code linting"
	@echo "  format           - Format code with black"
	@echo "  clean            - Clean generated files"
	@echo "  run              - Run the service"
	@echo "  dev              - Run in development mode"

install:
	@echo "Installing dependencies..."
	pip install --no-cache-dir -r requirements.txt
	@echo "Dependencies installed"

test:
	@echo "Running unit tests..."
	pytest tests/ -v

test-cov:
	@echo "Running tests with coverage..."
	pytest tests/ --cov=heightmap_service --cov-report=term-missing --cov-report=html
	@echo "Coverage report generated: htmlcov/index.html"

lint:
	@echo "Running linters..."
	flake8 heightmap_service/ tests/ --max-line-length=100 --exclude=generated,venv || true
	pylint heightmap_service/ || true

format:
	@echo "Formatting code..."
	black heightmap_service/ tests/ --line-length=100 || true
	isort heightmap_service/ tests/ || true

clean:
	@echo "Cleaning generated files..."
	rm -rf __pycache__
	rm -rf .pytest_cache
	rm -rf htmlcov
	rm -rf .coverage
	rm -rf *.egg-info
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete
	@echo "Clean complete"

run:
	@echo "Running heightmap service..."
	python -m heightmap_service

dev:
	@echo "Running in development mode..."
	LOG_LEVEL=DEBUG python -m heightmap_service

