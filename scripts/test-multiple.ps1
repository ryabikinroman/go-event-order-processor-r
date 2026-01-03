# Script for creating multiple test orders
param(
    [int]$Count = 5
)

Write-Host "Creating $Count test orders..." -ForegroundColor Cyan
Write-Host ""

for ($i = 1; $i -le $Count; $i++) {
    Write-Host "[$i/$Count] Creating order..." -ForegroundColor Yellow

    $jsonBody = @{
        customer_id = "test-customer-$i"
        amount = [math]::Round((Get-Random -Minimum 10 -Maximum 200), 2)
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/orders" `
            -Method POST `
            -Headers @{"Content-Type"="application/json"} `
            -Body $jsonBody `
            -ErrorAction Stop

        Write-Host "  Order created: $($response.order_id)" -ForegroundColor Green
    } catch {
        Write-Host "  Error: $_" -ForegroundColor Red
    }

    Start-Sleep -Milliseconds 500
}

Write-Host ""
Write-Host "Completed! Created $Count orders" -ForegroundColor Green
Write-Host ""
Write-Host "Check metrics:" -ForegroundColor Yellow
Write-Host "  Prometheus: http://localhost:9090" -ForegroundColor White
Write-Host "  Grafana:    http://localhost:3000" -ForegroundColor White
Write-Host "  Logs:       .\scripts\logs.ps1 payment-worker" -ForegroundColor White
