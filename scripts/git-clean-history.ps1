# Script to clean Git history (USE WITH CAUTION!)
# This will create a fresh repository with current state only

Write-Host "WARNING: This will erase ALL Git history!" -ForegroundColor Red
Write-Host "Only current state of files will remain." -ForegroundColor Red
Write-Host ""
Write-Host "Before running this:" -ForegroundColor Yellow
Write-Host "1. Make sure you have a backup" -ForegroundColor Yellow
Write-Host "2. Make sure all important changes are committed" -ForegroundColor Yellow
Write-Host ""

$confirmation = Read-Host "Are you ABSOLUTELY sure? Type 'YES' to continue"
if ($confirmation -ne 'YES') {
    Write-Host "Cancelled - good choice!" -ForegroundColor Green
    exit 0
}

Write-Host ""
Write-Host "Step 1: Removing .git folder..." -ForegroundColor Cyan
Remove-Item -Recurse -Force .git

Write-Host "Step 2: Initializing fresh repository..." -ForegroundColor Cyan
git init

Write-Host "Step 3: Creating initial commit..." -ForegroundColor Cyan
git add .
git commit -m "Initial commit: event-driven order processor"

Write-Host "Step 4: Creating develop branch..." -ForegroundColor Cyan
git branch develop

Write-Host ""
Write-Host "History cleaned! Now you need to force push:" -ForegroundColor Green
Write-Host ""
Write-Host "git remote add origin https://github.com/ryabikinroman/go-event-order-processor.git" -ForegroundColor Yellow
Write-Host "git push -u --force origin main" -ForegroundColor Yellow
Write-Host "git push -u --force origin develop" -ForegroundColor Yellow
Write-Host ""
Write-Host "WARNING: This will overwrite remote repository!" -ForegroundColor Red
