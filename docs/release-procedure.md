# Release Procedure

Complete guide for releasing a new version of hp.

## Prerequisites

- On the `dev` branch with all changes committed
- `gh` CLI authenticated (`command gh auth status`)
- Write access to the repository

## Release Steps

### 1. Prepare the Release

```bash
# Ensure you're on dev and up to date
git checkout dev
git pull origin dev

# Bump version (choose one)
just bump patch   # 0.7.7 → 0.7.8
just bump minor   # 0.7.7 → 0.8.0
just bump major   # 0.7.7 → 1.0.0
```

### 2. Update Documentation

Update `CHANGELOG.md`:
- Change `[Unreleased]` to `[X.Y.Z] - YYYY-MM-DD`
- Add new `[Unreleased]` section at top (if there will be more changes)

Update `ROADMAP.md`:
- Move completed items from "Ideas" to "Completed" section
- Add version header to completed section

### 3. Commit Release Preparation

```bash
git add main.go CHANGELOG.md ROADMAP.md
git commit -m "chore: prepare vX.Y.Z release"
git push origin dev
```

### 4. Merge to Main via PR

```bash
# Create PR from dev to main
command gh pr create --base main --head dev --title "vX.Y.Z - Description"

# Wait for CI checks to pass (lint, test, CodeQL)
command gh pr checks <PR-NUMBER> --watch

# Merge the PR
# IMPORTANT: Do NOT use --delete-branch (preserves dev)
command gh pr merge <PR-NUMBER> --merge
```

### 5. Create and Push Tag

```bash
# Switch to main and pull the merge
git checkout main
git pull origin main

# Create annotated tag
git tag -a vX.Y.Z -m "vX.Y.Z - Description"

# Push tag to trigger release workflow
git push origin vX.Y.Z
```

### 6. Monitor Release Workflow

```bash
# List recent runs to find the release workflow
command gh run list --limit 3

# Watch the release workflow (get run ID from above)
command gh run watch <RUN-ID>

# Or check status
command gh run view <RUN-ID> --json status,conclusion
```

The release workflow will:
1. Run tests and vulnerability check
2. Build binaries for all platforms (darwin, linux, windows × amd64, arm64)
3. Sign binaries with cosign (OIDC keyless)
4. Create GitHub Release with assets
5. Update Homebrew tap formula
6. Update Scoop bucket manifest
7. Build and push Docker container to GHCR

### 7. Verify Release

```bash
# View the release
command gh release view vX.Y.Z

# Check container image exists
docker manifest inspect ghcr.io/colangelo/hp:vX.Y.Z

# Or pull and test it
docker run --rm ghcr.io/colangelo/hp:vX.Y.Z --version
```

### 8. Post-Release

```bash
# Switch back to dev
git checkout dev

# Sync dev with main
git merge main
git push origin dev

# Optionally set custom release title
command gh release edit vX.Y.Z --title "vX.Y.Z - Custom Title"
```

## Release Artifacts

Each release produces:

| Artifact | Description |
|----------|-------------|
| `hp-darwin-amd64` | macOS Intel binary |
| `hp-darwin-arm64` | macOS Apple Silicon binary |
| `hp-linux-amd64` | Linux x86_64 binary |
| `hp-linux-arm64` | Linux ARM64 binary |
| `hp-windows-amd64.exe` | Windows x86_64 binary |
| `hp-windows-arm64.exe` | Windows ARM64 binary |
| `checksums.txt` | SHA256 checksums for all binaries |
| `*.sig` | Cosign signatures |
| `*.pem` | Cosign certificates |

Container images:
- `ghcr.io/colangelo/hp:vX.Y.Z`
- `ghcr.io/colangelo/hp:latest`

## Package Managers

Releases automatically update:
- **Homebrew**: `brew install colangelo/tap/hp`
- **Scoop**: `scoop bucket add colangelo https://github.com/colangelo/scoop-bucket && scoop install hp`

## Troubleshooting

### Merge conflict when merging dev to main

```bash
# On dev branch
git fetch origin main
git merge origin/main
# Resolve conflicts
git add .
git commit -m "Merge main into dev"
git push origin dev
# Then create/retry the PR
```

### Release workflow failed

```bash
# Check the failed run
command gh run view <run-id>

# Re-run failed jobs
command gh run rerun <run-id>

# Or delete the tag and re-push
git tag -d vX.Y.Z
git push origin :refs/tags/vX.Y.Z
# Fix the issue, then re-tag
git tag -a vX.Y.Z -m "message"
git push origin vX.Y.Z
```

### Wrong version released

```bash
# Delete the release
command gh release delete vX.Y.Z --yes

# Delete the tag
git push origin :refs/tags/vX.Y.Z
git tag -d vX.Y.Z

# Fix version, re-commit, re-tag
```

## Important Notes

- **Never delete the `dev` branch** - it's the main development branch
- **Always merge dev → main via PR** - never push directly to main
- **Use `command gh`** - to avoid shell aliases interfering with gh CLI
- **Cosign signing is automatic** - uses OIDC/keyless signing via GitHub Actions
