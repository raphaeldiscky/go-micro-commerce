#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE} $1${NC}"
}

print_success() {
    echo -e "${GREEN} $1${NC}"
}

print_error() {
    echo -e "${RED} $1${NC}"
}

# Variables for tracking validation results
failed_components=()
successful_components=()

# Check if kustomize is installed
check_kustomize() {
    if ! command -v kustomize &> /dev/null; then
        print_error "kustomize is not installed. Please install kustomize first."
        echo ""
        echo "Install with:"
        echo "  brew install kustomize  # macOS"
        echo "  curl -s https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh | bash  # Linux"
        exit 1
    fi
}

# Validate a single kustomization
validate_kustomization() {
    local kustomization_dir=$1
    local relative_path="${kustomization_dir#./}"

    print_status "Validating $relative_path..."

    local error_output
    if error_output=$(kustomize build "$kustomization_dir" 2>&1 > /dev/null); then
        echo "✓ $relative_path is valid"
        successful_components+=("$relative_path")
        return 0
    else
        print_error "✗ $relative_path validation failed"
        echo "$error_output" | sed 's/^/  /'
        failed_components+=("$relative_path")
        return 1
    fi
}

# Print validation summary
print_summary() {
    echo ""
    echo "==================== VALIDATION SUMMARY ===================="

    if [ ${#successful_components[@]} -gt 0 ]; then
        print_success "Successfully validated ${#successful_components[@]} kustomization(s)"
    fi

    if [ ${#failed_components[@]} -gt 0 ]; then
        echo ""
        print_error "Failed to validate ${#failed_components[@]} kustomization(s):"
        for component in "${failed_components[@]}"; do
            echo " $component"
        done
        return 1
    else
        echo ""
        print_success "All kustomize configurations are valid!"
        return 0
    fi
}

# Show usage
show_usage() {
    echo "Usage: $0 [PATH]"
    echo ""
    echo "Validate kustomize configurations"
    echo ""
    echo "ARGUMENTS:"
    echo "  PATH    Optional path to search for kustomizations (default: deployments/k8s)"
    echo ""
    echo "EXAMPLES:"
    echo "  $0                          # Validate all kustomizations"
    echo "  $0 deployments/k8s          # Validate all (skip local secrets issues in CI)"
    echo "  $0 deployments/k8s/workloads/overlays/prod  # Validate only prod workloads"
    echo ""
}

# Main execution
main() {
    local search_path="${1:-deployments/k8s}"

    # Handle help flag
    if [ "$search_path" = "-h" ] || [ "$search_path" = "--help" ]; then
        show_usage
        exit 0
    fi

    # Validate path exists
    if [ ! -d "$search_path" ]; then
        print_error "Error: Path '$search_path' does not exist"
        echo ""
        show_usage
        exit 1
    fi

    check_kustomize

    print_status "Discovering kustomization files in $search_path..."
    echo ""

    # Find all kustomization.yaml files in the specified path
    # Exclude local overlays by default (they require local secrets not available in CI)
    while IFS= read -r -d '' kustomization; do
        kustomization_dir=$(dirname "$kustomization")
        validate_kustomization "$kustomization_dir" || true
    done < <(find "$search_path" -path "*/overlays/local" -prune -o -name "kustomization.yaml" -print0 | sort -z)

    # Print summary and exit with appropriate code
    if print_summary; then
        exit 0
    else
        exit 1
    fi
}

# Run main function
main "$@"
