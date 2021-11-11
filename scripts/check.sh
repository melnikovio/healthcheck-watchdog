#!/usr/bin/env bash

echo "vendor"
go mod vendor && go mod verify && go mod tidy

echo "go vet"
go vet ./cmd

echo "golangci-lint"
golangci-lint run
