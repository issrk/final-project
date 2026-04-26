@echo off
REM Point-of-Sale System Startup Script

echo.
echo ================================
echo  POS System Startup
echo ================================
echo.

REM Check if Go is installed
go version >nul 2>&1
if errorlevel 1 (
    echo ERROR: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/
    pause
    exit /b 1
)

echo √ Go is installed
echo.

REM Check if in correct directory
if not exist "main.go" (
    echo ERROR: main.go not found
    echo Please run this script from the pos-system directory
    pause
    exit /b 1
)

echo √ Directory is correct
echo.

REM Install dependencies
echo Installing dependencies...
go get github.com/mattn/go-sqlite3
if errorlevel 1 (
    echo WARNING: Could not install dependencies
    echo You may need to manually run: go get github.com/mattn/go-sqlite3
)

echo.
echo ================================
echo  Starting POS System...
echo ================================
echo.
echo The system will start on http://localhost:8080
echo.
echo Press Ctrl+C to stop the server
echo.

REM Run the application
go run main.go

pause
