# Cloudflare Pages Setup Guide

This guide walks you through deploying the React + Vite frontend to Cloudflare Pages with automatic deployments from GitHub.

## Why Cloudflare Pages?

✅ **Free hosting**: Unlimited bandwidth and requests
✅ **Automatic deployments**: Push to GitHub -> auto-deploy
✅ **Global CDN**: 300+ edge locations worldwide
✅ **Built-in SPA routing**: No 404 redirect configuration needed
✅ **Preview deployments**: Every PR gets a preview URL
✅ **Zero infrastructure**: No GCS buckets or load balancers to manage

## Prerequisites

Before you start, ensure you have:

1. **GitHub repository** with the `frontend/` directory
2. **Cloudflare account** (free tier is sufficient)
3. **Domain added to Cloudflare**: `discky.com` must be active in your Cloudflare account
4. **Backend infrastructure deployed**: GKE cluster with Traefik LoadBalancer running

## Step 1: Get Your Backend API URL

After deploying your Terraform infrastructure, get the Traefik LoadBalancer IP:

```bash
cd terraform/environments/prod
terraform output traefik_load_balancer_ip
```

Note this IP address - you'll use it to configure Cloudflare DNS and environment variables.

## Step 2: Create Cloudflare Pages Project

1. **Log in to Cloudflare Dashboard**
   - Visit: https://dash.cloudflare.com/
   - Navigate to: **Workers & Pages** -> **Create application** -> **Pages** -> **Connect to Git**

2. **Connect GitHub Repository**
   - Click **Connect GitHub**
   - Authorize Cloudflare to access your repositories
   - Select the `go-micro-commerce` repository

3. **Configure Build Settings**

   **Framework preset**: Select "None" (we'll configure manually)

   **Build configuration**:
   - **Production branch**: `main`
   - **Build command**:
     ```bash
     pnpm run codegen && pnpm run build
     ```
   - **Build output directory**: `frontend/`
   - **Root directory**: `dist` (leave as repo root)

   **Environment variables** (click "Add variable"):

   ```bash
   VITE_APP_TITLE=Go Micro Commerce
   VITE_API_GATEWAY_URL=https://api.discky.com
   VITE_GRAPHQL_GATEWAY_URL=https://api.discky.com/graph
   VITE_GRAPHQL_SUBSCRIPTION_WS_URL=wss://api.discky.com/graph/subscriptions/ws
   VITE_GRAPHQL_SUBSCRIPTION_SSE_URL=https://api.discky.com/graph/subscriptions/sse
   VITE_STRIPE_PUBLISHABLE_KEY=pk_live_YOUR_LIVE_KEY
   ```

   > **Note**: Replace `api.discky.com` with your actual API subdomain, and update the Stripe key with your production key.

4. **Save and Deploy**
   - Click **Save and Deploy**
   - First deployment will take 3-5 minutes
   - You'll get a temporary URL like: `go-micro-commerce.pages.dev`

## Step 3: Configure Custom Domain

1. **Add Custom Domain**
   - In Cloudflare Pages project settings
   - Go to **Custom domains** -> **Set up a custom domain**
   - Enter: `go.micro.commerce.discky.com`
   - Click **Continue**

2. **DNS Configuration** (Automatic)
   - Cloudflare will automatically create a CNAME record
   - CNAME: `go.micro.commerce` -> `go-micro-commerce.pages.dev`
   - SSL/TLS certificate is provisioned automatically
   - Wait 2-5 minutes for DNS propagation

3. **Verify Domain**
   - Visit: https://go.micro.commerce.discky.com
   - You should see your frontend application
   - Certificate should be valid (check padlock icon)

## Step 4: Configure Backend API DNS

The backend API DNS is managed by Terraform, but you need to verify it's working:

1. **Check DNS Record Created**

   ```bash
   dig api.discky.com +short
   # Should return your Traefik LoadBalancer IP
   ```

2. **Test API Health**

   ```bash
   curl https://api.discky.com/health
   # Should return API health status
   ```

3. **If DNS doesn't resolve**, wait 5-10 minutes for Cloudflare propagation

## Step 5: Environment-Specific Configuration (Optional)

### Preview Deployments

Cloudflare Pages automatically creates preview deployments for:

- **Pull requests**: Each PR gets a unique URL
- **Branch deployments**: Each branch can have its own deployment

To configure different API URLs for previews:

1. Go to **Settings** -> **Environment variables**
2. Switch to **Preview** tab
3. Add preview-specific variables:
   ```bash
   VITE_API_GATEWAY_URL=https://api-staging.discky.com
   ```

### Production vs Preview Environments

| Environment    | API URL                        | Use Case                  |
| -------------- | ------------------------------ | ------------------------- |
| **Production** | https://api.discky.com         | Main branch, live traffic |
| **Preview**    | https://api-staging.discky.com | PRs, testing, staging     |

## Step 6: Automatic Deployment Workflow

Once configured, deployments happen automatically:

1. **Developer pushes to GitHub**

   ```bash
   git commit -m "feat: update homepage design"
   git push origin main
   ```

2. **Cloudflare Pages detects the push**
   - Triggers automatic build
   - Runs: `cd frontend && pnpm install && pnpm build`
   - Compiles React + Vite application

3. **Build succeeds -> Deploy**
   - Deploys `frontend/dist` to global CDN
   - Available at: https://go.micro.commerce.discky.com
   - Previous version is kept for instant rollback

4. **Notifications**
   - GitHub commit status is updated
   - Can configure webhooks for Slack/Discord

**Total time**: ~2-3 minutes from push to live

## Deployment Logs and Monitoring

### View Build Logs

1. Go to your Cloudflare Pages project
2. Click on a deployment
3. View build logs in real-time
4. Debug any build failures

### Common Issues

**Issue**: Frontend shows API errors
**Solution**: Verify `VITE_API_GATEWAY_URL` points to correct backend URL

**Issue**: Custom domain not working
**Solution**: Wait 5 minutes for DNS propagation, check CNAME record exists

### Monitoring

Cloudflare Pages provides:

- **Analytics**: Page views, bandwidth, requests
- **Real-time logs**: Build and deployment logs
- **Performance metrics**: Build time, deployment time
- **Uptime**: 99.99% SLA on free tier

## Advanced Configuration

### Headers and Redirects

Create `frontend/public/_headers` for custom headers:

```
# Security headers
/*
  X-Frame-Options: DENY
  X-Content-Type-Options: nosniff
  Referrer-Policy: strict-origin-when-cross-origin
  Permissions-Policy: geolocation=(), microphone=(), camera=()

# Cache static assets for 1 year
/assets/*
  Cache-Control: public, max-age=31536000, immutable
```

Create `frontend/public/_redirects` for custom redirects:

```
# SPA fallback (handled automatically, but can customize)
/*  /index.html  200

# Redirect old paths
/old-page  /new-page  301
```

### Build Watch Paths

By default, Cloudflare Pages watches the entire repository. To deploy only when frontend changes:

1. Go to **Settings** -> **Builds & deployments**
2. Under **Build watch paths**, add:
   ```
   frontend/**
   ```

This prevents deployments when only backend code changes.

### Branch Deployments

Enable deployments for specific branches:

1. **Settings** -> **Builds & deployments** -> **Branch deployments**
2. Enable: **All non-production branches**
3. Each branch gets URL: `feature-name.go-micro-commerce.pages.dev`

## Rollback Procedure

If a deployment causes issues:

1. Go to **Deployments** in Cloudflare Pages
2. Find previous working deployment
3. Click **...** -> **Rollback to this deployment**
4. Instant rollback (< 30 seconds)

## Cost and Limits

### Free Tier

- **500 builds/month** (more than enough for most teams)
- **Unlimited requests**
- **Unlimited bandwidth**
- **1 concurrent build** (subsequent builds queue)

### Pro Tier ($20/month)

- **5,000 builds/month**
- **5 concurrent builds**
- **Advanced build configuration**
- **Preview deployments on all branches**

For most projects, **free tier is sufficient**.

## Integration with Terraform

Terraform manages:

- ✅ GKE cluster and infrastructure
- ✅ Backend API DNS (`api.discky.com`)
- ❌ Frontend hosting (Cloudflare Pages)
- ❌ Frontend DNS (managed by Cloudflare Pages automatically)

This separation is intentional:

- **Infrastructure as Code**: Terraform for stable infrastructure
- **Application Deployment**: Cloudflare Pages for rapid iteration

## Support and Troubleshooting

### Resources

- **Cloudflare Pages Docs**: https://developers.cloudflare.com/pages/
- **Vite Build Guide**: https://vitejs.dev/guide/build.html
- **Community Discord**: https://discord.cloudflare.com

### Common Commands

**Trigger manual build:**

```bash
# Push empty commit to trigger rebuild
git commit --allow-empty -m "trigger build"
git push
```

**Local build test:**

```bash
cd frontend
pnpm install
pnpm build
pnpm preview  # Test production build locally
```

**Check DNS propagation:**

```bash
dig go.micro.commerce.discky.com
nslookup go.micro.commerce.discky.com
```

## Next Steps

After frontend is deployed:

1. **Configure monitoring**: Set up error tracking (Sentry, etc.)
2. **Set up CI/CD**: Add automated tests before deployment
3. **Enable caching**: Configure edge caching for API responses
4. **Add analytics**: Google Analytics, Plausible, etc.

---

**Questions?** Check the Terraform README or open an issue in the repository.
