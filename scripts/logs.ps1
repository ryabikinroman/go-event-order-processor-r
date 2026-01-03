# Script for viewing logs
param(
    [string]$Service = ""
)

if ($Service -eq "") {
    Write-Host "Logs of all services (Ctrl+C to exit)" -ForegroundColor Cyan
    Write-Host ""
    docker-compose -f deployments/docker-compose.yml logs -f
} else {
    Write-Host "Logs of service: $Service (Ctrl+C to exit)" -ForegroundColor Cyan
    Write-Host ""
    docker-compose -f deployments/docker-compose.yml logs -f $Service
}
