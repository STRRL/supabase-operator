# Quickstart: Automated Helm Chart Release

This guide helps you set up automated Helm chart releases for the Supabase Operator.

## Prerequisites

- GitHub repository with Helm charts in `charts/` directory
- GitHub account with permissions to:
  - Push tags to source repository
  - Create Personal Access Token
  - Access target repository (STRRL/helm.strrl.dev)

## Setup Steps

### 1. Create Personal Access Token

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Name: `HELM_REPO_TOKEN`
4. Expiration: 90 days (or your preference)
5. Select scopes:
   - `repo` (full control of private repositories)
   - Or `public_repo` if both repositories are public
6. Generate token and copy it (you won't see it again!)

### 2. Add Token to Repository Secrets

1. Navigate to your repository: https://github.com/STRRL/supabase-operator
2. Go to Settings → Secrets and variables → Actions
3. Click "New repository secret"
4. Name: `HELM_REPO_TOKEN`
5. Value: Paste your Personal Access Token
6. Click "Add secret"

### 3. Create GitHub Actions Workflow

Create file `.github/workflows/release-helm.yaml`:

```yaml
name: Release Helm Chart

on:
  push:
    tags:
      - 'v[0-9]+.[0-9]+.[0-9]+'

permissions:
  contents: write

concurrency:
  group: helm-release
  cancel-in-progress: false

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Extract version
        id: version
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          echo "version=$VERSION" >> $GITHUB_OUTPUT

      - name: Update Chart.yaml
        run: |
          sed -i "s/^version:.*/version: ${{ steps.version.outputs.version }}/" charts/supabase-operator/Chart.yaml
          sed -i "s/^appVersion:.*/appVersion: \"${{ steps.version.outputs.version }}\"/" charts/supabase-operator/Chart.yaml

      - name: Release Chart
        uses: stefanprodan/helm-gh-pages@master
        with:
          token: ${{ secrets.HELM_REPO_TOKEN }}
          charts_dir: charts
          owner: STRRL
          repository: helm.strrl.dev
          branch: gh-pages
          charts_url: https://helm.strrl.dev
          linting: on
```

### 4. Prepare Target Repository

Ensure the target repository exists and is configured:

1. Repository `STRRL/helm.strrl.dev` must exist
2. Initialize gh-pages branch:
   ```bash
   git clone https://github.com/STRRL/helm.strrl.dev
   cd helm.strrl.dev
   git checkout --orphan gh-pages
   git rm -rf .
   echo "# Helm Charts" > README.md
   git add README.md
   git commit -m "Initialize gh-pages"
   git push origin gh-pages
   ```

3. Enable GitHub Pages:
   - Go to repository Settings → Pages
   - Source: Deploy from a branch
   - Branch: `gh-pages` / `/ (root)`
   - Save

### 5. Trigger Your First Release

1. Commit and push the workflow file:
   ```bash
   git add .github/workflows/release-helm.yaml
   git commit -m "Add Helm chart release workflow"
   git push origin main
   ```

2. Create and push a version tag:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

3. Monitor the release:
   - Go to Actions tab in your repository
   - Watch the "Release Helm Chart" workflow
   - Check for green checkmark

4. Verify the release:
   - Visit https://helm.strrl.dev
   - Check for index.yaml and chart package

## Usage

### For Maintainers: Creating a Release

```bash
# 1. Ensure you're on main branch with latest changes
git checkout main
git pull origin main

# 2. Create and push a semantic version tag
git tag v1.2.3
git push origin v1.2.3

# 3. Monitor workflow in GitHub Actions
# The chart will be available within 5 minutes
```

### For Users: Installing the Chart

```bash
# Add the Helm repository
helm repo add supabase-operator https://helm.strrl.dev
helm repo update

# Search for available versions
helm search repo supabase-operator --versions

# Install the operator
helm install my-supabase supabase-operator/supabase-operator \
  --namespace supabase-system \
  --create-namespace

# Upgrade to a new version
helm upgrade my-supabase supabase-operator/supabase-operator \
  --namespace supabase-system
```

## Troubleshooting

### Workflow Fails with Permission Denied

**Issue**: Token doesn't have required permissions
**Solution**: Ensure PAT has `repo` scope and write access to target repository

### Chart Not Appearing in Repository

**Issue**: Workflow succeeded but chart not visible
**Solution**:
1. Check gh-pages branch exists in target repo
2. Verify GitHub Pages is enabled
3. Wait for GitHub Pages deployment (can take 10 minutes)

### Version Conflict

**Issue**: Version already exists
**Solution**: Charts are immutable - use a new version number

### Lint Failures

**Issue**: Helm lint step fails
**Solution**:
1. Run `helm lint charts/supabase-operator` locally
2. Fix reported issues
3. Commit fixes and create new tag

### Concurrent Release Issues

**Issue**: Multiple tags pushed simultaneously
**Solution**: Workflow has concurrency control - releases will queue automatically

## Best Practices

1. **Test Locally First**
   ```bash
   helm lint charts/supabase-operator
   helm package charts/supabase-operator
   ```

2. **Use Semantic Versioning**
   - MAJOR.MINOR.PATCH (e.g., 1.2.3)
   - Increment MAJOR for breaking changes
   - Increment MINOR for new features
   - Increment PATCH for bug fixes

3. **Document Changes**
   - Update Chart.yaml annotations with changelog
   - Create GitHub release with notes

4. **Monitor Releases**
   - Watch GitHub Actions for failures
   - Subscribe to workflow notifications
   - Test installation after release

## Security Notes

- Personal Access Token expires - set calendar reminder for renewal
- Never commit tokens in code
- Rotate tokens if compromised
- Consider using fine-grained PATs for better security

## Next Steps

- [ ] Set up chart signing with GPG
- [ ] Add installation tests to workflow
- [ ] Register chart on ArtifactHub
- [ ] Configure release notifications
- [ ] Add changelog generation

## Links

- [Workflow File](.github/workflows/release-helm.yaml)
- [Chart Repository](https://helm.strrl.dev)
- [GitHub Actions Runs](https://github.com/STRRL/supabase-operator/actions)
- [Helm Documentation](https://helm.sh/docs/)