#!/usr/bin/env bash

echo "Running vendor, verify & tidy"
go mod tidy && go mod vendor && go mod verify

echo "Running formatting rules"
go fmt ./...

echo "Running go vet"
go vet ./cmd

echo "Running golangci-lint"
golangci-lint run

echo "Running security check"
gosec -quiet -fmt=text -out=gosec.txt ./...

echo "Running unit-tests with race check"
go test -count=1 -race ./...

echo "Running secrets check"
rm -f ./gitleaks.json && gitleaks detect --no-git -f json -r ./gitleaks.json
