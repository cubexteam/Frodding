@echo off
go mod tidy
go build -o frodding.exe ./cmd/frodding
echo Build complete: frodding.exe
