#!/bin/sh
go mod tidy
GOOS=linux GOARCH=amd64 go build -o build/linux-amd64/fediverse-matrix-keygen
GOOS=linux GOARCH=arm64 go build -o build/linux-arm64/fediverse-matrix-keygen
tar -czf build/fediverse-matrix-keygen.tgz README.md build/linux*
