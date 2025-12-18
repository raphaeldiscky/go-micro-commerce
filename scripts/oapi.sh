#!/bin/bash

set -euo pipefail

SERVICES=()

for dir in */ ; do
  dir="${dir%/}"
  if [[ -f "$dir/go.mod" ]]; then
    SERVICES+=("$dir")
  fi
done

# Check if required tools are installed
check_dependencies() {
  if ! command -v oapi-codegen &> /dev/null; then
    echo "Error: oapi-codegen is not installed"
    echo "Please run: task install_tools"
    exit 1
  fi
}

# Find OpenAPI spec location (supports both patterns)
find_openapi_spec() {
  local service="$1"

  # New pattern: oapi/spec/openapi.yaml
  if [[ -f "$service/oapi/spec/openapi.yaml" ]]; then
    echo "$service/oapi/spec/openapi.yaml"
    return 0
  fi

  # Legacy pattern: api/openapi.yaml
  if [[ -f "$service/api/openapi.yaml" ]]; then
    echo "$service/api/openapi.yaml"
    return 0
  fi

  return 1
}

# Get output directory based on spec location
get_output_dir() {
  local service="$1"
  local spec_path="$2"

  # New pattern: oapi/spec/ -> oapi/
  if [[ "$spec_path" == *"/oapi/spec/"* ]]; then
    echo "$service/oapi"
  else
    # Legacy pattern: api/ -> internal/generated/
    echo "$service/internal/generated"
  fi
}

# Bundle OpenAPI spec to resolve external references
bundle_openapi_spec() {
  local spec_path="$1"
  local bundled_spec="${spec_path%.yaml}.bundled.yaml"

  # Check if the spec has external references
  if grep -q '\$ref.*\.\/.*\.yaml' "$spec_path" 2>/dev/null; then
    echo "  -> Bundling OpenAPI spec to resolve external references..." >&2

    if command -v npx &> /dev/null; then
      if npx --yes @redocly/cli bundle "$spec_path" -o "$bundled_spec" >&2 2>/dev/null; then
        echo "  -> Successfully bundled OpenAPI spec" >&2
        echo "$bundled_spec"
        return 0
      else
        echo "  -> Warning: Failed to bundle with @redocly/cli" >&2
      fi
    fi

    echo "  -> Error: Could not bundle OpenAPI spec with external references" >&2
    echo "  -> Please install @redocly/cli: npm install -g @redocly/cli" >&2
    return 1
  fi

  echo "$spec_path"
  return 0
}

# Generate OpenAPI code for a single service
generate_service() {
  local service="$1"
  echo "Generating OpenAPI code for $service..."

  # Find OpenAPI spec
  local spec_path
  if ! spec_path=$(find_openapi_spec "$service"); then
    echo "Warning: No OpenAPI spec found for $service, skipping"
    return 0
  fi

  echo "  -> Found spec at: $spec_path"

  # Get output directory
  local output_dir
  output_dir=$(get_output_dir "$service" "$spec_path")
  mkdir -p "$output_dir"

  # Bundle the OpenAPI spec if needed
  local spec_to_use
  if ! spec_to_use=$(bundle_openapi_spec "$spec_path"); then
    echo "Error: Failed to prepare OpenAPI spec for $service"
    return 1
  fi

  # Check for oapi-codegen config
  local config_file=""
  if [[ -f "$service/oapi/oapi-codegen.yaml" ]]; then
    config_file="$service/oapi/oapi-codegen.yaml"
    echo "  -> Using config: $config_file"

    # Generate using config file (new pattern)
    pushd "$service/oapi" > /dev/null
    if ! oapi-codegen -config oapi-codegen.yaml "spec/openapi.yaml"; then
      popd > /dev/null
      echo "Error: Failed to generate code for $service"
      return 1
    fi
    popd > /dev/null
  else
    # Generate using CLI args (legacy pattern)
    echo "  -> Generating types..."
    if ! oapi-codegen -package generated -generate types "$spec_to_use" > "$output_dir/types.go"; then
      echo "Error: Failed to generate types for $service"
      return 1
    fi

    echo "  -> Generating Echo server interface..."
    if ! oapi-codegen -package generated -generate echo-server "$spec_to_use" > "$output_dir/server.go"; then
      echo "Error: Failed to generate server interface for $service"
      return 1
    fi

    echo "  -> Generating embedded spec..."
    if ! oapi-codegen -package generated -generate spec "$spec_to_use" > "$output_dir/spec.go"; then
      echo "Error: Failed to generate embedded spec for $service"
      return 1
    fi
  fi

  echo "  Generated code for $service in $output_dir"

  # Clean up bundled file if it was created
  if [[ "$spec_to_use" == *".bundled.yaml" ]]; then
    echo "  -> Cleaning up bundled spec file"
    rm -f "$spec_to_use"
  fi

  return 0
}

# Preview documentation using Redocly
preview_docs() {
  local service="$1"
  local spec_path

  if ! spec_path=$(find_openapi_spec "$service"); then
    echo "Error: No OpenAPI spec found for $service"
    return 1
  fi

  echo "Starting Redocly preview for $service..."
  echo "  -> Spec: $spec_path"
  echo "  -> URL: http://localhost:8080"
  echo ""

  npx --yes @redocly/cli preview-docs "$spec_path" --port 8080
}

# Main script logic
check_dependencies

# Handle commands
case "${1:-}" in
  preview|docs)
    if [[ -z "${2:-}" ]]; then
      echo "Error: Please specify a service"
      echo "Usage: $0 preview <service>"
      exit 1
    fi
    preview_docs "$2"
    ;;
  *)
    # Default: generate code
    if [[ -n "${1:-}" ]]; then
      if [[ " ${SERVICES[*]} " =~ " $1 " ]]; then
        if generate_service "$1"; then
          echo "OpenAPI code generation complete for $1"
        else
          echo "OpenAPI code generation failed for $1"
          exit 1
        fi
      else
        echo "Error: '$1' is not a valid service directory."
        echo "Available services: ${SERVICES[*]}"
        exit 1
      fi
    else
      failed_services=()

      for service in "${SERVICES[@]}"; do
        if find_openapi_spec "$service" > /dev/null 2>&1; then
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
        exit 1
      fi

      echo "All OpenAPI code generation completed successfully!"
    fi
    ;;
esac
