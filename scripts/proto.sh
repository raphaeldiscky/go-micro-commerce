#!/bin/bash


set -e
echo "This is legacy script, use buf instead"
# Check if protoc-gen-go is installed
if ! command -v protoc-gen-go &> /dev/null; then
  echo "Installing protoc-gen-go..."
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

# Check if protoc-gen-go-grpc is installed
if ! command -v protoc-gen-go-grpc &> /dev/null; then
  echo "Installing protoc-gen-go-grpc..."
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Clean up old generated files
find proto -name "*.pb.go" -delete

# Generate Go code from all proto files recursively
echo "Generating Go code from proto files..."
find proto -name "*.proto" | while read -r file; do
  protoc --proto_path=proto \
         --go_out=. --go_opt=module=github.com/raphaeldiscky/go-micro-commerce \
         --go-grpc_out=. --go-grpc_opt=module=github.com/raphaeldiscky/go-micro-commerce \
         "$file"
done

echo "Protocol buffer files generated successfully!"
