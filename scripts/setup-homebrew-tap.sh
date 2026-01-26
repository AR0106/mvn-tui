#!/bin/bash
# Script to set up the Homebrew tap repository (run once)

set -e

TAP_REPO="alexritt/homebrew-tap"
TAP_URL="https://github.com/${TAP_REPO}.git"

echo "🍺 Setting up Homebrew Tap for mvn-tui"
echo "======================================="
echo ""

# Check if gh CLI is available
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) is not installed."
    echo "   Install it with: brew install gh"
    echo "   Or manually create the repository at: https://github.com/${TAP_REPO}"
    exit 1
fi

echo "📦 Creating repository: ${TAP_REPO}"
gh repo create "${TAP_REPO}" \
    --public \
    --description "Homebrew tap for mvn-tui" \
    --clone || echo "Repository may already exist, continuing..."

cd homebrew-tap

echo ""
echo "📝 Creating Formula directory and README"
mkdir -p Formula

cat > README.md << 'EOF'
# Homebrew Tap for mvn-tui

This tap contains Homebrew formulas for mvn-tui.

## Installation

```bash
brew install alexritt/tap/mvn-tui
```

## About mvn-tui

Terminal User Interface for Maven - make common workflows fast and discoverable.

### Features

- Interactive module selection
- Module creation with automatic pom.xml updates
- Quick task access for common Maven goals
- Smart run detection (Spring Boot, JAR, WAR)
- Profile management
- Command history and log viewer
- Real-time command execution with cancellation

For more information, visit: https://github.com/alexritt/mvn-tui

## Formulae

- `mvn-tui`: Terminal UI for Maven

---

*This tap is automatically updated by GoReleaser when new versions are released.*
EOF

echo ""
echo "📤 Committing and pushing initial setup"
git add .
git commit -m "Initial tap setup for mvn-tui" || true
git push origin main

cd ..

echo ""
echo "✅ Homebrew tap setup complete!"
echo ""
echo "Next steps:"
echo "  1. Create a release: git tag -a v0.1.0 -m 'Release v0.1.0' && git push origin v0.1.0"
echo "  2. Wait for GitHub Actions to build and publish"
echo "  3. Install with: brew install alexritt/tap/mvn-tui"
echo ""
echo "See RELEASING.md for detailed release instructions."
