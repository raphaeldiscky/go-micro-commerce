#!/bin/bash

set -e

TEST_FILE=$1

if [ -z "$TEST_FILE" ]; then
  echo "Please provide a test file to run (e.g., order-service/internal/foo/foo_test.go)"
  exit 1
fi

if [ ! -f "$TEST_FILE" ]; then
  echo "File does not exist: $TEST_FILE"
  exit 1
fi

# Extract directory path
TEST_DIR=$(dirname "$TEST_FILE")

echo "Running test file: $TEST_FILE"
echo "Test directory: $TEST_DIR"

# Run tests in that directory, only matching the specified file
go test -v -coverprofile=coverage.out -covermode=atomic "$TEST_DIR" -run "$(basename "$TEST_FILE" .go)"

echo "Test run complete."
echo "Coverage report:"
go tool cover -func=coverage.out
