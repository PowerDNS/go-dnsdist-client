#!/bin/sh

set -ex
go test ./...
go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.58.0 run
