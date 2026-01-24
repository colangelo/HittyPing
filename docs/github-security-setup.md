# GitHub Security Setup for CLI Projects

Reusable checklist for securing Go CLI projects on GitHub. Optimized for solo/small projects.

## Quick Reference

| Category | Items |
|----------|-------|
| Workflows | OpenSSF Scorecard, CodeQL, govulncheck |
| Release signing | cosign + checksums |
| GitHub settings | Dependabot, secret scanning, push protection |
| Branch protection | Block force push, require status checks |
| Documentation | SECURITY.md |

---

## 1. GitHub Actions Workflows

### OpenSSF Scorecard

Provides a public security score and badge.

`.github/workflows/scorecard.yml`:

```yaml
name: OpenSSF Scorecard

on:
  branch_protection_rule:
  schedule:
    - cron: '30 1 * * 6'  # weekly
  push:
    branches: ["main"]

permissions: read-all

jobs:
  analysis:
    name: Scorecard analysis
    runs-on: ubuntu-latest
    permissions:
      security-events: write
      id-token: write
    steps:
      - uses: actions/checkout@<SHA> # v4
        with:
          persist-credentials: false

      - uses: ossf/scorecard-action@<SHA> # v2.4.0
        with:
          results_file: results.sarif
          results_format: sarif
          publish_results: true

      - uses: github/codeql-action/upload-sarif@<SHA> # v3
        with:
          sarif_file: results.sarif
```

README badge:

```markdown
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/OWNER/REPO/badge)](https://scorecard.dev/viewer/?uri=github.com/OWNER/REPO)
```

### CodeQL

Finds code-level security bugs. Results appear in GitHub Security tab.

`.github/workflows/codeql.yml`:

```yaml
name: CodeQL

on:
  push:
    branches: ["main"]
  pull_request:
    branches: ["main"]
  schedule:
    - cron: '0 3 * * 1'  # weekly

permissions:
  security-events: write
  contents: read

jobs:
  analyze:
    name: Analyze (CodeQL)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@<SHA> # v4

      - uses: github/codeql-action/init@<SHA> # v3
        with:
          languages: go
          queries: security-extended

      - uses: github/codeql-action/autobuild@<SHA> # v3

      - uses: github/codeql-action/analyze@<SHA> # v3
```

### govulncheck

Go's official vulnerability checker. Lower noise than dependency CVE scanners.

`.github/workflows/govulncheck.yml` (or add to CI):

```yaml
name: govulncheck

on:
  push:
    branches: ["main"]
  pull_request:

permissions:
  contents: read

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@<SHA> # v4
      - uses: actions/setup-go@<SHA> # v5
        with:
          go-version-file: go.mod
      - run: go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

---

## 2. Release Signing with Cosign

Sign binaries with Sigstore cosign using GitHub Actions OIDC (keyless).

Add to your release workflow:

```yaml
permissions:
  contents: write
  id-token: write  # Required for cosign OIDC

jobs:
  release:
    steps:
      # ... build steps ...

      - name: Generate checksums
        run: sha256sum my-binary-* > checksums.txt

      - name: Install cosign
        uses: sigstore/cosign-installer@<SHA> # v3

      - name: Sign binaries with cosign
        run: |
          for file in my-binary-*; do
            cosign sign-blob --yes \
              --output-signature "${file}.sig" \
              --output-certificate "${file}.pem" \
              "${file}"
          done
          cosign sign-blob --yes \
            --output-signature checksums.txt.sig \
            --output-certificate checksums.txt.pem \
            checksums.txt

      - name: Upload Release Assets
        uses: softprops/action-gh-release@<SHA> # v2
        with:
          files: |
            my-binary-*
            checksums.txt*
```

Users verify with:

```bash
cosign verify-blob \
  --signature my-binary.sig \
  --certificate my-binary.pem \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp 'github.com/OWNER/REPO' \
  my-binary
```

---

## 3. GitHub Repository Settings

### Security Tab (Settings > Code security and analysis)

**Essential (enable these):**

- [x] Private vulnerability reporting
- [x] Dependency graph
- [x] Dependabot alerts
- [x] Dependabot security updates
- [x] Secret scanning
- [x] Push protection

**Recommended:**

- [x] Grouped security updates (reduces PR noise)
- [x] Copilot Autofix (AI-suggested fixes for CodeQL alerts)

**Optional (can be noisy):**

- [ ] Dependabot version updates (auto-PRs for non-security bumps)
- [ ] Automatic dependency submission

---

## 4. Branch Protection (Settings > Rules > Rulesets)

Create a ruleset for `main`:

**Recommended for solo/small projects:**

- [x] Restrict deletions
- [x] Block force pushes
- [x] Require status checks to pass (`lint`, `test`)

**Additional for stricter security:**

- [x] Require code scanning results (CodeQL)
- [ ] Require a pull request before merging (friction for solo dev)
- [ ] Require approvals (no one else to approve)

**Usually overkill:**

- [ ] Require signed commits
- [ ] Require linear history
- [ ] Restrict creations / Restrict updates

---

## 5. OpenSSF Scorecard Optimization

### High-impact fixes (easy wins)

| Check | Score Impact | Fix |
|-------|--------------|-----|
| Pinned-Dependencies | 0 → 10 | Pin actions to SHA: `uses: actions/checkout@<sha> # v4` |
| Token-Permissions | 0 → 10 | Add `permissions: read-all` at workflow top level |
| Signed-Releases | 0 → 10 | Add cosign signing (see section 2) |
| Branch-Protection | 3 → 8+ | Enable status checks, block force push |

### Lower priority (limited ROI for solo projects)

| Check | Why it's hard |
|-------|---------------|
| CII-Best-Practices | Requires manual badge registration |
| Code-Review | Requires PR workflow (painful for solo) |
| Contributors | Can't fix - solo project |
| Maintained | Time-based - new repos score low |
| Fuzzing | Overkill for simple CLI tools |

### Getting action SHAs

```bash
# Get current SHA for an action tag
gh api repos/actions/checkout/commits/v4 --jq '.sha'
gh api repos/actions/setup-go/commits/v5 --jq '.sha'
gh api repos/sigstore/cosign-installer/commits/v3 --jq '.sha'
```

---

## 6. Documentation Files

### SECURITY.md

Required for GitHub Security tab integration:

```markdown
# Security Policy

## Supported Versions

Only the latest release is supported with security fixes.

## Reporting a Vulnerability

Use [GitHub Private Vulnerability Reporting](../../security/advisories/new)

Include: version, OS, steps to reproduce, expected vs actual behavior.

## Verifying Releases

[Include cosign verification instructions]
```

### checksums.txt

Auto-generated in release workflow. Include SHA256 hashes of all binaries.

---

## Recommended Setup Summary

### Workflows to add

1. OpenSSF Scorecard + README badge
2. CodeQL
3. govulncheck (standalone or in CI)
4. Release signing with cosign

### GitHub settings to enable

- Private vulnerability reporting
- Dependency graph + Dependabot alerts/security updates
- Grouped security updates
- Secret scanning + Push protection

### Branch protection

- Block force pushes + deletions
- Require status checks (lint, test)
- Require code scanning (CodeQL)

### Files to add

- `SECURITY.md` with disclosure policy + verification instructions
- `checksums.txt` in releases (auto-generated)
- `*.sig` / `*.pem` signatures in releases (auto-generated)

---

## Expected Scorecard Impact

| Before | After | Notes |
|--------|-------|-------|
| ~4-5/10 | ~7-8/10 | Without CII badge or PR reviews |
| ~7-8/10 | ~9/10 | With CII badge registration |

The remaining gaps are typically: CII-Best-Practices (manual), Contributors (solo), and Maintained (time-based).
