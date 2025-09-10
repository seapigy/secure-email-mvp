# Makefile for Secure Email Backend

# Variables
DOCKER_IMAGE_NAME=secure-email-backend
CONTAINER_NAME=secure-email-app
DB_CONTAINER_NAME=secure-email-db
MIGRATION_PATH=migrations

# Default target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build     - Build Docker image"
	@echo "  run       - Run docker-compose up"
	@echo "  stop      - Stop docker-compose"
	@echo "  migrate   - Run migrations against MySQL container"
	@echo "  clean     - Clean up containers and images"
	@echo "  logs      - Show application logs"
	@echo "  test      - Run tests"

# Build Docker image
.PHONY: build
build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE_NAME) .
	@echo "Docker image built successfully!"

# Run docker-compose up
.PHONY: run
run:
	@echo "Starting services with docker-compose..."
	docker-compose up -d
	@echo "Services started! Backend: http://localhost:8080, MySQL: localhost:3306"

# Stop docker-compose
.PHONY: stop
stop:
	@echo "Stopping services..."
	docker-compose down
	@echo "Services stopped!"

# Run migrations
.PHONY: migrate
migrate:
	@echo "Running database migrations..."
	@echo "Waiting for database to be ready..."
	@sleep 10
	@echo "Creating users table..."
	docker exec $(DB_CONTAINER_NAME) mysql -u secureuser -psecurepass securesystem -e "SOURCE /docker-entrypoint-initdb.d/001_create_users_table.sql;"
	@echo "Migration completed!"

# Clean up containers and images
.PHONY: clean
clean:
	@echo "Cleaning up containers and images..."
	docker-compose down -v
	docker rmi $(DOCKER_IMAGE_NAME) 2>/dev/null || true
	docker system prune -f
	@echo "Cleanup completed!"

# Show application logs
.PHONY: logs
logs:
	@echo "Showing application logs..."
	docker-compose logs -f app

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test ./... -v

# Development setup
.PHONY: dev
dev: build run migrate
	@echo "Development environment ready!"
	@echo "Backend: http://localhost:8080"
	@echo "MySQL: localhost:3306 (user: secureuser, password: securepass, database: securesystem)"
	@echo "Run 'make logs' to see application logs"

# Check if services are running
.PHONY: status
status:
	@echo "Checking service status..."
	@docker-compose ps

