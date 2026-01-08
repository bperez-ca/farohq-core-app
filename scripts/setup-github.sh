#!/bin/bash

set -e

echo "🚀 Setting up GitHub repository..."
echo ""

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) is not installed."
    echo "Install it from: https://cli.github.com/"
    echo ""
    echo "Alternatively, you can create the repo manually:"
    echo "1. Go to https://github.com/new"
    echo "2. Create a new repository named 'farohq-core-app'"
    echo "3. Run: git remote add origin https://github.com/YOUR_USERNAME/farohq-core-app.git"
    echo "4. Run: git push -u origin main"
    exit 1
fi

# Check if user is logged in
if ! gh auth status &> /dev/null; then
    echo "❌ Not logged in to GitHub CLI."
    echo "Run: gh auth login"
    exit 1
fi

# Get repository name
REPO_NAME="${1:-farohq-core-app}"
echo "📦 Creating repository: $REPO_NAME"

# Create repository on GitHub
gh repo create "$REPO_NAME" \
    --public \
    --description "FaroHQ Core Application - Go microservices for multi-tenant SaaS platform" \
    --source=. \
    --remote=origin \
    --push

echo ""
echo "✅ Repository created and code pushed!"
echo "🌐 View at: https://github.com/$(gh api user --jq .login)/$REPO_NAME"
