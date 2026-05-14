@echo off
REM Impermanence build script

set SCRIPT_DIR=%~dp0
cd /d %SCRIPT_DIR%

echo Building for Windows...
go build -o anatta.exe anatta.go
go build -o duhkha.exe duhkha.go

echo Build complete.