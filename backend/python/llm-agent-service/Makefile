# Makefile for LLM Agent Service (using uv)

.PHONY: help install sync proto test lint format clean run dev

help:  ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install:  ## Install all dependencies (including dev)
	uv sync --all-extras

sync:  ## Sync dependencies from lock file
	uv sync

proto:  ## Generate proto files
	@echo "Generating proto files..."
	bash generate_proto.sh

test:  ## Run tests with coverage
	uv run pytest

test-verbose:  ## Run tests with verbose output
	uv run pytest -v -s

test-cov:  ## Run tests and generate HTML coverage report
	uv run pytest --cov=app --cov-report=html
	@echo "Coverage report generated in htmlcov/index.html"

lint:  ## Run linting (ruff)
	uv run ruff check .

lint-fix:  ## Run linting and auto-fix issues
	uv run ruff check --fix .

format:  ## Format code with black
	uv run black .

format-check:  ## Check code formatting
	uv run black --check .

type-check:  ## Run type checking with mypy
	uv run mypy app/

quality:  ## Run all quality checks
	@echo "Running linting..."
	@$(MAKE) lint
	@echo "Running formatting check..."
	@$(MAKE) format-check
	@echo "Running type check..."
	@$(MAKE) type-check

clean:  ## Clean generated files and cache
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete
	find . -type d -name ".pytest_cache" -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name ".ruff_cache" -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name "htmlcov" -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name ".coverage" -delete
	rm -rf app/api/generated/*.py 2>/dev/null || true

run:  ## Run the service
	uv run python -m app.main

dev:  ## Run in development mode with auto-reload
	LLM_AGENT_DEBUG=true uv run uvicorn app.main:app --reload --host 0.0.0.0 --port 8080

docker-build:  ## Build Docker image
	docker build -t llm-agent-service:latest .

docker-run:  ## Run Docker container
	docker run -p 8080:8080 -p 50051:50051 --env-file .env llm-agent-service:latest

# Development helpers
.PHONY: setup-dev
setup-dev:  ## Setup development environment
	@echo "Setting up development environment..."
	@$(MAKE) install
	@$(MAKE) proto
	@echo "Setup complete! Run 'make dev' to start the service."

.PHONY: ci
ci:  ## Run CI checks (lint, format, type-check, test)
	@echo "Running CI checks..."
	@$(MAKE) quality
	@$(MAKE) test
	@echo "All CI checks passed!"

.PHONY: lock
lock:  ## Update lock file
	uv lock
