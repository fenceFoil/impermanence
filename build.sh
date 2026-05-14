#!/bin/bash
# Impermanence build script

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

mkdir -p build

echo "Building for Linux..."
go build -o build/anatta anatta.go
go build -o build/duhkha duhkha.go

echo "Build complete."