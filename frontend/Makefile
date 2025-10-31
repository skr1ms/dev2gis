.PHONY: help install dev build lint format clean test

help:
	@echo "Available targets:"
	@echo "  install          - Install dependencies"
	@echo "  dev              - Run development server"
	@echo "  build            - Build production bundle"
	@echo "  lint             - Run ESLint"
	@echo "  format           - Format code with Prettier"
	@echo "  clean            - Clean generated files"
	@echo "  test             - Run tests"

install:
	@echo "Installing dependencies..."
	npm install
	@echo "Dependencies installed"

dev:
	@echo "Starting development server..."
	npm run dev

build:
	@echo "Building production bundle..."
	npm run build

lint:
	@echo "Running ESLint..."
	npm run lint

format:
	@echo "Formatting code..."
	npx prettier --write "src/**/*.{ts,tsx,js,jsx,json,css}"

clean:
	@echo "Cleaning generated files..."
	rm -rf .next
	rm -rf out
	rm -rf node_modules/.cache
	rm -rf coverage
	@echo "Clean complete"

test:
	@echo "Running tests..."
	npm test

