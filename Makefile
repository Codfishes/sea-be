
# Variables
APP_NAME=sea-catering-backend
MAIN_PATH=./cmd/app
BUILD_DIR=./bin
MIGRATION_DIR=./database/migrations
SEED_DIR=./database/seeds

# Database
DB_HOST?=localhost
DB_PORT?=5432
DB_USER?=tyokeren
DB_PASSWORD?=14Oktober04.
DB_NAME?=sea_catering
DB_SSLMODE?=disable
DATABASE_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# Migration version (for create command)
VERSION?=""
NAME?=""

# Docker
DOCKER_IMAGE=$(APP_NAME)
DOCKER_TAG?=latest

# Colors for output
RED=\033[31m
GREEN=\033[32m
YELLOW=\033[33m
BLUE=\033[34m
RESET=\033[0m

## help: Display this help message
help:
	@echo "$(BLUE)SEA Catering Backend - Available Commands:$(RESET)"
	@echo ""
	@echo "$(GREEN)Setup & Installation:$(RESET)"
	@echo "  setup           - Install tools and setup development environment"
	@echo "  install-tools   - Install required development tools"
	@echo "  deps            - Download and verify dependencies"
	@echo ""
	@echo "$(GREEN)Development:$(RESET)"
	@echo "  dev             - Run development server with hot reload"
	@echo "  build           - Build the application"
	@echo "  run             - Run the application"
	@echo "  clean           - Clean build files"
	@echo ""
	@echo "$(GREEN)Code Quality:$(RESET)"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  test-race       - Run tests with race detection"
	@echo "  lint            - Run golangci-lint"
	@echo "  fmt             - Format code"
	@echo "  vet             - Run go vet"
	@echo ""
	@echo "$(GREEN)Database:$(RESET)"
	@echo "  migrate-up      - Run all up migrations"
	@echo "  migrate-down    - Run one down migration"
	@echo "  migrate-force   - Force migration version (VERSION=x)"
	@echo "  migrate-version - Show current migration version"
	@echo "  migrate-create  - Create new migration (NAME=migration_name)"
	@echo "  migrate-drop    - Drop all migrations (WARNING: destructive)"
	@echo "  seed            - Run database seeds"
	@echo ""
	@echo "$(GREEN)Docker:$(RESET)"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run with Docker Compose"
	@echo "  docker-down     - Stop Docker Compose"
	@echo "  docker-logs     - Show Docker logs"
	@echo ""
	@echo "$(YELLOW)Examples:$(RESET)"
	@echo "  make migrate-create NAME=add_user_preferences"
	@echo "  make migrate-force VERSION=3"
	@echo "  make test-coverage"

## setup: Install tools and setup development environment
setup: install-tools deps
	@echo "$(GREEN)✓ Development environment setup complete$(RESET)"

## install-tools: Install required development tools
install-tools:
	@echo "$(BLUE)Installing development tools...$(RESET)"
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)✓ Tools installed successfully$(RESET)"

## deps: Download and verify dependencies
deps:
	@echo "$(BLUE)Downloading dependencies...$(RESET)"
	@go mod download
	@go mod verify
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(RESET)"

## dev: Run development server with hot reload
dev:
	@echo "$(BLUE)Starting development server...$(RESET)"
	@air -c .air.toml

## build: Build the application
build:
	@echo "$(BLUE)Building application...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✓ Build successful: $(BUILD_DIR)/$(APP_NAME)$(RESET)"

## run: Run the application
run:
	@echo "$(BLUE)Running application...$(RESET)"
	@go run $(MAIN_PATH)/main.go

## clean: Clean build files
clean:
	@echo "$(BLUE)Cleaning build files...$(RESET)"
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "$(GREEN)✓ Clean complete$(RESET)"

## test: Run tests
test:
	@echo "$(BLUE)Running tests...$(RESET)"
	@go test -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "$(BLUE)Running tests with coverage...$(RESET)"
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(RESET)"

## test-race: Run tests with race detection
test-race:
	@echo "$(BLUE)Running tests with race detection...$(RESET)"
	@go test -v -race ./...

## lint: Run golangci-lint
lint:
	@echo "$(BLUE)Running linter...$(RESET)"
	@golangci-lint run

## fmt: Format code
fmt:
	@echo "$(BLUE)Formatting code...$(RESET)"
	@go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(RESET)"

## vet: Run go vet
vet:
	@echo "$(BLUE)Running go vet...$(RESET)"
	@go vet ./...
	@echo "$(GREEN)✓ Vet complete$(RESET)"

## migrate-up: Run all up migrations
migrate-up:
	@echo "$(BLUE)Running migrations up...$(RESET)"
	@migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" up
	@echo "$(GREEN)✓ Migrations applied$(RESET)"

## migrate-down: Run one down migration
migrate-down:
	@echo "$(YELLOW)Running one migration down...$(RESET)"
	@migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" down 1
	@echo "$(GREEN)✓ Migration rolled back$(RESET)"

## migrate-force: Force migration version
migrate-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)Error: VERSION is required. Usage: make migrate-force VERSION=3$(RESET)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Forcing migration to version $(VERSION)...$(RESET)"
	@migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" force $(VERSION)
	@echo "$(GREEN)✓ Migration forced to version $(VERSION)$(RESET)"

## migrate-version: Show current migration version
migrate-version:
	@echo "$(BLUE)Current migration version:$(RESET)"
	@migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" version

## migrate-create: Create new migration
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)Error: NAME is required. Usage: make migrate-create NAME=add_user_preferences$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Creating new migration: $(NAME)...$(RESET)"
	@migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME)
	@echo "$(GREEN)✓ Migration files created$(RESET)"

## migrate-drop: Drop all migrations (WARNING: destructive)
migrate-drop:
	@echo "$(RED)WARNING: This will drop all tables and data!$(RESET)"
	@echo "$(YELLOW)Are you sure? [y/N]: $(RESET)"
	@read -r confirm && [ "$confirm" = "y" ] || [ "$confirm" = "Y" ] || (echo "Cancelled." && exit 1)
	@migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" drop
	@echo "$(GREEN)✓ Database dropped$(RESET)"

## seed: Run database seeds
seed:
	@echo "$(BLUE)Running database seeds...$(RESET)"
	@for file in $(SEED_DIR)/*.sql; do \
		echo "Running seed: $file"; \
		psql "$(DATABASE_URL)" -f $file; \
	done
	@echo "$(GREEN)✓ Seeds applied$(RESET)"

## docker-build: Build Docker image
docker-build:
	@echo "$(BLUE)Building Docker image...$(RESET)"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(GREEN)✓ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(RESET)"

## docker-run: Run with Docker Compose
docker-run:
	@echo "$(BLUE)Starting services with Docker Compose...$(RESET)"
	@docker-compose up -d
	@echo "$(GREEN)✓ Services started$(RESET)"

## docker-down: Stop Docker Compose
docker-down:
	@echo "$(BLUE)Stopping Docker Compose services...$(RESET)"
	@docker-compose down
	@echo "$(GREEN)✓ Services stopped$(RESET)"

## docker-logs: Show Docker logs
docker-logs:
	@docker-compose logs -f

# Internal targets
.PHONY: check-migrate
check-migrate:
	@which migrate > /dev/null || (echo "$(RED)Error: migrate tool not found. Run 'make install-tools' first.$(RESET)" && exit 1)