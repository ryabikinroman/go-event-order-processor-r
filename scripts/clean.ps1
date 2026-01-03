# Script for full cleanup (removes volumes and images)
Write-Host "WARNING: Full cleanup will remove ALL data!" -ForegroundColor Red
Write-Host ""

$confirmation = Read-Host "Continue? (y/N)"
if ($confirmation -ne 'y' -and $confirmation -ne 'Y') {
    Write-Host "Cancelled" -ForegroundColor Yellow
    exit 0
}

Write-Host ""
Write-Host "Removing containers, volumes and images..." -ForegroundColor Yellow
docker-compose -f deployments/docker-compose.yml down -v --rmi all

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "Cleanup completed" -ForegroundColor Green
    Write-Host ""
    Write-Host "To start fresh use: .\scripts\start.ps1" -ForegroundColor Cyan
} else {
    Write-Host ""
    Write-Host "Error during cleanup" -ForegroundColor Red
    exit 1
}
