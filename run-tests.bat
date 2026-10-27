@echo off
setlocal enabledelayedexpansion

cd /d "%~dp0"

echo =================================================================
echo  Kjoerer Go Bank Transfer Test Suite (Windows CMD)
echo =================================================================

echo.
echo 1. Kjoerer go test -v ./...
go test -v ./...
if %errorlevel% neq 0 (
    echo Noen tester feilet!
    exit /b %errorlevel%
)

echo.
echo 2. Kjoerer go test -race ./... (Data Race Detector)
go test -race ./...
if %errorlevel% neq 0 (
    echo Data race detektert!
    exit /b %errorlevel%
)

echo.
echo =================================================================
echo  ALLE TESTER BESTAATT UTEN DATA RACES!
echo =================================================================
