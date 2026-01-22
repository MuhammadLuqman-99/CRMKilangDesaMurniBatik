# CRM Kilang Desa Murni Batik - Setup Script (PowerShell)
# =======================================================

$ErrorActionPreference = "Stop"

Write-Host "========================================"
Write-Host "CRM Kilang Desa Murni Batik - Setup"
Write-Host "========================================"

# Function to check if a command exists
function Test-Command($cmdname) {
    return [bool](Get-Command -Name $cmdname -ErrorAction SilentlyContinue)
}

# Check prerequisites
Write-Host "`nChecking prerequisites..." -ForegroundColor Yellow

if (Test-Command "go") {
    Write-Host "[OK] Go is installed" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Go is not installed" -ForegroundColor Red
    exit 1
}

if (Test-Command "docker") {
    Write-Host "[OK] Docker is installed" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Docker is not installed" -ForegroundColor Red
    exit 1
}

# Install development tools
Write-Host "`nInstalling development tools..." -ForegroundColor Yellow

Write-Host "Installing golangci-lint..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

Write-Host "Installing air (hot reload)..."
go install github.com/air-verse/air@latest

Write-Host "Installing mockgen..."
go install github.com/golang/mock/mockgen@latest

Write-Host "Installing swag (Swagger)..."
go install github.com/swaggo/swag/cmd/swag@latest

Write-Host "Installing migrate..."
go install -tags 'postgres mongodb' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Download Go dependencies
Write-Host "`nDownloading Go dependencies..." -ForegroundColor Yellow
go mod download
go mod tidy

# Create necessary directories
Write-Host "`nCreating directories..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path bin | Out-Null
New-Item -ItemType Directory -Force -Path tmp | Out-Null
New-Item -ItemType Directory -Force -Path coverage | Out-Null

# Copy example configuration
Write-Host "`nSetting up configuration..." -ForegroundColor Yellow
if (-not (Test-Path "configs/config.yaml")) {
    Copy-Item "configs/dev/config.yaml" "configs/config.yaml"
    Write-Host "Created configs/config.yaml from development template"
}

# Start infrastructure services
Write-Host "`nStarting infrastructure services..." -ForegroundColor Yellow
docker-compose up -d postgres mongodb redis rabbitmq

# Wait for services to be ready
Write-Host "`nWaiting for services to be ready..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "Setup complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:"
Write-Host "1. Wait for Docker services to be healthy: docker-compose ps"
Write-Host "2. Run database migrations: make migrate-up"
Write-Host "3. Start development servers: make dev-iam (or other services)"
Write-Host ""
Write-Host "Available commands:"
Write-Host "  make help          - Show all available commands"
Write-Host "  make docker-up     - Start all services"
Write-Host "  make docker-logs   - View logs"
Write-Host "  make test          - Run tests"
Write-Host ""
