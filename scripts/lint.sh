#!/bin/bash

# Lint each Go module separately
echo "Linting pkg module..."
if [ -d "./pkg" ] && [ -f "./pkg/go.mod" ]; then
    cd ./pkg && golangci-lint run ./... --fix --timeout 5m --config ../.golangci.yml
    cd ..
fi

echo "Linting services..."
for service_dir in ./services/*/; do
    if [ -d "$service_dir" ] && [ -f "${service_dir}go.mod" ]; then
        service_name=$(basename "$service_dir")
        echo "Linting $service_name..."
        cd "$service_dir" && golangci-lint run ./... --fix --timeout 5m --config ../../.golangci.yml
        cd ../..
    fi
done

echo "Linting completed." 
