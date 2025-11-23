#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}Linting Dockerfiles with hadolint...${NC}"
echo ""

# Check if hadolint is installed
if ! command -v hadolint &> /dev/null; then
    echo -e "${RED}Error: hadolint is not installed${NC}"
    echo "Install with: brew install hadolint (macOS) or see https://github.com/hadolint/hadolint"
    exit 1
fi

# Find all Dockerfiles
DOCKERFILES=$(find . -type f \( -name "Dockerfile" -o -name "Dockerfile.*" \) -not -path "*/node_modules/*" -not -path "*/.git/*" | sort)

if [ -z "$DOCKERFILES" ]; then
    echo -e "${YELLOW}No Dockerfiles found${NC}"
    exit 0
fi

# Count files
TOTAL=$(echo "$DOCKERFILES" | wc -l | tr -d ' ')
PASSED=0
FAILED=0
FAILED_FILES=""

echo -e "${BLUE}Found ${TOTAL} Dockerfile(s) to lint${NC}"
echo ""

# Lint each Dockerfile
for dockerfile in $DOCKERFILES; do
    echo -n "  Linting ${dockerfile}... "

    if hadolint --config .hadolint.yaml "$dockerfile" 2>&1; then
        echo -e "${GREEN}OK${NC}"
        ((PASSED++))
    else
        echo -e "${RED}FAILED${NC}"
        ((FAILED++))
        FAILED_FILES="${FAILED_FILES}\n  - ${dockerfile}"
    fi
done

echo ""
echo "==================== HADOLINT SUMMARY ===================="
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All ${TOTAL} Dockerfile(s) passed linting${NC}"
    echo ""
    echo "=========================================================="
    exit 0
else
    echo -e "${RED}${FAILED}/${TOTAL} Dockerfile(s) failed linting${NC}"
    echo -e "${RED}Failed files:${FAILED_FILES}${NC}"
    echo ""
    echo "=========================================================="
    exit 1
fi
