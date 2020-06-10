#!/bin/sh

echo "Use -short to skip testing against dnsdist"

set -ex
go test ./... "$@"
go run github.com/golangci/golangci-lint/cmd/golangci-lint run
