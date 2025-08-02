#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Running tests for multi-module Go project...${NC}"

# Find all go.mod files and run tests for each module
modules=$(find . -name "go.mod" -not -path "./vendor/*" | sort)

if [ -z "$modules" ]; then
    echo -e "${RED}No Go modules found!${NC}"
    exit 1
fi

# Initialize coverage files array
coverage_files=()
success_count=0
skip_count=0

# Run tests for each module
for module_file in $modules; do
    module_dir=$(dirname "$module_file")
    module_name=$(grep '^module ' "$module_file" | cut -d' ' -f2)
    
    echo -e "\n${YELLOW}Testing module: ${module_name} (${module_dir})${NC}"
    
    cd "$module_dir"
    
    # Check if there are any packages to test
    packages=$(go list ./... 2>/dev/null || echo "")
    if [ -z "$packages" ]; then
        echo -e "${YELLOW}No packages to test in ${module_name}, skipping...${NC}"
        skip_count=$((skip_count + 1))
        cd - > /dev/null
        continue
    fi
    
    # Generate coverage file name based on module path (relative to root)
    if [ "$module_dir" = "." ]; then
        coverage_file="coverage-root.out"
    else
        coverage_file="coverage-$(echo "$module_dir" | tr '/' '-' | sed 's/^-//').out"
    fi
    
    # Run tests with coverage
    echo "Running: go test -v ./... -coverprofile=$coverage_file -covermode=atomic"
    if go test -v ./... -coverprofile="$coverage_file" -covermode=atomic; then
        echo -e "${GREEN}✓ Tests passed for ${module_name}${NC}"
        success_count=$((success_count + 1))
        if [ -f "$coverage_file" ]; then
            coverage_files+=("$coverage_file")
        fi
    else
        echo -e "${RED}✗ Tests failed for ${module_name}${NC}"
        cd - > /dev/null
        exit 1
    fi
    
    cd - > /dev/null
    
    # Move coverage file to root if it was generated in a subdirectory
    if [ -f "$module_dir/$coverage_file" ] && [ "$module_dir" != "." ]; then
        mv "$module_dir/$coverage_file" "./$coverage_file"
    fi
done

# Summary
echo -e "\n${YELLOW}Test Summary:${NC}"
echo -e "  Modules tested: $success_count"
echo -e "  Modules skipped (no packages): $skip_count"

# Merge coverage files if multiple modules have coverage
if [ ${#coverage_files[@]} -gt 1 ]; then
    echo -e "\n${YELLOW}Merging coverage files...${NC}"
    # Create merged coverage file
    echo "mode: atomic" > coverage.out
    for file in "${coverage_files[@]}"; do
        if [ -f "$file" ]; then
            tail -n +2 "$file" >> coverage.out
            rm "$file"
        fi
    done
    echo -e "${GREEN}✓ Coverage files merged into coverage.out${NC}"
elif [ ${#coverage_files[@]} -eq 1 ]; then
    # Rename single coverage file to expected name
    mv "${coverage_files[0]}" coverage.out
    echo -e "${GREEN}✓ Coverage file created: coverage.out${NC}"
else
    echo -e "${YELLOW}No coverage files generated (no test files found)${NC}"
fi

if [ $success_count -eq 0 ] && [ $skip_count -gt 0 ]; then
    echo -e "\n${YELLOW}All modules were skipped (no packages to test). This might be expected for a template project.${NC}"
    exit 0
fi

echo -e "\n${GREEN}All tests completed successfully!${NC}"