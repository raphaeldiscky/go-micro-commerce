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

# install node.js tools
npm install

# add husky hooks
npx husky init
echo "
if git diff --cached --name-only --diff-filter=ACM | grep -qE '\.go$'; then
  echo "Go files detected → running task full_check..."
  task full_check
else
  echo "No Go files detected → skipping task full_check"
fi

if git diff --cached --name-only --diff-filter=ACM | grep -q '^frontend/'; then
  echo "frontend/ changes detected → running frontend lint..."
  (
    cd frontend || exit 1
    npm run lint
  )
fi

git add -A .
" > .husky/pre-commit
echo "npx --no-install commitlint --edit \$1" > .husky/commit-msg