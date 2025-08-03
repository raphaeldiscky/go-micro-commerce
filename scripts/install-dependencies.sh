#!/bin/bash

# Usage: ./install-dependencies.sh [folder|service-name]
# Examples:
#   ./install-dependencies.sh              # Process all go.mod files
#   ./install-dependencies.sh pkg          # Process pkg directory
#   ./install-dependencies.sh product      # Process services/product-service
#   ./install-dependencies.sh auth         # Process services/auth-service
#   ./install-dependencies.sh services     # Process all services

# Function to determine search directory
get_search_dir() {
    local input="$1"
    
    # If no input, use current directory
    if [ -z "$input" ]; then
        echo "."
        return
    fi
    
    # If input is "pkg", use pkg directory
    if [ "$input" = "pkg" ]; then
        echo "pkg"
        return
    fi

    # If input is "api-gateway", use api-gateway directory
    if [ "$input" = "api-gateway" ]; then
        echo "api-gateway"
        return
    fi
    
    # If input is "services", use services directory
    if [ "$input" = "services" ]; then
        echo "services"
        return
    fi
    
    # If input exists as a directory, use it directly
    if [ -d "$input" ]; then
        echo "$input"
        return
    fi
    
    # Check if it's a service name (try services/{input}-service)
    if [ -d "services/${input}-service" ]; then
        echo "services/${input}-service"
        return
    fi
    
    # Check if it's a service name without -service suffix
    if [ -d "services/${input}" ]; then
        echo "services/${input}"
        return
    fi
    
    # If nothing matches, return the input as-is
    echo "$input"
}

# Default to current directory if no argument provided
SEARCH_DIR=$(get_search_dir "$1")

# Validate that the directory exists
if [ ! -d "$SEARCH_DIR" ]; then
    echo "Error: Directory '$SEARCH_DIR' does not exist"
    exit 1
fi

echo "Searching for go.mod files in: $SEARCH_DIR"

# Find all go.mod files and run go mod tidy && go mod vendor for each
find "$SEARCH_DIR" -name "go.mod" -type f | while read -r gomod; do
    dir=$(dirname "$gomod")
    echo "Processing $gomod in directory $dir"
    (cd "$dir" && go mod tidy && go mod vendor)
done

echo "Finished processing all go.mod files"