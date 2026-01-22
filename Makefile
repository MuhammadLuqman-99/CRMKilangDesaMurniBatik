# CRM Kilang Desa Murni Batik - Makefile
# =========================================

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Binary names
IAM_BINARY=iam-service
CUSTOMER_BINARY=customer-service
SALES_BINARY=sales-service
NOTIFICATION_BINARY=notification-service
GATEWAY_BINARY=api-gateway

# Directories
CMD_DIR=./cmd
BIN_DIR=./bin
PKG_DIR=./pkg
INTERNAL_DIR=./internal

# Docker parameters
DOCKER_COMPOSE=docker-compose
DOCKER=docker

# Version info
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Colors for terminal output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: all build clean test coverage lint fmt help
.PHONY: build-iam build-customer build-sales build-notification build-gateway
.PHONY: run-iam run-customer run-sales run-notification run-gateway run-all
.PHONY: docker-build docker-up docker-down docker-logs docker-clean
.PHONY: migrate-up migrate-down migrate-create
.PHONY: proto-gen swagger-gen
.PHONY: deps deps-update deps-tidy deps-verify
.PHONY: dev dev-iam dev-customer dev-sales dev-notification dev-gateway

# Default target
all: deps lint test build

# ==========================================
# HELP
# ==========================================

help: ## Display this help message
	@echo "CRM Kilang Desa Murni Batik - Available Commands"
	@echo "================================================"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(BLUE)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""

# ==========================================
# DEPENDENCIES
# ==========================================

deps: ## Download all dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	$(GOMOD) download
	@echo "$(GREEN)Dependencies downloaded successfully$(NC)"

deps-update: ## Update all dependencies
	@echo "$(YELLOW)Updating dependencies...$(NC)"
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "$(GREEN)Dependencies updated successfully$(NC)"

deps-tidy: ## Tidy go.mod and go.sum
	@echo "$(YELLOW)Tidying dependencies...$(NC)"
	$(GOMOD) tidy
	@echo "$(GREEN)Dependencies tidied successfully$(NC)"

deps-verify: ## Verify dependencies
	@echo "$(YELLOW)Verifying dependencies...$(NC)"
	$(GOMOD) verify
	@echo "$(GREEN)Dependencies verified successfully$(NC)"

# ==========================================
# BUILD
# ==========================================

build: build-iam build-customer build-sales build-notification build-gateway ## Build all services
	@echo "$(GREEN)All services built successfully$(NC)"

build-iam: ## Build IAM service
	@echo "$(YELLOW)Building IAM service...$(NC)"
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(IAM_BINARY) $(CMD_DIR)/iam-service
	@echo "$(GREEN)IAM service built: $(BIN_DIR)/$(IAM_BINARY)$(NC)"

build-customer: ## Build Customer service
	@echo "$(YELLOW)Building Customer service...$(NC)"
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(CUSTOMER_BINARY) $(CMD_DIR)/customer-service
	@echo "$(GREEN)Customer service built: $(BIN_DIR)/$(CUSTOMER_BINARY)$(NC)"

build-sales: ## Build Sales service
	@echo "$(YELLOW)Building Sales service...$(NC)"
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(SALES_BINARY) $(CMD_DIR)/sales-service
	@echo "$(GREEN)Sales service built: $(BIN_DIR)/$(SALES_BINARY)$(NC)"

build-notification: ## Build Notification service
	@echo "$(YELLOW)Building Notification service...$(NC)"
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(NOTIFICATION_BINARY) $(CMD_DIR)/notification-service
	@echo "$(GREEN)Notification service built: $(BIN_DIR)/$(NOTIFICATION_BINARY)$(NC)"

build-gateway: ## Build API Gateway
	@echo "$(YELLOW)Building API Gateway...$(NC)"
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(GATEWAY_BINARY) $(CMD_DIR)/api-gateway
	@echo "$(GREEN)API Gateway built: $(BIN_DIR)/$(GATEWAY_BINARY)$(NC)"

build-linux: ## Build all services for Linux
	@echo "$(YELLOW)Building all services for Linux...$(NC)"
	@mkdir -p $(BIN_DIR)/linux
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/linux/$(IAM_BINARY) $(CMD_DIR)/iam-service
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/linux/$(CUSTOMER_BINARY) $(CMD_DIR)/customer-service
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/linux/$(SALES_BINARY) $(CMD_DIR)/sales-service
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/linux/$(NOTIFICATION_BINARY) $(CMD_DIR)/notification-service
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/linux/$(GATEWAY_BINARY) $(CMD_DIR)/api-gateway
	@echo "$(GREEN)All Linux binaries built in $(BIN_DIR)/linux/$(NC)"

# ==========================================
# RUN
# ==========================================

run-iam: ## Run IAM service
	@echo "$(GREEN)Starting IAM service...$(NC)"
	$(GORUN) $(CMD_DIR)/iam-service/main.go

run-customer: ## Run Customer service
	@echo "$(GREEN)Starting Customer service...$(NC)"
	$(GORUN) $(CMD_DIR)/customer-service/main.go

run-sales: ## Run Sales service
	@echo "$(GREEN)Starting Sales service...$(NC)"
	$(GORUN) $(CMD_DIR)/sales-service/main.go

run-notification: ## Run Notification service
	@echo "$(GREEN)Starting Notification service...$(NC)"
	$(GORUN) $(CMD_DIR)/notification-service/main.go

run-gateway: ## Run API Gateway
	@echo "$(GREEN)Starting API Gateway...$(NC)"
	$(GORUN) $(CMD_DIR)/api-gateway/main.go

# ==========================================
# DEVELOPMENT (Hot Reload with Air)
# ==========================================

dev-iam: ## Run IAM service with hot reload
	@echo "$(GREEN)Starting IAM service with hot reload...$(NC)"
	cd $(CMD_DIR)/iam-service && air

dev-customer: ## Run Customer service with hot reload
	@echo "$(GREEN)Starting Customer service with hot reload...$(NC)"
	cd $(CMD_DIR)/customer-service && air

dev-sales: ## Run Sales service with hot reload
	@echo "$(GREEN)Starting Sales service with hot reload...$(NC)"
	cd $(CMD_DIR)/sales-service && air

dev-notification: ## Run Notification service with hot reload
	@echo "$(GREEN)Starting Notification service with hot reload...$(NC)"
	cd $(CMD_DIR)/notification-service && air

dev-gateway: ## Run API Gateway with hot reload
	@echo "$(GREEN)Starting API Gateway with hot reload...$(NC)"
	cd $(CMD_DIR)/api-gateway && air

# ==========================================
# TEST
# ==========================================

test: ## Run all tests
	@echo "$(YELLOW)Running tests...$(NC)"
	$(GOTEST) -v -race -short ./...
	@echo "$(GREEN)Tests completed$(NC)"

test-unit: ## Run unit tests only
	@echo "$(YELLOW)Running unit tests...$(NC)"
	$(GOTEST) -v -short -tags=unit ./...

test-integration: ## Run integration tests
	@echo "$(YELLOW)Running integration tests...$(NC)"
	$(GOTEST) -v -tags=integration ./...

test-coverage: ## Run tests with coverage
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	@mkdir -p coverage
	$(GOTEST) -v -race -coverprofile=coverage/coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "$(GREEN)Coverage report generated: coverage/coverage.html$(NC)"

test-bench: ## Run benchmarks
	@echo "$(YELLOW)Running benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...

# ==========================================
# CODE QUALITY
# ==========================================

lint: ## Run linter
	@echo "$(YELLOW)Running linter...$(NC)"
	$(GOLINT) run ./...
	@echo "$(GREEN)Linting completed$(NC)"

lint-fix: ## Run linter with auto-fix
	@echo "$(YELLOW)Running linter with auto-fix...$(NC)"
	$(GOLINT) run --fix ./...
	@echo "$(GREEN)Linting with fixes completed$(NC)"

fmt: ## Format code
	@echo "$(YELLOW)Formatting code...$(NC)"
	$(GOFMT) -s -w .
	@echo "$(GREEN)Code formatted$(NC)"

fmt-check: ## Check code formatting
	@echo "$(YELLOW)Checking code formatting...$(NC)"
	@test -z "$$($(GOFMT) -l .)" || (echo "$(RED)Code is not formatted. Run 'make fmt'$(NC)" && exit 1)
	@echo "$(GREEN)Code formatting check passed$(NC)"

vet: ## Run go vet
	@echo "$(YELLOW)Running go vet...$(NC)"
	$(GOCMD) vet ./...
	@echo "$(GREEN)Vet completed$(NC)"

security: ## Run security scanner
	@echo "$(YELLOW)Running security scanner...$(NC)"
	gosec ./...
	@echo "$(GREEN)Security scan completed$(NC)"

# ==========================================
# DOCKER
# ==========================================

docker-build: ## Build all Docker images
	@echo "$(YELLOW)Building Docker images...$(NC)"
	$(DOCKER_COMPOSE) build
	@echo "$(GREEN)Docker images built$(NC)"

docker-build-iam: ## Build IAM service Docker image
	$(DOCKER) build -t crm-iam-service:$(VERSION) -f deployments/docker/Dockerfile.iam .

docker-build-customer: ## Build Customer service Docker image
	$(DOCKER) build -t crm-customer-service:$(VERSION) -f deployments/docker/Dockerfile.customer .

docker-build-sales: ## Build Sales service Docker image
	$(DOCKER) build -t crm-sales-service:$(VERSION) -f deployments/docker/Dockerfile.sales .

docker-build-notification: ## Build Notification service Docker image
	$(DOCKER) build -t crm-notification-service:$(VERSION) -f deployments/docker/Dockerfile.notification .

docker-build-gateway: ## Build API Gateway Docker image
	$(DOCKER) build -t crm-api-gateway:$(VERSION) -f deployments/docker/Dockerfile.gateway .

docker-up: ## Start all containers
	@echo "$(GREEN)Starting containers...$(NC)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Containers started$(NC)"

docker-down: ## Stop all containers
	@echo "$(YELLOW)Stopping containers...$(NC)"
	$(DOCKER_COMPOSE) down
	@echo "$(GREEN)Containers stopped$(NC)"

docker-logs: ## Show container logs
	$(DOCKER_COMPOSE) logs -f

docker-logs-iam: ## Show IAM service logs
	$(DOCKER_COMPOSE) logs -f iam-service

docker-logs-customer: ## Show Customer service logs
	$(DOCKER_COMPOSE) logs -f customer-service

docker-logs-sales: ## Show Sales service logs
	$(DOCKER_COMPOSE) logs -f sales-service

docker-logs-notification: ## Show Notification service logs
	$(DOCKER_COMPOSE) logs -f notification-service

docker-logs-gateway: ## Show API Gateway logs
	$(DOCKER_COMPOSE) logs -f api-gateway

docker-ps: ## Show running containers
	$(DOCKER_COMPOSE) ps

docker-clean: ## Remove all containers and volumes
	@echo "$(RED)Removing all containers and volumes...$(NC)"
	$(DOCKER_COMPOSE) down -v --remove-orphans
	@echo "$(GREEN)Cleanup completed$(NC)"

docker-restart: docker-down docker-up ## Restart all containers

# ==========================================
# DATABASE MIGRATIONS
# ==========================================

migrate-up: ## Run all migrations
	@echo "$(YELLOW)Running migrations...$(NC)"
	migrate -path migrations/iam -database "$(IAM_DATABASE_URL)" up
	migrate -path migrations/sales -database "$(SALES_DATABASE_URL)" up
	@echo "$(GREEN)Migrations completed$(NC)"

migrate-down: ## Rollback all migrations
	@echo "$(YELLOW)Rolling back migrations...$(NC)"
	migrate -path migrations/iam -database "$(IAM_DATABASE_URL)" down
	migrate -path migrations/sales -database "$(SALES_DATABASE_URL)" down
	@echo "$(GREEN)Rollback completed$(NC)"

migrate-up-iam: ## Run IAM migrations
	@echo "$(YELLOW)Running IAM migrations...$(NC)"
	migrate -path migrations/iam -database "$(IAM_DATABASE_URL)" up

migrate-down-iam: ## Rollback IAM migrations
	@echo "$(YELLOW)Rolling back IAM migrations...$(NC)"
	migrate -path migrations/iam -database "$(IAM_DATABASE_URL)" down 1

migrate-up-sales: ## Run Sales migrations
	@echo "$(YELLOW)Running Sales migrations...$(NC)"
	migrate -path migrations/sales -database "$(SALES_DATABASE_URL)" up

migrate-down-sales: ## Rollback Sales migrations
	@echo "$(YELLOW)Rolling back Sales migrations...$(NC)"
	migrate -path migrations/sales -database "$(SALES_DATABASE_URL)" down 1

migrate-create: ## Create a new migration (usage: make migrate-create name=migration_name service=iam)
	@echo "$(YELLOW)Creating migration...$(NC)"
	migrate create -ext sql -dir migrations/$(service) -seq $(name)
	@echo "$(GREEN)Migration created$(NC)"

migrate-force: ## Force migration version (usage: make migrate-force version=1 service=iam)
	migrate -path migrations/$(service) -database "$($(shell echo $(service) | tr a-z A-Z)_DATABASE_URL)" force $(version)

# ==========================================
# CODE GENERATION
# ==========================================

proto-gen: ## Generate code from proto files
	@echo "$(YELLOW)Generating protobuf code...$(NC)"
	protoc --go_out=. --go-grpc_out=. api/proto/*.proto
	@echo "$(GREEN)Protobuf code generated$(NC)"

swagger-gen: ## Generate Swagger documentation
	@echo "$(YELLOW)Generating Swagger documentation...$(NC)"
	swag init -g cmd/api-gateway/main.go -o api/openapi
	@echo "$(GREEN)Swagger documentation generated$(NC)"

mock-gen: ## Generate mocks for testing
	@echo "$(YELLOW)Generating mocks...$(NC)"
	mockgen -source=internal/iam/domain/repository.go -destination=internal/iam/domain/mock/repository_mock.go
	mockgen -source=internal/customer/domain/repository.go -destination=internal/customer/domain/mock/repository_mock.go
	mockgen -source=internal/sales/domain/repository.go -destination=internal/sales/domain/mock/repository_mock.go
	@echo "$(GREEN)Mocks generated$(NC)"

# ==========================================
# CLEANUP
# ==========================================

clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	rm -rf $(BIN_DIR)
	rm -rf tmp
	rm -rf coverage
	rm -rf vendor
	$(GOCMD) clean -cache -testcache
	@echo "$(GREEN)Cleanup completed$(NC)"

clean-docker: ## Clean Docker resources
	@echo "$(RED)Cleaning Docker resources...$(NC)"
	$(DOCKER) system prune -f
	@echo "$(GREEN)Docker cleanup completed$(NC)"

clean-all: clean clean-docker ## Clean everything

# ==========================================
# INSTALLATION
# ==========================================

install-tools: ## Install development tools
	@echo "$(YELLOW)Installing development tools...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/air-verse/air@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install -tags 'postgres mongodb' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "$(GREEN)Development tools installed$(NC)"

# ==========================================
# UTILITIES
# ==========================================

version: ## Show version info
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

info: ## Show project info
	@echo "$(BLUE)CRM Kilang Desa Murni Batik$(NC)"
	@echo "================================"
	@echo "Go Version: $(shell go version)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo ""
	@echo "Services:"
	@echo "  - IAM Service"
	@echo "  - Customer Service"
	@echo "  - Sales Service"
	@echo "  - Notification Service"
	@echo "  - API Gateway"

check: fmt-check vet lint test ## Run all checks

ci: deps check build ## Run CI pipeline
