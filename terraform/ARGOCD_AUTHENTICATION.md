# ArgoCD Git Repository Authentication Guide

This guide explains how to configure ArgoCD to access your private Git repository using either HTTPS (Personal Access Token) or SSH (Deploy Key) authentication.

## Table of Contents

- [Overview](#overview)
- [Method 1: HTTPS with Personal Access Token](#method-1-https-with-personal-access-token)
- [Method 2: SSH with Deploy Key](#method-2-ssh-with-deploy-key)
- [Comparison](#comparison)
- [Troubleshooting](#troubleshooting)
- [Security Best Practices](#security-best-practices)

## Overview

ArgoCD needs credentials to access your private GitHub repository to:

- Read ApplicationSet manifests from `deployments/k8s/apps/applicationsets/`
- Read infrastructure manifests from `deployments/k8s/infrastructure/`
- Read workload overlays from `deployments/k8s/workloads/overlays/prod/`
- Monitor repository changes for GitOps synchronization

You can use either HTTPS or SSH authentication. **Both methods are secure when configured properly.**

## Method 1: HTTPS with Personal Access Token

### When to Use

- ✅ Quick setup for development/testing
- ✅ Simpler configuration
- ✅ Works across multiple repositories
- ⚠️ Token has broad GitHub account permissions

### Prerequisites

- GitHub account with access to the repository
- Terraform installed and configured

### Step 1: Create GitHub Personal Access Token

1. Go to GitHub Settings: https://github.com/settings/tokens
2. Click **"Generate new token"** -> **"Generate new token (classic)"**
3. Configure the token:
   - **Note**: `ArgoCD GKE Production`
   - **Expiration**: Choose based on your security policy (e.g., 90 days)
   - **Scopes**: Select `repo` (Full control of private repositories)
4. Click **"Generate token"**
5. **IMPORTANT**: Copy the token immediately (format: `ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`)

### Step 2: Update Repository URL

Edit `terraform/environments/prod/terraform.tfvars`:

```hcl
argocd_git_repo_url = "https://github.com/raphaeldiscky/go-micro-commerce.git"
```

### Step 3: Configure Terraform Variables

**Option A: Environment Variables (Recommended)**

```bash
# Set credentials
export TF_VAR_argocd_git_username="git"
export TF_VAR_argocd_git_token="ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# Navigate to terraform directory
cd terraform/environments/prod

# Apply
terraform init
terraform plan
terraform apply
```

**Option B: Pass at Runtime**

```bash
cd terraform/environments/prod

terraform apply \
  -var="argocd_git_username=git" \
  -var="argocd_git_token=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

### Step 4: Verify Configuration

```bash
# Check if Secret was created
kubectl get secret git-repo-credentials -n argocd

# View Secret contents (base64 encoded)
kubectl get secret git-repo-credentials -n argocd -o yaml

# Check ArgoCD detected the repository
kubectl exec -n argocd deploy/argocd-server -- argocd repo list
```

### Step 5: Update ApplicationSets (if needed)

Ensure ApplicationSet files use HTTPS URLs:

```yaml
# deployments/k8s/apps/applicationsets/infrastructure.yaml
# deployments/k8s/apps/applicationsets/workloads.yaml
spec:
  generators:
    - git:
        repoURL: https://github.com/raphaeldiscky/go-micro-commerce.git
```

---

## Method 2: SSH with Deploy Key

### When to Use

- ✅ **Recommended for production**
- ✅ More secure (repository-specific access)
- ✅ Easy credential rotation
- ✅ No expiration (unless you set one)
- ⚠️ Requires SSH URL format in all manifests

### Prerequisites

- SSH client installed (`ssh-keygen` command available)
- GitHub repository admin access (to add deploy key)
- Terraform installed and configured

### Step 1: Generate SSH Key Pair

```bash
# Generate ED25519 key pair (more secure and performant than RSA)
ssh-keygen -t ed25519 -C "argocd@discky.com" -f ~/.ssh/argocd-repo -N ""

# This creates:
# - ~/.ssh/argocd-repo (private key) - Keep this SECRET
# - ~/.ssh/argocd-repo.pub (public key) - Safe to share
```

### Step 2: Add Public Key to GitHub

**Option A: Deploy Key (Recommended - Repository-Specific)**

1. Go to your repository: https://github.com/raphaeldiscky/go-micro-commerce
2. Navigate to: **Settings** -> **Deploy keys**
3. Click **"Add deploy key"**
4. Configure:
   - **Title**: `ArgoCD Production GKE`
   - **Key**: Paste the output of:
     ```bash
     cat ~/.ssh/argocd-repo.pub
     ```
   - **Allow write access**: ❌ Leave unchecked (read-only is sufficient)
5. Click **"Add key"**

**Option B: SSH Key (Account-Wide - Less Secure)**

1. Go to: https://github.com/settings/keys
2. Click **"New SSH key"**
3. Configure:
   - **Title**: `ArgoCD GKE Production`
   - **Key type**: `Authentication Key`
   - **Key**: Paste the output of:
     ```bash
     cat ~/.ssh/argocd-repo.pub
     ```
4. Click **"Add SSH key"**

### Step 3: Update Repository URL to SSH Format

Edit `terraform/environments/prod/terraform.tfvars`:

```hcl
# Change from HTTPS:
# argocd_git_repo_url = "https://github.com/raphaeldiscky/go-micro-commerce.git"

# To SSH:
argocd_git_repo_url = "git@github.com:raphaeldiscky/go-micro-commerce.git"
```

### Step 4: Configure Terraform with Private Key

**Option A: Environment Variable (Recommended)**

```bash
# Set private key from file
export TF_VAR_argocd_git_ssh_private_key=$(cat ~/.ssh/argocd-repo)

# Verify it's set (should show the key)
echo "$TF_VAR_argocd_git_ssh_private_key" | head -2

# Apply Terraform
cd terraform/environments/prod
terraform init
terraform plan
terraform apply
```

**Option B: Pass at Runtime**

```bash
cd terraform/environments/prod

terraform apply -var="argocd_git_ssh_private_key=$(cat ~/.ssh/argocd-repo)"
```

### Step 5: Update ApplicationSets to SSH URLs

Edit both ApplicationSet files to use SSH URLs:

**File: `deployments/k8s/apps/applicationsets/infrastructure.yaml`**

```yaml
spec:
  generators:
    - git:
        # Change from HTTPS to SSH
        repoURL: git@github.com:raphaeldiscky/go-micro-commerce.git
        revision: main
  template:
    spec:
      source:
        repoURL: git@github.com:raphaeldiscky/go-micro-commerce.git
```

**File: `deployments/k8s/apps/applicationsets/workloads.yaml`**

```yaml
spec:
  generators:
    - git:
        repoURL: git@github.com:raphaeldiscky/go-micro-commerce.git
        revision: main
  template:
    spec:
      source:
        repoURL: git@github.com:raphaeldiscky/go-micro-commerce.git
```

### Step 6: Test SSH Connection

```bash
# Test SSH authentication
ssh -T git@github.com -i ~/.ssh/argocd-repo

# Expected output:
# Hi raphaeldiscky! You've successfully authenticated, but GitHub does not provide shell access.
```

### Step 7: Verify Configuration

```bash
# Check if Secret was created with SSH key
kubectl get secret git-repo-credentials -n argocd -o yaml

# Should contain:
# data:
#   sshPrivateKey: <base64-encoded-key>
#   type: Z2l0  # "git" in base64
#   url: <base64-encoded-ssh-url>

# Check ArgoCD repositories
kubectl exec -n argocd deploy/argocd-server -- argocd repo list
```

---

## Comparison

| Feature              | HTTPS (Token)            | SSH (Deploy Key)           |
| -------------------- | ------------------------ | -------------------------- |
| **Setup Complexity** | Simple                   | Moderate                   |
| **Security**         | Good (token-based)       | Better (key-based)         |
| **Scope**            | Account-wide             | Repository-specific        |
| **Expiration**       | Yes (configurable)       | No (unless you set)        |
| **Rotation**         | Generate new token       | Generate new key pair      |
| **Best For**         | Development/Testing      | Production                 |
| **URL Format**       | `https://github.com/...` | `git@github.com:...`       |
| **Write Access**     | Based on token scope     | Based on deploy key config |

---

## Troubleshooting

### HTTPS Authentication Issues

#### Error: "Authentication failed"

```bash
# Verify token is set correctly
echo "$TF_VAR_argocd_git_token" | cut -c1-10  # Should show "ghp_xxxxxx"

# Check Secret exists
kubectl get secret git-repo-credentials -n argocd

# View Secret (decode to verify)
kubectl get secret git-repo-credentials -n argocd -o jsonpath='{.data.password}' | base64 -d

# Test manually
git clone https://git:<YOUR_TOKEN>@github.com/raphaeldiscky/go-micro-commerce.git /tmp/test-clone
```

#### Error: "Repository not found"

- Verify the repository URL is correct
- Ensure the token has `repo` scope
- Check if repository is actually private

### SSH Authentication Issues

#### Error: "Permission denied (publickey)"

```bash
# Verify SSH key was added to GitHub
ssh -T git@github.com -i ~/.ssh/argocd-repo -v

# Check key format
head -1 ~/.ssh/argocd-repo  # Should be: -----BEGIN OPENSSH PRIVATE KEY-----

# Verify public key on GitHub matches
cat ~/.ssh/argocd-repo.pub
```

#### Error: "Host key verification failed"

```bash
# Add GitHub to known hosts
ssh-keyscan github.com >> ~/.ssh/known_hosts

# Or accept interactively
ssh -T git@github.com
```

#### Secret exists but ArgoCD can't connect

```bash
# Check ArgoCD repo server logs
kubectl logs -n argocd -l app.kubernetes.io/name=argocd-repo-server --tail=50

# Check ApplicationSet controller logs
kubectl logs -n argocd -l app.kubernetes.io/name=argocd-applicationset-controller --tail=50

# Manually test from within ArgoCD pod
kubectl exec -n argocd deploy/argocd-repo-server -- ssh -T git@github.com
```

### General Issues

#### No applications appearing in ArgoCD

```bash
# Check if bootstrap ApplicationSet exists
kubectl get applicationset -n argocd

# Check bootstrap ApplicationSet status
kubectl describe applicationset bootstrap-applicationsets -n argocd

# Check if repository is accessible
kubectl exec -n argocd deploy/argocd-server -- argocd repo list

# Force refresh
kubectl delete applicationset bootstrap-applicationsets -n argocd
terraform apply  # Will recreate it
```

---

## Security Best Practices

### For Both Methods

1. **Never commit credentials to git**
   - Add `terraform.tfvars` to `.gitignore`
   - Use environment variables or secret management systems

2. **Use environment variables**

   ```bash
   export TF_VAR_argocd_git_token="..."        # HTTPS
   export TF_VAR_argocd_git_ssh_private_key="..."  # SSH
   ```

3. **Rotate credentials regularly**
   - HTTPS: Regenerate tokens every 90 days
   - SSH: Regenerate keys annually or when compromised

4. **Use least privilege**
   - HTTPS: Token only needs `repo` scope
   - SSH: Deploy key should be read-only

5. **Monitor access**
   - Check GitHub audit log for unusual access
   - Review ArgoCD sync activity regularly

### HTTPS-Specific

- Set token expiration (don't use "No expiration")
- Create separate tokens for different environments
- Revoke old tokens immediately after rotation

### SSH-Specific

- Use ED25519 keys (more secure than RSA)
- Protect private key with file permissions:
  ```bash
  chmod 600 ~/.ssh/argocd-repo
  ```
- Use deploy keys (not personal SSH keys)
- Don't allow write access unless necessary

---

## Additional Resources

- [ArgoCD Private Repositories Documentation](https://argo-cd.readthedocs.io/en/stable/user-guide/private-repositories/)
- [GitHub Personal Access Tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)
- [GitHub Deploy Keys](https://docs.github.com/en/developers/overview/managing-deploy-keys)
- [Terraform Environment Variables](https://www.terraform.io/docs/cli/config/environment-variables.html#tf_var_name)
