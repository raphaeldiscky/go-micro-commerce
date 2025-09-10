#!/bin/bash

# format & lint tools
if ! command -v gofumpt &> /dev/null; then
    echo "Installing gofumpt..."
    go install mvdan.cc/gofumpt@latest
else
    echo "gofumpt already installed"
fi

if ! command -v goimports &> /dev/null; then
    echo "Installing goimports..."
    go install golang.org/x/tools/cmd/goimports@latest
else
    echo "goimports already installed"
fi

if ! command -v golangci-lint &> /dev/null; then
    echo "Installing golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6
else
    echo "golangci-lint already installed"
fi

if ! command -v deadcode &> /dev/null; then
    echo "Installing deadcode..."
    go install golang.org/x/tools/cmd/deadcode@latest
else 
    echo "deadcode already installed"
fi

if ! command -v task &> /dev/null; then
    echo "Installing task..."
    curl -sSfL https://taskfile.dev/install.sh | sh
else
    echo "task already installed"
fi

if ! command -v protoc &> /dev/null; then
    echo "Installing protoc..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
else
    echo "protoc already installed"
fi

if ! command -v air &> /dev/null; then
    echo "Installing air..."
    go install github.com/air-verse/air@latest
else
    echo "air already installed"
fi

# install node.js tools
npm install

# add husky hooks
npx husky init
echo "task format && task lint && git add -A ." > .husky/pre-commit
echo "task test" > .husky/pre-push
echo "npx --no-install commitlint --edit \$1" > .husky/commit-msg