#!/bin/bash
set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source common functions
source "${SCRIPT_DIR}/common.sh"

print_info "=== PHASE 1: Talos Cluster + CNI Deployment ==="
echo ""

# Check prerequisites
print_info "Checking prerequisites..."

if ! command_exists terraform; then
    print_error "Terraform is not installed. Please install Terraform >= 1.13.5"
    exit 1
fi

if ! command_exists talosctl; then
    print_error "talosctl is not installed. Please install talosctl"
    print_info "Install with: curl -sL https://talos.dev/install | sh"
    exit 1
fi

# Check Terraform version
TERRAFORM_VERSION=$(terraform version -json | grep -o '"terraform_version":"[^"]*"' | cut -d'"' -f4)
print_success "Terraform version: $TERRAFORM_VERSION"

# Check talosctl version
TALOSCTL_VERSION=$(talosctl version --client --short)
print_success "talosctl version: $TALOSCTL_VERSION"

# Check if terraform.tfvars exists
if [ ! -f "terraform.tfvars" ]; then
    print_error "terraform.tfvars not found!"
    print_info "Please create terraform.tfvars from terraform.tfvars.example"
    print_info "cp terraform.tfvars.example terraform.tfvars"
    print_info "Then edit terraform.tfvars with your actual values"
    exit 1
fi

# Install helm if not present (required for CNI deployment)
if ! command_exists helm; then
    print_info "Installing Helm..."
    curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
    chmod 700 get_helm.sh
    sudo ./get_helm.sh
    rm get_helm.sh
fi

# Install kubectl if not present (only needed for Helm operations and final tests)
if ! command_exists kubectl; then
    print_info "Installing kubectl..."
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
    chmod +x kubectl
    sudo mv kubectl /usr/local/bin/
fi

print_success "All prerequisites met"
echo ""

# Create dummy kubeconfig if it doesn't exist (prevents provider initialization errors)
if [ ! -f "./kubeconfig-production" ]; then
    print_info "Creating placeholder kubeconfig for provider initialization..."
    cat > ./kubeconfig-production << 'EOF'
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: placeholder
contexts:
- context:
    cluster: placeholder
    user: placeholder
  name: placeholder
current-context: placeholder
preferences: {}
users:
- name: placeholder
  user:
    token: placeholder
EOF
    chmod 600 ./kubeconfig-production
    print_success "Placeholder kubeconfig created (will be replaced by real config)"
fi
echo ""

# Initialize Terraform
print_info "Initializing Terraform..."
terraform init
print_success "Terraform initialized"
echo ""

# Validate configuration
print_info "Validating Terraform configuration..."
terraform validate
print_success "Configuration is valid"
echo ""

# Show plan for Phase 1
print_info "=== Deploying Talos Cluster ==="
print_info "This will create the Talos cluster and generate kubeconfig/talosconfig"
echo ""

read -p "Do you want to see the plan first? (y/n) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    terraform plan -target=module.talos_cluster
    echo ""
    read -p "Proceed with deployment? (y/n) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Deployment cancelled by user"
        exit 0
    fi
fi

# Deploy Talos cluster
print_info "Deploying Talos cluster..."
terraform apply -target=module.talos_cluster -auto-approve

if [ $? -ne 0 ]; then
    print_error "Talos cluster deployment failed. Please check the errors above."
    exit 1
fi

print_success "Talos cluster deployed successfully"
echo ""

# Check if kubeconfig was created
if [ ! -f "./kubeconfig-production" ]; then
    print_error "Kubeconfig file not found at ./kubeconfig-production"
    print_error "Cluster may not be ready yet"
    exit 1
fi

print_success "Kubeconfig generated at ./kubeconfig-production"
echo ""

# Check if talosconfig was created
if [ ! -f "./talosconfig-production" ]; then
    print_error "Talosconfig file not found at ./talosconfig-production"
    print_error "Terraform may not have generated it properly"
    exit 1
fi

print_success "Talosconfig generated at ./talosconfig-production"
echo ""

# Wait for Kubernetes API to become minimally responsive
print_info "Waiting for Kubernetes API to become responsive..."
print_info "Note: API won't be fully functional until CNI is installed"
export KUBECONFIG=./kubeconfig-production
export TALOSCONFIG=./talosconfig-production

API_MAX_WAIT=300  # 5 minutes
API_WAIT_INTERVAL=10
API_ELAPSED=0
API_READY=false

while [ $API_ELAPSED -lt $API_MAX_WAIT ]; do
    # Try to connect to API - we just need ANY response, even an error
    if kubectl version --short 2>&1 | grep -q "Server Version\|error\|refused\|timeout"; then
        # If we get any response (including errors), the API endpoint is at least reachable
        if kubectl cluster-info 2>&1 | grep -q "Kubernetes\|running"; then
            print_success "Kubernetes API is responding"
            API_READY=true
            break
        elif kubectl version --short 2>/dev/null | grep -q "Server Version"; then
            print_success "Kubernetes API is minimally responsive"
            API_READY=true
            break
        fi
    fi

    print_info "API not ready yet, waiting... (${API_ELAPSED}s/${API_MAX_WAIT}s)"
    sleep $API_WAIT_INTERVAL
    API_ELAPSED=$((API_ELAPSED + API_WAIT_INTERVAL))
done

if [ "$API_READY" = false ]; then
    print_warning "Kubernetes API did not become responsive after ${API_MAX_WAIT} seconds"
    print_info "This is normal for Talos - API needs CNI to be fully functional"
    print_info "Proceeding with CNI installation anyway..."
fi
echo ""

# Install CNI (Cilium)
print_info "=== Installing CNI (Cilium) ==="
print_info "Installing Container Network Interface for pod networking..."
echo ""

# Add Cilium Helm repository
helm repo add cilium https://helm.cilium.io/
helm repo update

# Check for existing Cilium installation and remove it if present
if helm list -n kube-system 2>/dev/null | grep -q cilium; then
    print_warning "Existing Cilium installation found. Uninstalling..."
    helm uninstall cilium -n kube-system
    print_info "Waiting for Cilium pods to terminate..."

    # Wait for pods to be deleted
    print_info "Checking pod termination status..."
    for i in {1..30}; do
        POD_COUNT=$(kubectl get pods -n kube-system --no-headers 2>/dev/null | grep "cilium" | wc -l)
        if [ "$POD_COUNT" -eq 0 ]; then
            print_success "All Cilium pods terminated"
            break
        fi
        echo -n "."
        sleep 1
    done
    echo ""
    sleep 5
fi

# Install Cilium with Talos-specific configuration
print_info "Deploying Cilium CNI with Talos-specific settings..."
helm install cilium cilium/cilium --namespace kube-system \
  --set autoDirectNodeRoutes=true \
  --set enableBandwidthManager=true \
  --set bpf.masquerade=true \
  --set kubeProxyReplacement=true \
  --set securityContext.privileged=true \
  --set cni.chainingMode=portmap \
  --set hostServices.enabled=false \
  --set hostServices.protocols=tcp \
  --set enableIPv4=true \
  --set enableIPv6=false \
  --set operator.replicas=1

# Wait for Cilium pods to be ready
print_info "Waiting for Cilium pods to be ready..."
export TALOSCONFIG=./talosconfig-production

# Wait for Cilium to be deployed and ready (longer timeout for Talos)
print_info "This may take 2-3 minutes on Talos..."
print_info "Monitoring with talosctl..."
echo ""

# Poll for ready pods with timeout
MAX_WAIT=180
WAIT_INTERVAL=10
ELAPSED=0

while [ $ELAPSED -lt $MAX_WAIT ]; do
    RUNNING_COUNT=$(kubectl get pods -n kube-system --no-headers 2>/dev/null | grep "cilium" | grep "Running" | wc -l)
    # Ensure RUNNING_COUNT is a valid number
    if ! [[ "$RUNNING_COUNT" =~ ^[0-9]+$ ]]; then
        RUNNING_COUNT=0
    fi
    if [ "$RUNNING_COUNT" -gt 0 ]; then
        print_info "Found $RUNNING_COUNT Cilium pod(s) running after ${ELAPSED}s"
        break
    fi
    print_info "Waiting for Cilium pods... (${ELAPSED}s/${MAX_WAIT}s)"
    sleep $WAIT_INTERVAL
    ELAPSED=$((ELAPSED + WAIT_INTERVAL))
done

# Check if pods are ready using kubectl
CILIUM_STATUS=$(kubectl get pods -n kube-system --no-headers 2>/dev/null | grep "cilium" | grep "Running" | wc -l)
# Ensure CILIUM_STATUS is a valid number
if ! [[ "$CILIUM_STATUS" =~ ^[0-9]+$ ]]; then
    CILIUM_STATUS=0
fi
if [ "$CILIUM_STATUS" -eq 0 ]; then
    print_warning "Cilium pods are taking longer than expected. Checking status..."

    # Use kubectl for pod status
    print_info "Pod status via kubectl:"
    kubectl get pods -n kube-system -l k8s-app=cilium 2>/dev/null || print_warning "Could not get pod status"

    echo ""
    # Get pod details
    print_info "Pod details:"
    kubectl describe pods -n kube-system -l k8s-app=cilium 2>/dev/null | grep -A 10 "Events:" || true

    echo ""
    # Get logs if pods exist
    print_info "Cilium logs (if available):"
    kubectl logs -n kube-system -l k8s-app=cilium --tail=20 2>/dev/null || print_warning "Logs not available yet"

    echo ""
    print_info "For detailed troubleshooting, you can use:"
    print_info "  kubectl get pods -n kube-system -l k8s-app=cilium"
    print_info "  kubectl describe pods -n kube-system -l k8s-app=cilium"
    print_info "  kubectl logs -n kube-system -l k8s-app=cilium"

    read -p "Continue anyway? (y/n) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_error "CNI installation aborted"
        exit 1
    fi
else
    print_success "Cilium pods are running (${CILIUM_STATUS} pods ready)"
fi
echo ""

# Verify CNI functionality using talosctl and kubectl
print_info "Verifying CNI functionality..."

# Use talosctl for cluster member status
print_info "Cluster members via talosctl:"
talosctl get members 2>/dev/null || print_warning "Could not get cluster members"

echo ""
# Use kubectl for node status (more reliable)
print_info "Node status via kubectl:"
export KUBECONFIG=./kubeconfig-production
kubectl get nodes 2>/dev/null || print_warning "Nodes not ready yet (expected without CNI)"

echo ""
# Test DNS resolution (still needs kubectl for pod networking)
print_info "Testing DNS resolution (using kubectl for final validation)..."
export KUBECONFIG=./kubeconfig-production
kubectl run dns-test --image=busybox:1.35 --rm -i --restart=Never -- nslookup kubernetes.default.svc.cluster.local || {
    print_warning "DNS test failed, but continuing. This may resolve once all pods are ready."
}

print_success "CNI (Cilium) installed successfully"
echo ""
print_info "Monitoring commands:"
print_info "  # Check Talos cluster:"
print_info "  export TALOSCONFIG=./talosconfig-production"
print_info "  talosctl get members"
print_info "  talosctl health"
print_info ""
print_info "  # Check Kubernetes cluster:"
print_info "  export KUBECONFIG=./kubeconfig-production"
print_info "  kubectl get nodes"
print_info "  kubectl get pods -n kube-system -l k8s-app=cilium"
print_info "  kubectl get pods -n kube-system"
echo ""

# Create marker file to indicate Phase 1 completion
touch .phase1-complete
print_success "=== PHASE 1 COMPLETED SUCCESSFULLY ==="
echo ""
print_info "Configuration files generated:"
echo "   - ./kubeconfig-production (Kubernetes access)"
echo "   - ./talosconfig-production (Talos management)"
echo ""
print_info "Next step: Run Phase 2 to deploy Kubernetes resources"
echo "   ./deploy-phase2.sh"
echo ""
print_warning "IMPORTANT: Do not commit config files to version control!"
