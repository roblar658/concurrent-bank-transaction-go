# Powershell script for å kjøre parallelle simuleringer med CSV-data i Go
param (
    [string]$AccountsCsv = "data/accounts_100.csv",
    [string]$TransactionsCsv = "data/transactions_adversarial_10k.csv",
    [int]$Workers = 16
)

$ErrorActionPreference = "Stop"
$ProjectRoot = $PSScriptRoot
Set-Location $ProjectRoot

Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host " Kjører Go Bank Simulering med CSV-data" -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "  Konto-fil        : $AccountsCsv"
Write-Host "  Transaksjons-fil : $TransactionsCsv"
Write-Host "  Arbeidstråder    : $Workers"

go run ./cmd/simulation $AccountsCsv $TransactionsCsv $Workers
