# Script for testing the API
Write-Host "Testing order-api" -ForegroundColor Cyan
Write-Host ""

Write-Host "Creating test order..." -ForegroundColor Yellow

# Create JSON file temporarily to avoid PowerShell escaping issues
$jsonBody = @{
    customer_id = "test-customer-123"
    amount = 99.99
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/orders" `
    -Method POST `
    -Headers @{"Content-Type"="application/json"} `
    -Body $jsonBody `
    -ErrorAction SilentlyContinue

if ($response) {
    Write-Host ""
    Write-Host "Order created successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Response:" -ForegroundColor Cyan
    $response | ConvertTo-Json -Depth 10
    Write-Host ""
    Write-Host "View processing: .\scripts\logs.ps1 payment-worker" -ForegroundColor Yellow
} else {
    Write-Host ""
    Write-Host "Error creating order" -ForegroundColor Red
    Write-Host "Check that services are running: .\scripts\status.ps1" -ForegroundColor Yellow
}
