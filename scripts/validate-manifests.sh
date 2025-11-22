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
    echo -e "${BLUE}▶ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Variables for tracking validation results
failed_kustomize=()
failed_kubeconform=()
failed_kubelinter=()
successful_components=()

# Check if required tools are installed
check_tools() {
    local missing_tools=()

    if ! command -v kustomize &> /dev/null; then
        missing_tools+=("kustomize")
    fi

    if ! command -v kubeconform &> /dev/null; then
        missing_tools+=("kubeconform")
    fi

    if ! command -v kube-linter &> /dev/null; then
        missing_tools+=("kube-linter")
    fi

    if [ ${#missing_tools[@]} -gt 0 ]; then
        print_error "Missing required tools: ${missing_tools[*]}"
        echo ""
        echo "Install missing tools:"
        echo ""

        if [[ " ${missing_tools[*]} " =~ " kustomize " ]]; then
            echo "kustomize:"
            echo "  brew install kustomize  # macOS"
            echo "  curl -s https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh | bash  # Linux"
            echo ""
        fi

        if [[ " ${missing_tools[*]} " =~ " kubeconform " ]]; then
            echo "kubeconform:"
            echo "  go install github.com/yannh/kubeconform/cmd/kubeconform@latest"
            echo "  # OR download binary:"
            echo "  curl -L https://github.com/yannh/kubeconform/releases/latest/download/kubeconform-linux-amd64.tar.gz | tar xz"
            echo ""
        fi

        if [[ " ${missing_tools[*]} " =~ " kube-linter " ]]; then
            echo "kube-linter:"
            echo "  go install golang.stackrox.io/kube-linter/cmd/kube-linter@latest"
            echo "  # OR download binary:"
            echo "  curl -L https://github.com/stackrox/kube-linter/releases/latest/download/kube-linter-linux.tar.gz | tar xz"
            echo ""
        fi

        exit 1
    fi
}

# Validate a single kustomization with three-layer validation
validate_kustomization() {
    local kustomization_dir=$1
    local relative_path="${kustomization_dir#./}"
    local validation_failed=false

    print_status "Validating $relative_path"

    # Create temporary file for manifests
    local temp_manifest=$(mktemp)
    trap "rm -f $temp_manifest" RETURN

    # Layer 1: Kustomize Build
    echo "  [1/3] Kustomize build..."
    local kustomize_output
    if kustomize_output=$(kustomize build "$kustomization_dir" 2>&1 > "$temp_manifest"); then
        echo "  ✓ Kustomize build successful"
    else
        print_error "  ✗ Kustomize build failed"
        echo "$kustomize_output" | sed 's/^/    /'
        failed_kustomize+=("$relative_path")
        validation_failed=true
        return 1
    fi

    # Layer 2: Kubeconform Schema Validation
    echo "  [2/3] Kubeconform schema validation..."
    local kubeconform_output
    if kubeconform_output=$(kubeconform \
        -strict \
        -ignore-missing-schemas \
        -schema-location default \
        -schema-location 'https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json' \
        -summary \
        -output json \
        "$temp_manifest" 2>&1); then
        echo "  ✓ Kubeconform validation passed"
    else
        print_warning "  ⚠ Kubeconform validation issues"
        echo "$kubeconform_output" | sed 's/^/    /'
        failed_kubeconform+=("$relative_path")
        validation_failed=true
    fi

    # Layer 3: Kube-linter Best Practices
    echo "  [3/3] Kube-linter best practices..."
    local kubelinter_output
    if kubelinter_output=$(kube-linter lint "$temp_manifest" 2>&1); then
        echo "  ✓ Kube-linter checks passed"
    else
        print_warning "  ⚠ Kube-linter found issues"
        echo "$kubelinter_output" | sed 's/^/    /'
        failed_kubelinter+=("$relative_path")
        # Don't mark as failed - linter warnings are informational
    fi

    if [ "$validation_failed" = false ]; then
        print_success "$relative_path validated successfully"
        successful_components+=("$relative_path")
        echo ""
        return 0
    else
        echo ""
        return 1
    fi
}

# Print validation summary
print_summary() {
    echo ""
    echo "==================== VALIDATION SUMMARY ===================="
    echo ""

    # Successful validations
    if [ ${#successful_components[@]} -gt 0 ]; then
        print_success "Successfully validated ${#successful_components[@]} kustomization(s)"
    fi

    # Failed validations by type
    local has_failures=false

    if [ ${#failed_kustomize[@]} -gt 0 ]; then
        echo ""
        print_error "Kustomize build failures (${#failed_kustomize[@]}):"
        for component in "${failed_kustomize[@]}"; do
            echo "  - $component"
        done
        has_failures=true
    fi

    if [ ${#failed_kubeconform[@]} -gt 0 ]; then
        echo ""
        print_warning "Kubeconform schema validation issues (${#failed_kubeconform[@]}):"
        for component in "${failed_kubeconform[@]}"; do
            echo "  - $component"
        done
        has_failures=true
    fi

    if [ ${#failed_kubelinter[@]} -gt 0 ]; then
        echo ""
        print_warning "Kube-linter best practice warnings (${#failed_kubelinter[@]}):"
        for component in "${failed_kubelinter[@]}"; do
            echo "  - $component"
        done
        echo ""
        print_warning "Note: Linter warnings are informational and don't fail the build"
    fi

    echo ""
    echo "============================================================"

    if [ "$has_failures" = true ]; then
        return 1
    else
        print_success "All manifest validations passed!"
        return 0
    fi
}

# Show usage
show_usage() {
    echo "Usage: $0 [PATH]"
    echo ""
    echo "Validate Kubernetes manifests using Kustomize"
    echo ""
    echo "Three-layer validation:"
    echo "  1. Kustomize build - Syntax and structure"
    echo "  2. Kubeconform - Schema validation against K8s API"
    echo "  3. Kube-linter - Best practices and security checks"
    echo ""
    echo "ARGUMENTS:"
    echo "  PATH    Optional path to search for kustomizations (default: deployments/k8s)"
    echo ""
    echo "EXAMPLES:"
    echo "  $0                          # Validate all kustomizations"
    echo "  $0 deployments/k8s          # Validate all (skip local overlays)"
    echo "  $0 deployments/k8s/workloads/overlays/prod  # Validate only prod"
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

    # Check required tools
    check_tools

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
