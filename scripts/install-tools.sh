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
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.4.0
else
    echo "golangci-lint already installed"
fi

if ! command -v deadcode &> /dev/null; then
    echo "Installing deadcode..."
    go install golang.org/x/tools/cmd/deadcode@latest
else 
    echo "deadcode already installed"
fi

if ! command -v govulncheck &> /dev/null; then
    echo "Installing govulncheck..."
    go install golang.org/x/vuln/cmd/govulncheck@latest
else
    echo "govulncheck already installed"
fi

if ! command -v oapi-codegen &> /dev/null; then
    echo "Installing oapi..."
    go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

else
    echo "oapi already installed"
fi

# install node.js tools
npm install

# install lefthook (git hooks manager) if missing — pinned via @vX.Y.Z
if ! command -v lefthook >/dev/null 2>&1; then
  if command -v go >/dev/null 2>&1; then
    go install github.com/evilmartians/lefthook/v2@v2.1.8
  else
    echo "go not found on PATH; skipping lefthook install (git hooks will NOT be wired)"
  fi
fi

# wire git hooks via lefthook (config lives in lefthook.yml)
if command -v lefthook >/dev/null 2>&1; then
  lefthook install
fi