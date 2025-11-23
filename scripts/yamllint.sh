#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Linting YAML files with yamllint...${NC}"
echo ""

# Check if yamllint is installed
if ! command -v yamllint &> /dev/null; then
    echo -e "${RED}Error: yamllint is not installed${NC}"
    echo "Install with: pip install yamllint"
    exit 1
fi

# Count YAML files
TOTAL=$(find . -type f \( -name "*.yaml" -o -name "*.yml" \) -not -path "*/node_modules/*" -not -path "*/.git/*" -not -path "*/vendor/*" | wc -l | tr -d ' ')

echo -e "${BLUE}Found ${TOTAL} YAML file(s) to lint${NC}"
echo ""

# Run yamllint on entire project
if yamllint -c .yamllint.yaml . 2>&1; then
    echo ""
    echo "==================== YAMLLINT SUMMARY ===================="
    echo ""
    echo -e "${GREEN}All YAML files passed linting${NC}"
    echo ""
    echo "=========================================================="
    exit 0
else
    echo ""
    echo "==================== YAMLLINT SUMMARY ===================="
    echo ""
    echo -e "${RED}Some YAML files failed linting${NC}"
    echo ""
    echo "=========================================================="
    exit 1
fi
