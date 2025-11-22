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
failed_helm_lint=()
failed_helm_template=()
failed_kubeconform=()
failed_kubelinter=()
successful_charts=()

# Check if required tools are installed
check_tools() {
    local missing_tools=()

    if ! command -v helm &> /dev/null; then
        missing_tools+=("helm")
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

        if [[ " ${missing_tools[*]} " =~ " helm " ]]; then
            echo "helm:"
            echo "  brew install helm  # macOS"
            echo "  curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash  # Linux"
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

# Validate a single Helm chart with four-layer validation
validate_chart() {
    local chart_dir=$1
    local relative_path="${chart_dir#./}"
    local validation_failed=false

    print_status "Validating Helm chart: $relative_path"

    # Create temporary file for rendered manifests
    local temp_manifest=$(mktemp)
    trap "rm -f $temp_manifest" RETURN

    # Layer 1: Helm Lint
    echo "  [1/4] Helm lint..."
    local helm_lint_output
    if helm_lint_output=$(helm lint "$chart_dir" 2>&1); then
        echo "  ✓ Helm lint passed"
    else
        print_error "  ✗ Helm lint failed"
        echo "$helm_lint_output" | sed 's/^/    /'
        failed_helm_lint+=("$relative_path")
        validation_failed=true
        return 1
    fi

    # Layer 2: Helm Template
    echo "  [2/4] Helm template rendering..."
    local helm_template_output
    if helm_template_output=$(helm template "$chart_dir" 2>&1 > "$temp_manifest"); then
        echo "  ✓ Helm template rendered successfully"
    else
        print_error "  ✗ Helm template failed"
        echo "$helm_template_output" | sed 's/^/    /'
        failed_helm_template+=("$relative_path")
        validation_failed=true
        return 1
    fi

    # Layer 3: Kubeconform Schema Validation
    echo "  [3/4] Kubeconform schema validation..."
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

    # Layer 4: Kube-linter Best Practices
    echo "  [4/4] Kube-linter best practices..."
    local kubelinter_output
    if kubelinter_output=$(kube-linter lint --config .kube-linter.yaml "$temp_manifest" 2>&1); then
        echo "  ✓ Kube-linter checks passed"
    else
        print_error "  ✗ Kube-linter found issues"
        echo "$kubelinter_output" | sed 's/^/    /'
        failed_kubelinter+=("$relative_path")
        validation_failed=true
    fi

    if [ "$validation_failed" = false ]; then
        print_success "$relative_path validated successfully"
        successful_charts+=("$relative_path")
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
    if [ ${#successful_charts[@]} -gt 0 ]; then
        print_success "Successfully validated ${#successful_charts[@]} Helm chart(s)"
    fi

    # Failed validations by type
    local has_failures=false

    if [ ${#failed_helm_lint[@]} -gt 0 ]; then
        echo ""
        print_error "Helm lint failures (${#failed_helm_lint[@]}):"
        for chart in "${failed_helm_lint[@]}"; do
            echo "  - $chart"
        done
        has_failures=true
    fi

    if [ ${#failed_helm_template[@]} -gt 0 ]; then
        echo ""
        print_error "Helm template failures (${#failed_helm_template[@]}):"
        for chart in "${failed_helm_template[@]}"; do
            echo "  - $chart"
        done
        has_failures=true
    fi

    if [ ${#failed_kubeconform[@]} -gt 0 ]; then
        echo ""
        print_warning "Kubeconform schema validation issues (${#failed_kubeconform[@]}):"
        for chart in "${failed_kubeconform[@]}"; do
            echo "  - $chart"
        done
        has_failures=true
    fi

    if [ ${#failed_kubelinter[@]} -gt 0 ]; then
        echo ""
        print_warning "Kube-linter best practice warnings (${#failed_kubelinter[@]}):"
        for chart in "${failed_kubelinter[@]}"; do
            echo "  - $chart"
        done
        echo ""
        print_warning "Note: Linter warnings are informational and don't fail the build"
    fi

    echo ""
    echo "============================================================"

    if [ "$has_failures" = true ]; then
        return 1
    else
        print_success "All Helm chart validations passed!"
        return 0
    fi
}

# Show usage
show_usage() {
    echo "Usage: $0 [PATH]"
    echo ""
    echo "Validate Helm charts with comprehensive validation"
    echo ""
    echo "Four-layer validation:"
    echo "  1. Helm lint - Chart structure and best practices"
    echo "  2. Helm template - Template rendering validation"
    echo "  3. Kubeconform - Schema validation against K8s API"
    echo "  4. Kube-linter - Best practices and security checks"
    echo ""
    echo "ARGUMENTS:"
    echo "  PATH    Optional path to search for Helm charts (default: deployments/helm)"
    echo ""
    echo "EXAMPLES:"
    echo "  $0                          # Validate all Helm charts"
    echo "  $0 deployments/helm         # Validate charts in specific path"
    echo "  $0 deployments/helm/my-app  # Validate specific chart"
    echo ""
}

# Main execution
main() {
    local search_path="${1:-deployments/helm}"

    # Handle help flag
    if [ "$search_path" = "-h" ] || [ "$search_path" = "--help" ]; then
        show_usage
        exit 0
    fi

    # Check if path exists
    if [ ! -d "$search_path" ]; then
        print_warning "Path '$search_path' does not exist"
        print_warning "No Helm charts found - skipping validation"
        echo ""
        print_success "Chart validation skipped (no charts to validate)"
        exit 0
    fi

    # Check required tools
    check_tools

    print_status "Discovering Helm charts in $search_path..."
    echo ""

    # Find all Chart.yaml files
    local charts_found=false
    while IFS= read -r -d '' chart_file; do
        charts_found=true
        chart_dir=$(dirname "$chart_file")
        validate_chart "$chart_dir" || true
    done < <(find "$search_path" -name "Chart.yaml" -print0 | sort -z)

    if [ "$charts_found" = false ]; then
        print_warning "No Helm charts found in $search_path"
        print_success "Chart validation skipped (no charts to validate)"
        exit 0
    fi

    # Print summary and exit with appropriate code
    if print_summary; then
        exit 0
    else
        exit 1
    fi
}

# Run main function
main "$@"
