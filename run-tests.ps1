# Powershell script for å kjøre alle tester og race detector i Go
$ErrorActionPreference = "Stop"
$ProjectRoot = $PSScriptRoot
Set-Location $ProjectRoot

Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host " Kjører Go Bank Transfer Test Suite (-race og grundig validering)" -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan

Write-Host "`n1. Kjører go test -v ./... (Detaljert testkjøring)..." -ForegroundColor Yellow
go test -v ./...

Write-Host "`n2. Kjører go test -race ./... (Data Race Detector)..." -ForegroundColor Yellow
go test -race ./...

Write-Host "`n=================================================================" -ForegroundColor Green
Write-Host " ALLE GO-TESTER BESTÅTT SUKSESSFULLT UTEN DATA RACES!" -ForegroundColor Green
Write-Host "=================================================================" -ForegroundColor Green
