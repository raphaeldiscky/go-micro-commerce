#!/bin/bash
# Cleanup script for Strimzi Kafka Operator RBAC conflicts
# Run this once before starting Tilt if you encounter rolebindings errors
#
# This script removes all Strimzi-related resources from previous installations
# that might cause conflicts when deploying via Tilt/Helm

set -e

echo "=== Strimzi Kafka Operator Cleanup ==="
echo ""
echo "This script will remove all existing Strimzi resources to resolve RBAC conflicts."
echo ""

# Function to run commands and show status
run_command() {
    local cmd="$1"
    local desc="$2"

    echo "→ $desc"
    if eval "$cmd"; then
        echo "  ✓ Success"
    else
        echo "  ⚠ Warning: Command failed (this may be normal if resource doesn't exist)"
    fi
    echo ""
}

# 1. Delete Strimzi Helm release from default namespace (old location)
run_command \
    "helm uninstall strimzi-operator --namespace default --ignore-not-found 2>/dev/null || true" \
    "Uninstalling Strimzi operator Helm release from default namespace"

# 2. Delete strimzi-system namespace entirely (new dedicated namespace)
run_command \
    "kubectl delete namespace strimzi-system --ignore-not-found=true --timeout=60s" \
    "Deleting strimzi-system namespace (if exists)"

# 3. Delete RoleBindings with strimzi label from all namespaces
run_command \
    "kubectl delete rolebindings -l app.kubernetes.io/name=strimzi --all-namespaces --ignore-not-found=true" \
    "Deleting Strimzi rolebindings from all namespaces"

# 4. Delete specific rolebindings by name (from old installations)
run_command \
    "kubectl delete rolebindings strimzi-cluster-operator strimzi-cluster-operator-watched strimzi-cluster-operator-entity-operator-delegation --ignore-not-found=true -n default" \
    "Deleting specific Strimzi rolebindings from default namespace"

# 5. Delete ClusterRoleBindings
run_command \
    "kubectl delete clusterrolebindings -l app.kubernetes.io/name=strimzi --ignore-not-found=true" \
    "Deleting Strimzi clusterrolebindings"

# 6. Delete ServiceAccounts
run_command \
    "kubectl delete serviceaccounts -l app.kubernetes.io/name=strimzi --all-namespaces --ignore-not-found=true" \
    "Deleting Strimzi serviceaccounts"

# 7. Delete Kafka CRDs (optional - uncomment if you want a complete clean slate)
# WARNING: This will delete all Kafka cluster configurations!
# run_command \
#     "kubectl delete crd kafkas.kafka.strimzi.io kafkanodepools.kafka.strimzi.io --ignore-not-found=true" \
#     "Deleting Kafka CRDs (WARNING: removes all Kafka cluster configs)"

echo "=== Cleanup Complete ==="
echo ""
echo "✓ All Strimzi RBAC resources have been removed"
echo ""
echo "Next steps:"
echo "1. Start Tilt: tilt up"
echo "2. The operator will be installed in the strimzi-system namespace"
echo "3. It will watch the default namespace for Kafka CRDs"
echo ""
echo "Note: The new configuration follows Kubernetes operator best practices:"
echo "  - Operator runs in: strimzi-system namespace"
echo "  - Operator watches: default namespace"
echo "  - This prevents RBAC conflicts"
echo ""
