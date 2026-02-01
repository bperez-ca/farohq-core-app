# GitHub Repository Setup

## Quick Setup (Using GitHub CLI)

If you have GitHub CLI installed:

```bash
./scripts/setup-github.sh
```

Or specify a custom repository name:

```bash
./scripts/setup-github.sh my-repo-name
```

## Manual Setup

### 1. Create Repository on GitHub

1. Go to https://github.com/new
2. Repository name: `farohq-core-app` (or your preferred name)
3. Description: "FaroHQ Core Application - Go microservices for multi-tenant SaaS platform"
4. Choose visibility (Public/Private)
5. **Do NOT** initialize with README, .gitignore, or license (we already have these)
6. Click "Create repository"

### 2. Add Remote and Push

```bash
# Add remote (replace YOUR_USERNAME with your GitHub username)
git remote add origin https://github.com/YOUR_USERNAME/farohq-core-app.git

# Or if using SSH:
git remote add origin git@github.com:YOUR_USERNAME/farohq-core-app.git

# Push to GitHub
git push -u origin main
```

### 3. Verify

Visit: `https://github.com/YOUR_USERNAME/farohq-core-app`

## Install GitHub CLI (Optional)

If you don't have `gh` CLI installed:

**macOS:**
```bash
brew install gh
gh auth login
```

**Linux:**
```bash
# See: https://cli.github.com/manual/installation
```

**Windows:**
```powershell
# See: https://cli.github.com/manual/installation
```

## Next Steps

After pushing to GitHub:

1. **Set up GitHub Actions** (if needed)
2. **Add branch protection rules** (if needed)
3. **Add collaborators** (if needed)
4. **Configure secrets** for CI/CD (if needed)
