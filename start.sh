#!/bin/bash

# Point-of-Sale System Startup Script (Linux/macOS)

echo ""
echo "================================"
echo "  POS System Startup"
echo "================================"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "ERROR: Go is not installed"
    echo "Please install Go from https://golang.org/"
    exit 1
fi

echo "✓ Go is installed"
echo ""

# Check if in correct directory
if [ ! -f "main.go" ]; then
    echo "ERROR: main.go not found"
    echo "Please run this script from the pos-system directory"
    exit 1
fi

echo "✓ Directory is correct"
echo ""

# Install dependencies
echo "Installing dependencies..."
go get github.com/mattn/go-sqlite3

echo ""
echo "================================"
echo "  Starting POS System..."
echo "================================"
echo ""
echo "The system will start on http://localhost:8080"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# Run the application
go run main.go
