#!/bin/bash

set -euo pipefail

SERVICES=(
  "auth-service"
  "notification-service"
  "order-service"
  "product-service"
  "payment-service"
  "fulfillment-service"
  "search-service"
  "chat-service"
)

# Check if required tools are installed
check_dependencies() {
  if ! command -v oapi-codegen &> /dev/null; then
    echo "Error: oapi-codegen is not installed"
    echo "Please run: task install_tools"
    exit 1
  fi
}

# Bundle OpenAPI spec to resolve external references
bundle_openapi_spec() {
  local service="$1"
  local api_spec="$service/api/openapi.yaml"
  local bundled_spec="$service/api/bundled.yaml"

  # Check if the spec has external references
  if grep -q '\$ref.*\.\/.*\.yaml' "$api_spec" 2>/dev/null; then
    echo "  -> Bundling OpenAPI spec to resolve external references..." >&2

    # Try to use @redocly/cli to bundle the spec
    if command -v npx &> /dev/null; then
      if npx --yes @redocly/cli bundle "$api_spec" -o "$bundled_spec" >&2 2>/dev/null; then
        echo "  -> Successfully bundled OpenAPI spec" >&2
        echo "$bundled_spec"
        return 0
      else
        echo "  -> Warning: Failed to bundle with @redocly/cli, trying alternatives..." >&2
      fi
    fi

    # Fallback: try swagger-codegen-cli
    if command -v npx &> /dev/null; then
      if npx --yes swagger-codegen-cli generate -i "$api_spec" -l openapi-yaml -o "/tmp/${service}-bundle" 2>/dev/null; then
        if [[ -f "/tmp/${service}-bundle/openapi.yaml" ]]; then
          cp "/tmp/${service}-bundle/openapi.yaml" "$bundled_spec"
          rm -rf "/tmp/${service}-bundle"
          echo "  -> Successfully bundled OpenAPI spec with swagger-codegen" >&2
          echo "$bundled_spec"
          return 0
        fi
      fi
    fi

    echo "  -> Error: Could not bundle OpenAPI spec with external references" >&2
    echo "  -> Please install @redocly/cli: npm install -g @redocly/cli" >&2
    echo "  -> Or consider consolidating your OpenAPI spec into a single file" >&2
    return 1
  fi

  # No external references, use original file
  echo "$api_spec"
  return 0
}

# Generate OpenAPI code for a single service
generate_service() {
  local service="$1"
  echo "Generating OpenAPI code for $service..."

  # Check if service directory exists
  if [[ ! -d "$service" ]]; then
    echo "Error: Service directory '$service' does not exist"
    return 1
  fi

  # Check if OpenAPI spec exists
  local api_spec="$service/api/openapi.yaml"
  if [[ ! -f "$api_spec" ]]; then
    echo "Warning: OpenAPI spec not found at '$api_spec', skipping $service"
    return 0
  fi

  # Create output directory
  local output_dir="$service/internal/generated"
  mkdir -p "$output_dir"

  # Bundle the OpenAPI spec if needed
  local spec_to_use
  if ! spec_to_use=$(bundle_openapi_spec "$service"); then
    echo "Error: Failed to prepare OpenAPI spec for $service"
    return 1
  fi

  # Generate Go types
  echo "  -> Generating types..."
  if ! oapi-codegen -package generated -generate types "$spec_to_use" > "$output_dir/types.go"; then
    echo "Error: Failed to generate types for $service"
    return 1
  fi

  # Generate Echo server interface
  echo "  -> Generating Echo server interface..."
  if ! oapi-codegen -package generated -generate echo-server "$spec_to_use" > "$output_dir/server.go"; then
    echo "Error: Failed to generate server interface for $service"
    return 1
  fi

  # Generate embedded spec
  echo "  -> Generating embedded spec..."
  if ! oapi-codegen -package generated -generate spec "$spec_to_use" > "$output_dir/spec.go"; then
    echo "Error: Failed to generate embedded spec for $service"
    return 1
  fi

  echo "  Generated code for $service in $output_dir"

  # Clean up bundled file if it was created
  if [[ "$spec_to_use" == *"/bundled.yaml" ]]; then
    echo "  -> Cleaning up bundled spec file"
    rm -f "$spec_to_use"
  fi

  return 0
}

# Check dependencies first
check_dependencies

# If a specific service is provided as an argument
if [ -n "${1-}" ]; then
  if [[ " ${SERVICES[*]} " =~ " $1 " ]]; then
    if generate_service "$1"; then
      echo "OpenAPI code generation complete for $1"
    else
      echo "OpenAPI code generation failed for $1"
      echo ""
      echo "Tips for fixing external reference issues:"
      echo "   1. Install bundling tool: npm install -g @redocly/cli"
      echo "   2. Or consolidate your OpenAPI spec into a single file"
      exit 1
    fi
  else
    echo "Error: '$1' is not a valid service directory."
    echo "Available services: ${SERVICES[*]}"
    exit 1
  fi
# If no argument, run for all services
else
  failed_services=()

  for service in "${SERVICES[@]}"; do
    # Only process services that have OpenAPI specs
    if [[ -f "$service/api/openapi.yaml" ]]; then
      if ! generate_service "$service"; then
        failed_services+=("$service")
      fi
    else
      echo "Skipping $service (no OpenAPI spec found)"
    fi
  done

  if [ ${#failed_services[@]} -gt 0 ]; then
    echo ""
    echo "OpenAPI code generation failed for: ${failed_services[*]}"
    echo ""
    echo "Tips for fixing external reference issues:"
    echo "   1. Install bundling tool: npm install -g @redocly/cli"
    echo "   2. Or consolidate OpenAPI specs into single files for failed services"
    exit 1
  fi

  echo "All OpenAPI code generation completed successfully!"
fi