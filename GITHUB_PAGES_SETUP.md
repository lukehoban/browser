# GitHub Pages Setup Guide

This guide explains how to enable GitHub Pages deployment for the browser WASM demo.

## Prerequisites

- GitHub repository with the browser code
- Push access to the repository

## Setup Steps

### 1. Enable GitHub Pages

1. Go to your repository on GitHub
2. Click on **Settings** (gear icon)
3. In the left sidebar, click **Pages**
4. Under "Build and deployment":
   - **Source**: Select **GitHub Actions**
5. Save the settings

### 2. Verify Workflow Permissions

The workflow needs specific permissions to deploy to GitHub Pages. These are already configured in `.github/workflows/pages.yml`:

```yaml
permissions:
  contents: read
  pages: write
  id-token: write
```

If you encounter permission errors, verify these are set correctly.

### 3. Trigger the Deployment

The deployment can be triggered in two ways:

#### Automatic (on push to main)
```bash
git push origin main
```

#### Manual (from GitHub UI)
1. Go to the **Actions** tab in your repository
2. Select the "Deploy to GitHub Pages" workflow
3. Click **Run workflow**
4. Select the `main` branch
5. Click **Run workflow**

### 4. Access the Demo

After the workflow completes (usually 1-2 minutes):

1. Go back to **Settings** → **Pages**
2. You'll see: "Your site is live at https://[username].github.io/[repository]/"
3. Click the link or visit directly

For this repository: **https://lukehoban.github.io/browser/**

## Troubleshooting

### Workflow Fails with "Permission denied"

**Solution**: Ensure GitHub Actions has permission to deploy to Pages:
1. Settings → Actions → General
2. Scroll to "Workflow permissions"
3. Select "Read and write permissions"
4. Check "Allow GitHub Actions to create and approve pull requests"
5. Save

### Pages Not Enabled

**Error**: "GitHub Pages is not enabled for this repository"

**Solution**: Follow Step 1 above to enable GitHub Pages with "GitHub Actions" as the source.

### Deployment Successful but Page Shows 404

**Solution**: 
1. Check that `index.html` exists in the `wasm/` directory
2. Wait a few minutes for GitHub's CDN to update
3. Try accessing with a cache-busting parameter: `?v=1`
4. Clear your browser cache

### WASM Module Fails to Load

**Solution**:
1. Check browser console for errors
2. Ensure `browser.wasm` and `wasm_exec.js` are in the same directory as `index.html`
3. Check that files are being served with correct MIME types (GitHub Pages handles this automatically)

## Workflow Details

The `.github/workflows/pages.yml` workflow:

1. **Triggers**: On push to `main` or manual trigger
2. **Build Job**: 
   - Checks out code
   - Sets up Go 1.23
   - Builds WASM module: `GOOS=js GOARCH=wasm go build`
   - Uploads the `wasm/` directory as artifact
3. **Deploy Job**:
   - Deploys artifact to GitHub Pages
   - Outputs the deployment URL

## Custom Domain (Optional)

To use a custom domain:

1. Settings → Pages → Custom domain
2. Enter your domain (e.g., `browser.example.com`)
3. Follow DNS setup instructions
4. Wait for DNS propagation (can take up to 48 hours)

## Monitoring

View deployment status:
1. **Actions** tab shows all workflow runs
2. **Environments** → **github-pages** shows deployment history
3. Each deployment includes the URL it was deployed to

## Updating the Demo

Any push to `main` automatically rebuilds and redeploys:

```bash
git add .
git commit -m "Update WASM demo"
git push origin main
```

The new version will be live in 1-2 minutes.
