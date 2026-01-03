# Script for rebuilding the project (after code changes)
Write-Host "Rebuilding go-event-order-processor..." -ForegroundColor Yellow
Write-Host ""

# Stop
Write-Host "Step 1: Stopping services..." -ForegroundColor Cyan
docker-compose -f deployments/docker-compose.yml down

# Rebuild
Write-Host ""
Write-Host "Step 2: Rebuilding images..." -ForegroundColor Cyan
docker-compose -f deployments/docker-compose.yml build

# Start
Write-Host ""
Write-Host "Step 3: Starting updated services..." -ForegroundColor Cyan
docker-compose -f deployments/docker-compose.yml up -d

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "Rebuild completed successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Available services:" -ForegroundColor Yellow
    Write-Host "   order-api:      http://localhost:8080" -ForegroundColor White
    Write-Host "   payment-worker: http://localhost:8081/metrics" -ForegroundColor White
    Write-Host "   Prometheus:     http://localhost:9090" -ForegroundColor White
    Write-Host "   Grafana:        http://localhost:3000 (admin/admin)" -ForegroundColor White
} else {
    Write-Host ""
    Write-Host "Error during rebuild" -ForegroundColor Red
    exit 1
}
