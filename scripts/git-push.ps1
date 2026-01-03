# Quick git add, commit, push script
param(
    [Parameter(Mandatory=$true)]
    [string]$Message
)

Write-Host "Checking git status..." -ForegroundColor Cyan
git status --short

Write-Host ""
$confirmation = Read-Host "Continue with commit? (y/N)"
if ($confirmation -ne 'y' -and $confirmation -ne 'Y') {
    Write-Host "Cancelled" -ForegroundColor Yellow
    exit 0
}

Write-Host ""
Write-Host "Adding files..." -ForegroundColor Cyan
git add .

Write-Host "Creating commit..." -ForegroundColor Cyan
git commit -m "$Message"

Write-Host "Pushing to remote..." -ForegroundColor Cyan
$branch = git branch --show-current
git push origin $branch

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "Successfully pushed to $branch!" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "Error during push" -ForegroundColor Red
    exit 1
}
