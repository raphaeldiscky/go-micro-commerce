#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to show usage
usage() {
    echo "Usage: $0 <file_path> [test_type]"
    echo ""
    echo "Arguments:"
    echo "  file_path   Path to the Go test file or package to test"
    echo "  test_type   Type of test (default: unit)"
    echo ""
    echo "Examples:"
    echo "  $0 ./pkg/logger/logger_test.go"
    echo "  $0 ./services/product-service/internal/service"
    echo "  $0 ./pkg/config unit"
    exit 1
}

# Check if file argument is provided
if [ $# -lt 1 ]; then
    echo -e "${RED}Error: File path is required${NC}"
    usage
fi

FILE_PATH="$1"
TEST_TYPE="${2:-unit}"

echo -e "${YELLOW}Running ${TEST_TYPE} tests for: ${FILE_PATH}${NC}"

# Check if the file/directory exists
if [ ! -e "$FILE_PATH" ]; then
    echo -e "${RED}Error: File or directory '$FILE_PATH' does not exist${NC}"
    exit 1
fi

# Determine the module directory and test target
if [ -f "$FILE_PATH" ]; then
    # If it's a file, get its directory
    if [[ "$FILE_PATH" == *_test.go ]]; then
        # It's a test file
        FILE_DIR=$(dirname "$FILE_PATH")
        TEST_TARGET="$FILE_PATH"
    else
        # It's a regular Go file, look for corresponding test file
        FILE_DIR=$(dirname "$FILE_PATH")
        FILENAME=$(basename "$FILE_PATH" .go)
        TEST_FILE="${FILE_DIR}/${FILENAME}_test.go"
        if [ -f "$TEST_FILE" ]; then
            TEST_TARGET="$TEST_FILE"
        else
            echo -e "${YELLOW}No test file found for $FILE_PATH, testing entire package${NC}"
            TEST_TARGET="$FILE_DIR"
        fi
    fi
elif [ -d "$FILE_PATH" ]; then
    # It's a directory
    FILE_DIR="$FILE_PATH"
    TEST_TARGET="$FILE_PATH"
else
    echo -e "${RED}Error: '$FILE_PATH' is neither a file nor a directory${NC}"
    exit 1
fi

# Find the nearest go.mod file
MODULE_DIR="$FILE_DIR"
while [ "$MODULE_DIR" != "/" ] && [ "$MODULE_DIR" != "." ]; do
    if [ -f "$MODULE_DIR/go.mod" ]; then
        break
    fi
    MODULE_DIR=$(dirname "$MODULE_DIR")
done

if [ ! -f "$MODULE_DIR/go.mod" ]; then
    echo -e "${RED}Error: No go.mod file found in '$FILE_DIR' or its parent directories${NC}"
    exit 1
fi

MODULE_NAME=$(grep '^module ' "$MODULE_DIR/go.mod" | cut -d' ' -f2)
echo -e "${YELLOW}Found module: ${MODULE_NAME} (${MODULE_DIR})${NC}"

# Change to module directory
cd "$MODULE_DIR"

# Convert absolute path to relative path from module root
if [[ "$TEST_TARGET" == /* ]]; then
    # Absolute path - convert to relative
    TEST_TARGET=$(realpath --relative-to="$MODULE_DIR" "$TEST_TARGET")
fi

# Ensure TEST_TARGET starts with ./ if it's a relative path
if [[ "$TEST_TARGET" != ./* ]] && [[ "$TEST_TARGET" != /* ]]; then
    TEST_TARGET="./$TEST_TARGET"
fi

# Generate coverage file name
COVERAGE_FILE="coverage-$(echo "$MODULE_DIR" | tr '/' '-' | sed 's/^-//' | sed 's/^\./root/')-test-file.out"

echo -e "${YELLOW}Test target: ${TEST_TARGET}${NC}"
echo -e "${YELLOW}Coverage file: ${COVERAGE_FILE}${NC}"

# Run the test
if [ -f "$TEST_TARGET" ]; then
    # Testing a specific test file
    echo -e "${YELLOW}Running test file: ${TEST_TARGET}${NC}"
    TEST_PACKAGE=$(dirname "$TEST_TARGET")
    if go test -v "$TEST_PACKAGE" -run "$(basename "$TEST_TARGET" .go | sed 's/_test$//')" -coverprofile="$COVERAGE_FILE" -covermode=atomic; then
        echo -e "${GREEN}✓ Tests passed for ${TEST_TARGET}${NC}"
    else
        echo -e "${RED}✗ Tests failed for ${TEST_TARGET}${NC}"
        exit 1
    fi
elif [ -d "$TEST_TARGET" ]; then
    # Testing a package/directory
    echo -e "${YELLOW}Running tests for package: ${TEST_TARGET}${NC}"
    if go test -v "$TEST_TARGET" -coverprofile="$COVERAGE_FILE" -covermode=atomic; then
        echo -e "${GREEN}✓ Tests passed for ${TEST_TARGET}${NC}"
    else
        echo -e "${RED}✗ Tests failed for ${TEST_TARGET}${NC}"
        exit 1
    fi
else
    echo -e "${RED}Error: Test target '$TEST_TARGET' not found${NC}"
    exit 1
fi

# Move coverage file to project root if we're in a subdirectory
PROJECT_ROOT=$(pwd)
while [ "$PROJECT_ROOT" != "/" ]; do
    if [ -f "$PROJECT_ROOT/go.mod" ] && [ -f "$PROJECT_ROOT/taskfile.yml" ]; then
        break
    fi
    PROJECT_ROOT=$(dirname "$PROJECT_ROOT")
done

if [ "$PROJECT_ROOT" != "$MODULE_DIR" ] && [ -f "$COVERAGE_FILE" ]; then
    echo -e "${YELLOW}Moving coverage file to project root...${NC}"
    mv "$COVERAGE_FILE" "$PROJECT_ROOT/$COVERAGE_FILE"
    echo -e "${GREEN}✓ Coverage file moved to: $PROJECT_ROOT/$COVERAGE_FILE${NC}"
elif [ -f "$COVERAGE_FILE" ]; then
    echo -e "${GREEN}✓ Coverage file created: $COVERAGE_FILE${NC}"
fi

echo -e "\n${GREEN}Test completed successfully!${NC}"