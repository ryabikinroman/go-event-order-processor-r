# Script for checking container status
Write-Host "Status of go-event-order-processor containers" -ForegroundColor Cyan
Write-Host ""

docker-compose -f deployments/docker-compose.yml ps
