#!/bin/bash
# Create Docker Hub imagePullSecret for Kubernetes
# This resolves Docker Hub rate limit issues (429 errors)
# Uses ServiceAccount approach per Kubernetes best practices

set -e

echo "=== Docker Hub imagePullSecret Creator ==="
echo ""
echo "This script creates a Kubernetes secret to authenticate with Docker Hub,"
echo "which increases your pull rate limit from 100 to 200 pulls per 6 hours."
echo ""
echo "The secret will be added to the default ServiceAccount, automatically"
echo "applying to all pods in the namespace without modifying manifests."
echo ""

# Prompt for Docker Hub credentials
read -p "Enter your Docker Hub username: " DOCKER_USERNAME
read -sp "Enter your Docker Hub password or access token: " DOCKER_PASSWORD
echo ""
read -p "Enter your Docker Hub email: " DOCKER_EMAIL

# Validate inputs
if [ -z "$DOCKER_USERNAME" ] || [ -z "$DOCKER_PASSWORD" ] || [ -z "$DOCKER_EMAIL" ]; then
    echo "Error: All fields are required"
    exit 1
fi

echo ""
echo "Creating imagePullSecret in Kubernetes..."

# Create the secret
kubectl create secret docker-registry dockerhub-secret \
    --docker-server=https://index.docker.io/v1/ \
    --docker-username="$DOCKER_USERNAME" \
    --docker-password="$DOCKER_PASSWORD" \
    --docker-email="$DOCKER_EMAIL" \
    --dry-run=client -o yaml | kubectl apply -f -

echo ""
echo "Docker Hub secret created successfully!"
echo ""
echo "Adding secret to default ServiceAccount..."

# Patch the default ServiceAccount to use the secret
kubectl patch serviceaccount default -p '{"imagePullSecrets": [{"name": "dockerhub-secret"}]}'

echo ""
echo "ServiceAccount patched successfully!"
echo ""
echo "Next steps:"
echo "1. The secret 'dockerhub-secret' has been created and added to the default ServiceAccount"
echo "2. All pods in the default namespace will now use authenticated Docker Hub pulls"
echo "3. Delete existing pods to recreate them with the secret:"
echo "   kubectl delete pods -l component=database"
echo "4. Or restart Tilt (Ctrl+C and run 'tilt up' again)"
echo ""
echo "Note: Free Docker Hub accounts get 200 pulls per 6 hours (vs 100 unauthenticated)."
echo "For unlimited pulls, consider upgrading to Docker Pro."
echo ""
echo "Documentation: https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/"
