# Script for stopping the project
Write-Host "Stopping go-event-order-processor..." -ForegroundColor Yellow
Write-Host ""

docker-compose -f deployments/docker-compose.yml stop

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "All services stopped" -ForegroundColor Green
    Write-Host ""
    Write-Host "Data saved. To restart use: .\scripts\start.ps1" -ForegroundColor Cyan
} else {
    Write-Host ""
    Write-Host "Error stopping services" -ForegroundColor Red
    exit 1
}
