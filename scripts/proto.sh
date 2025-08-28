#!/bin/bash

# Install protoc-gen-go and protoc-gen-go-grpc if not already installed
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Create output directory
mkdir -p proto/

# Generate Go code from proto files
# Specify the directory containing your .proto files as the --proto_path
# Then refer to the .proto files relative to that path
protoc --proto_path=./proto \
       --go_out=./proto --go_opt=paths=source_relative \
       --go-grpc_out=./proto --go-grpc_opt=paths=source_relative \
       ./proto/*.proto

echo "Protocol buffer files generated successfully!"