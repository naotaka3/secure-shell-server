#!/bin/bash

# Set up temporary Go environment
export GOPATH=/Users/yu.shimizu/go
export GOMODCACHE=$GOPATH/pkg/mod

# Run the tests
go test -race -covermode=atomic -coverprofile=coverage.out -coverpkg=./... ./...

# Return the exit code from the test command
exit $?