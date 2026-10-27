@echo off
setlocal enabledelayedexpansion

cd /d "%~dp0"

set ACCOUNTS_CSV=%1
if "%ACCOUNTS_CSV%"=="" set ACCOUNTS_CSV=data\accounts_100.csv

set TX_CSV=%2
if "%TX_CSV%"=="" set TX_CSV=data\transactions_adversarial_10k.csv

set WORKERS=%3
if "%WORKERS%"=="" set WORKERS=16

echo =================================================================
echo  Kjoerer Go Bank Simulering med CSV-data (Windows CMD)
echo =================================================================
echo  Konto-fil        : %ACCOUNTS_CSV%
echo  Transaksjons-fil : %TX_CSV%
echo  Arbeidstraader   : %WORKERS%
echo.

go run ./cmd/simulation "%ACCOUNTS_CSV%" "%TX_CSV%" %WORKERS%
