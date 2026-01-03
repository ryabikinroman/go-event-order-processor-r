# Script for quick project startup
Write-Host "Starting go-event-order-processor..." -ForegroundColor Green
Write-Host ""

# Check Docker
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "Docker not found. Please install Docker Desktop." -ForegroundColor Red
    exit 1
}

# Check if Docker is running
docker info 2>&1 | Out-Null
if ($LASTEXITCODE -ne 0) {
    Write-Host "Docker Desktop is not running. Please start Docker Desktop and try again." -ForegroundColor Red
    exit 1
}

# Start services
Write-Host "Starting containers..." -ForegroundColor Cyan
docker-compose -f deployments/docker-compose.yml up -d

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "All services started successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Available services:" -ForegroundColor Yellow
    Write-Host "   order-api:      http://localhost:8080" -ForegroundColor White
    Write-Host "   payment-worker: http://localhost:8081/metrics" -ForegroundColor White
    Write-Host "   Prometheus:     http://localhost:9090" -ForegroundColor White
    Write-Host "   Grafana:        http://localhost:3000 (admin/admin)" -ForegroundColor White
    Write-Host ""
    Write-Host "Check status: .\scripts\status.ps1" -ForegroundColor Yellow
    Write-Host "Stop:         .\scripts\stop.ps1" -ForegroundColor Yellow
    Write-Host "Logs:         .\scripts\logs.ps1" -ForegroundColor Yellow
} else {
    Write-Host ""
    Write-Host "Error starting services" -ForegroundColor Red
    exit 1
}
