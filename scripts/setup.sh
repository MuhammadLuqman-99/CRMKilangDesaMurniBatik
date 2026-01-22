#!/bin/bash
# CRM Kilang Desa Murni Batik - Setup Script
# ==========================================

set -e

echo "========================================"
echo "CRM Kilang Desa Murni Batik - Setup"
echo "========================================"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Check prerequisites
echo -e "\n${YELLOW}Checking prerequisites...${NC}"

check_command() {
    if command -v $1 &> /dev/null; then
        echo -e "${GREEN}✓ $1 is installed${NC}"
        return 0
    else
        echo -e "${RED}✗ $1 is not installed${NC}"
        return 1
    fi
}

check_command "go" || exit 1
check_command "docker" || exit 1
check_command "docker-compose" || echo -e "${YELLOW}Warning: docker-compose not found, using 'docker compose'${NC}"

# Install development tools
echo -e "\n${YELLOW}Installing development tools...${NC}"

echo "Installing golangci-lint..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest || echo "Failed to install golangci-lint"

echo "Installing air (hot reload)..."
go install github.com/air-verse/air@latest || echo "Failed to install air"

echo "Installing mockgen..."
go install github.com/golang/mock/mockgen@latest || echo "Failed to install mockgen"

echo "Installing swag (Swagger)..."
go install github.com/swaggo/swag/cmd/swag@latest || echo "Failed to install swag"

echo "Installing migrate..."
go install -tags 'postgres mongodb' github.com/golang-migrate/migrate/v4/cmd/migrate@latest || echo "Failed to install migrate"

# Download Go dependencies
echo -e "\n${YELLOW}Downloading Go dependencies...${NC}"
go mod download
go mod tidy

# Create necessary directories
echo -e "\n${YELLOW}Creating directories...${NC}"
mkdir -p bin tmp coverage

# Copy example configuration
echo -e "\n${YELLOW}Setting up configuration...${NC}"
if [ ! -f configs/config.yaml ]; then
    cp configs/dev/config.yaml configs/config.yaml
    echo "Created configs/config.yaml from development template"
fi

# Start infrastructure services
echo -e "\n${YELLOW}Starting infrastructure services...${NC}"
if command -v docker-compose &> /dev/null; then
    docker-compose up -d postgres mongodb redis rabbitmq
else
    docker compose up -d postgres mongodb redis rabbitmq
fi

# Wait for services to be ready
echo -e "\n${YELLOW}Waiting for services to be ready...${NC}"
sleep 10

# Run database migrations
echo -e "\n${YELLOW}Note: Run 'make migrate-up' after services are healthy${NC}"

echo -e "\n${GREEN}========================================"
echo "Setup complete!"
echo "========================================${NC}"
echo ""
echo "Next steps:"
echo "1. Wait for Docker services to be healthy: docker-compose ps"
echo "2. Run database migrations: make migrate-up"
echo "3. Start development servers: make dev-iam (or other services)"
echo ""
echo "Available commands:"
echo "  make help          - Show all available commands"
echo "  make docker-up     - Start all services"
echo "  make docker-logs   - View logs"
echo "  make test          - Run tests"
echo ""
