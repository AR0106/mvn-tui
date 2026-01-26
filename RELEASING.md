# Releasing mvn-tui

This document describes how to create a new release of mvn-tui.

## Prerequisites

1. You have push access to the repository
2. The code is tested and ready for release
3. The CHANGELOG or relevant documentation is updated

## Release Process

### 1. Choose a Version Number

Follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backwards-compatible)
- **PATCH**: Bug fixes (backwards-compatible)

Example: `v1.0.0`, `v1.1.0`, `v1.1.1`

### 2. Create and Push a Git Tag

```bash
# Make sure you're on the main branch and up to date
git checkout main
git pull

# Create a tag with the version number (include the 'v' prefix)
git tag -a v1.0.0 -m "Release v1.0.0"

# Push the tag to GitHub
git push origin v1.0.0
```

### 3. Automated Release

Once you push the tag, GitHub Actions will automatically:

1. **Build binaries** for:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)

2. **Create a GitHub Release** with:
   - Release notes
   - Downloadable archives for each platform
   - Checksums file

3. **Update Homebrew Tap** at `github.com/AR0106/homebrew-tap`
   - Automatically creates/updates the formula
   - Users can install with: `brew install AR0106/tap/mvn-tui`

### 4. Verify the Release

1. Go to https://github.com/AR0106/mvn-tui/releases
2. Verify the release was created with all binaries
3. Check the Homebrew tap repository was updated
4. Test installation:
   ```bash
   # On macOS/Linux
   brew install AR0106/tap/mvn-tui
   mvn-tui --version
   ```

## First Release Setup (One-Time)

Before creating the first release, you need to:

### 1. Create Homebrew Tap Repository

```bash
# Create a new repository on GitHub named: homebrew-tap
# (Use GitHub's web interface or gh CLI)
gh repo create AR0106/homebrew-tap --public --description "Homebrew tap for mvn-tui"
```

### 2. Initialize the Tap Repository

```bash
git clone https://github.com/AR0106/homebrew-tap.git
cd homebrew-tap

# Create Formula directory
mkdir -p Formula

# Create initial README
cat > README.md << 'EOF'
# Homebrew Tap for mvn-tui

This tap contains Homebrew formulas for mvn-tui.

## Installation

```bash
brew install AR0106/tap/mvn-tui
```

## About mvn-tui

Terminal User Interface for Maven - make common workflows fast and discoverable.

For more information, visit: https://github.com/AR0106/mvn-tui
EOF

git add .
git commit -m "Initial tap setup"
git push origin main
```

### 3. Verify GitHub Token Permissions

The GitHub Actions workflow uses `GITHUB_TOKEN` which is automatically provided. Ensure your repository has:
- **Contents**: Write permission (for creating releases)
- **Actions**: Read permission

You can check this in: Repository Settings → Actions → General → Workflow permissions

## Troubleshooting

### Release Failed

Check the GitHub Actions logs:
1. Go to the "Actions" tab in GitHub
2. Click on the failed workflow run
3. Review the logs to see what went wrong

Common issues:
- **Tag already exists**: Delete and recreate the tag
- **Homebrew tap not accessible**: Verify the tap repository exists and is public
- **Build failures**: Check Go version compatibility and dependencies

### Manual Release with GoReleaser

If you need to test locally or manually create a release:

```bash
# Install GoReleaser
brew install goreleaser

# Test the release build (doesn't publish)
goreleaser release --snapshot --clean

# Create a real release (requires GITHUB_TOKEN)
export GITHUB_TOKEN="your-github-token"
goreleaser release --clean
```

### Updating the Homebrew Formula Manually

If automatic updates fail, you can manually update the formula:

```bash
cd homebrew-tap
cd Formula

# GoReleaser should have created mvn-tui.rb
# Edit if needed, then commit and push
git add mvn-tui.rb
git commit -m "Update mvn-tui to vX.Y.Z"
git push
```

## Release Checklist

- [ ] Code is tested and working
- [ ] Version number chosen (semantic versioning)
- [ ] Documentation updated (README, CHANGELOG, etc.)
- [ ] All tests pass: `go test ./...`
- [ ] Tag created and pushed: `git tag -a vX.Y.Z -m "Release vX.Y.Z" && git push origin vX.Y.Z`
- [ ] GitHub Actions workflow completed successfully
- [ ] GitHub release created with binaries
- [ ] Homebrew tap updated
- [ ] Installation tested: `brew install AR0106/tap/mvn-tui`
- [ ] Version command works: `mvn-tui --version`

## Version History

Document major releases here:

### v1.0.0 (TBD)
- Initial release
- Core Maven TUI functionality
- Module creation with automatic pom.xml updates
- Real command execution with cancellation
- Homebrew distribution
