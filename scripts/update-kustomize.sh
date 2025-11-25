#!/bin/bash

set -e

# Emergency recovery script for manual kustomization updates
# Usage: ./scripts/update-kustomize.sh <version> [--dry-run]

VERSION="$1"
DRY_RUN="${2:-}"

if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version> [--dry-run]"
  echo "Example: $0 v0.1.1"
  echo "Example: $0 v0.1.1 --dry-run"
  exit 1
fi

# Validate version format (should start with 'v')
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
  echo "Error: Version must follow semantic versioning format (e.g., v0.1.1)"
  exit 1
fi

BRANCH="chore/update-kustomize-$VERSION"

# List of all microservices
SERVICES=(
  "api-gateway"
  "auth-service"
  "product-service"
  "order-service"
  "payment-service"
  "fulfillment-service"
  "notification-service"
  "search-service"
  "chat-service"
  "cart-service"
)

echo "========================================"
echo "Kustomization Update Script"
echo "========================================"
echo "Version: $VERSION"
echo "Branch: $BRANCH"
if [ "$DRY_RUN" = "--dry-run" ]; then
  echo "Mode: DRY RUN (no changes will be committed)"
else
  echo "Mode: LIVE (changes will be committed and pushed)"
fi
echo "========================================"
echo ""

# Check if we're in the right directory
if [ ! -d "deployments/k8s/workloads/overlays/prod" ]; then
  echo "Error: This script must be run from the repository root"
  exit 1
fi

# Check if gh CLI is available
if ! command -v gh &> /dev/null; then
  echo "Error: GitHub CLI (gh) is not installed"
  echo "Install it from: https://cli.github.com/"
  exit 1
fi

if [ "$DRY_RUN" != "--dry-run" ]; then
  # Fetch latest changes
  echo "Fetching latest changes from origin..."
  git fetch origin

  # Checkout main and pull latest
  echo "Checking out main branch..."
  git checkout main
  git pull origin main

  # Check if branch already exists
  if git show-ref --verify --quiet "refs/heads/$BRANCH"; then
    echo "Warning: Branch $BRANCH already exists locally"
    read -p "Delete and recreate? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
      git branch -D "$BRANCH"
    else
      echo "Aborted"
      exit 1
    fi
  fi

  # Create feature branch
  echo "Creating branch: $BRANCH"
  git checkout -b "$BRANCH"
  echo ""
fi

# Update all kustomization files
UPDATED_SERVICES=()
echo "Updating kustomization files..."
echo ""

for service in "${SERVICES[@]}"; do
  KUSTOMIZATION_PATH="deployments/k8s/workloads/overlays/prod/$service/kustomization.yaml"

  if [ -f "$KUSTOMIZATION_PATH" ]; then
    echo "Processing $service..."

    # Get current tag
    CURRENT_TAG=$(grep "newTag:" "$KUSTOMIZATION_PATH" | awk '{print $2}')

    if [ "$CURRENT_TAG" = "$VERSION" ]; then
      echo "  ↳ Already at version $VERSION"
    else
      echo "  ↳ Updating from $CURRENT_TAG to $VERSION"

      if [ "$DRY_RUN" != "--dry-run" ]; then
        # Update the newTag field
        sed -i "s/newTag: .*/newTag: $VERSION/" "$KUSTOMIZATION_PATH"
      fi

      UPDATED_SERVICES+=("$service")
    fi
  else
    echo "⚠ Warning: Kustomization file not found for $service"
  fi
done

echo ""
echo "========================================"
echo "Summary"
echo "========================================"
echo "Total services: ${#SERVICES[@]}"
echo "Updated services: ${#UPDATED_SERVICES[@]}"

if [ ${#UPDATED_SERVICES[@]} -eq 0 ]; then
  echo ""
  echo "No services needed updating. All files are already at version $VERSION"
  if [ "$DRY_RUN" != "--dry-run" ]; then
    git checkout main
  fi
  exit 0
fi

echo ""
echo "Updated services list:"
for service in "${UPDATED_SERVICES[@]}"; do
  echo "  - $service"
done
echo ""

if [ "$DRY_RUN" = "--dry-run" ]; then
  echo "DRY RUN - No changes committed"
  exit 0
fi

# Commit changes
echo "Committing changes..."
git add deployments/k8s/workloads/overlays/prod/*/kustomization.yaml

cat > commit_msg.txt << EOF
chore(k8s): update kustomization image tags to $VERSION

Updated image tags for ${#UPDATED_SERVICES[@]} services:
$(printf '%s\n' "${UPDATED_SERVICES[@]}" | sed 's/^/- /')

This is a manual update using the emergency recovery script.

Release: $VERSION
EOF

git commit -F commit_msg.txt
rm commit_msg.txt

echo "Commit created successfully"
echo ""

# Push branch
echo "Pushing branch to origin..."
git push origin "$BRANCH"
echo "Branch pushed successfully"
echo ""

# Create PR
echo "Creating pull request..."

cat > pr_body.md << EOF
## Summary

Manual update of Kustomization image tags for release **$VERSION**.

## Changes

Updated \`newTag\` field in kustomization files for **${#UPDATED_SERVICES[@]}** services:

$(for service in "${UPDATED_SERVICES[@]}"; do
  echo "- \`$service\`: deployments/k8s/workloads/overlays/prod/$service/kustomization.yaml"
done)

## Verification

- [x] All kustomization files updated to version $VERSION
- [ ] ArgoCD sync will pick up these changes automatically

## Metadata

- **Release**: $VERSION
- **Method**: Manual update using emergency recovery script
- **Operator**: \$(git config user.name)

---

*This PR was manually created using the emergency recovery script.*
EOF

gh pr create \
  --title "chore(k8s): update kustomization image tags to $VERSION" \
  --body-file pr_body.md \
  --base main \
  --head "$BRANCH" \
  --label "chore" \
  --label "k8s" \
  --label "manual"

rm pr_body.md

echo ""
echo "========================================"
echo "Success!"
echo "========================================"
echo "Branch: $BRANCH"
echo "Pull request created successfully"
echo ""
echo "Next steps:"
echo "1. Review the PR on GitHub"
echo "2. Merge the PR when ready"
echo "3. ArgoCD will automatically sync the changes"
echo ""
