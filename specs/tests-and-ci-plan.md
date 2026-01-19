# Plan: Add Tests and GitHub Actions CI/Release

## Overview

Add unit tests and GitHub Actions workflow for CI and automated releases.

---

## Phase 1: Add Unit Tests

### File: `main_test.go`

Test these pure functions:

- `getEnvInt()` - env var parsing with defaults
- `getURLForProto()` - URL scheme selection
- `getBlock()` - latency → block character mapping (edge cases: min, thresholds, max)

```go
// Example test structure
func TestGetEnvInt(t *testing.T) { ... }
func TestGetURLForProto(t *testing.T) { ... }
func TestGetBlock(t *testing.T) { ... }
```

**Note:** `getBlock()` uses package-level vars (`minLatency`, `greenThreshold`, `yellowThreshold`). Tests will need to set/restore these.

---

## Phase 2: GitHub Actions Workflow

### File: `.github/workflows/ci.yml`

**Triggers:**

- `push` to any branch → run tests
- `pull_request` → run tests

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go test -v ./...
```

### File: `.github/workflows/release.yml`

**Trigger:** Tags matching `v*`

**Build matrix:**

| OS | Arch | Variant | Output |
|----|------|---------|--------|
| darwin | amd64 | default | hp-darwin-amd64 |
| darwin | amd64 | http3 | hp-http3-darwin-amd64 |
| darwin | arm64 | default | hp-darwin-arm64 |
| darwin | arm64 | http3 | hp-http3-darwin-arm64 |
| linux | amd64 | default | hp-linux-amd64 |
| linux | amd64 | http3 | hp-http3-linux-amd64 |

**Steps:**

1. Run tests first
2. Build binaries with `GOOS`/`GOARCH` env vars
3. Upload artifacts to GitHub Release using `softprops/action-gh-release`

```yaml
- name: Build
  run: |
    GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} go build -o hp-${{ matrix.goos }}-${{ matrix.goarch }} .

- name: Build HTTP/3
  run: |
    GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} go build -tags http3 -o hp-http3-${{ matrix.goos }}-${{ matrix.goarch }} .

- name: Upload Release Assets
  uses: softprops/action-gh-release@v2
  with:
    files: hp-*
```

---

## How Release Binaries Work

1. Create and push a tag: `git tag v0.8.0 && git push origin v0.8.0`
2. GitHub Actions detects the `v*` tag pattern
3. Workflow builds all platform/variant combinations
4. `softprops/action-gh-release` automatically:
   - Creates a GitHub Release for the tag
   - Uploads all binaries as release assets
5. Users download from: `https://github.com/colangelo/HittyPing/releases`

---

## Files to Create/Modify

| File | Action |
|------|--------|
| `main_test.go` | Create - unit tests |
| `.github/workflows/ci.yml` | Create - test on push/PR |
| `.github/workflows/release.yml` | Create - build & release on tags |

---

## Verification

1. Run tests locally: `go test -v ./...`
2. Push to dev branch → verify CI runs tests
3. Create test tag: `git tag v0.7.2-test && git push origin v0.7.2-test`
4. Check GitHub Actions → verify builds complete
5. Check Releases page → verify binaries attached
6. Delete test tag/release after verification
