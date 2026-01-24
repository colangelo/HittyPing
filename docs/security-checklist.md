# Security Posture

For a CLI utility (especially one people curl | sh / brew install / go install), the best “trust signals” are automated, verifiable security checks + a clear security policy.

Your repo already has CI + release badges in the README (nice). Here are the most useful additions to make it feel meaningfully “secure to use”, with copy-paste snippets.

⸻

## 1) Add an OpenSSF Scorecard badge (quick "security posture" signal)

OpenSSF Scorecard is widely recognized and gives you a public score + badge.  ￼

README badge (copy/paste):

```markdown
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/colangelo/HittyPing/badge)](https://scorecard.dev/viewer/?uri=github.com/colangelo/HittyPing)
```

Workflow (.github/workflows/scorecard.yml):

```yaml
name: OpenSSF Scorecard
on:
  branch_protection_rule:
  schedule:
    - cron: '30 1 * * 6'  # weekly
  push:
    branches: [ "main" ]

permissions: read-all

jobs:
  analysis:
    name: Scorecard analysis
    runs-on: ubuntu-latest
    permissions:
      security-events: write
      id-token: write
    steps:
      - uses: actions/checkout@v4
      - uses: ossf/scorecard-action@v2.4.0
        with:
          results_file: results.sarif
          results_format: sarif
          publish_results: true
      - uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif
```

(That’s directly based on the official Scorecard Action docs.)  ￼

⸻

## 2) Enable CodeQL scanning (find real code-level security bugs)

This adds the “Code scanning alerts” view under the GitHub Security tab and catches a lot of common issues.  ￼

Workflow (.github/workflows/codeql.yml):

```yaml
name: CodeQL
on:
  push:
    branches: [ "main" ]
  pull_request:
    branches: [ "main" ]
  schedule:
    - cron: '0 3 ** 1'  # weekly

permissions:
  security-events: write
  contents: read

jobs:
  analyze:
    name: Analyze (CodeQL)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: github/codeql-action/init@v3
        with:
          languages: go
          queries: security-extended
      - uses: github/codeql-action/autobuild@v3
      - uses: github/codeql-action/analyze@v3
```

⸻

## 3) Add govulncheck (Go's official vuln checker)

This is a big credibility boost for Go projects: it flags vulnerabilities that actually affect your reachable code paths (lower-noise than just dependency CVEs).  ￼

Workflow (.github/workflows/govulncheck.yml):

```yaml
name: govulncheck
on:
  push:
    branches: [ "main" ]
  pull_request:
    branches: [ "main" ]
  schedule:
    - cron: '0 2 ** 3'  # weekly

permissions:
  contents: read

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - uses: golang/govulncheck-action@v1
        with:
          go-version-file: go.mod
          repo-checkout: false
```

If you want results to show up in GitHub “Code scanning alerts”, there are SARIF-upload approaches too, but even the basic workflow is already valuable.  ￼

⸻

## 4) Add a SECURITY.md (sets expectations + disclosure flow)

This is small but important for user trust.

SECURITY.md template:

# Security Policy

## Supported Versions

Only the latest release is supported with security fixes.

## Reporting a Vulnerability

Please DO NOT open a public issue for security reports.

Email: <YOUR_EMAIL>
Include:

- version / commit hash
- OS + architecture
- steps to reproduce
- expected vs actual behavior

## What this project does / does not do

- Does not execute shell commands.
- Does not collect telemetry.
- Only performs outbound network requests to the target URL provided by the user.

⸻

## 5) Supply-chain verification for releases (best "this binary is legit" proof)

For a CLI tool, the strongest reassurance is: signed artifacts + checksums + provenance.

Minimum good practice:
 • publish checksums.txt in every GitHub Release
 • sign it (GPG or cosign)

Even better:
 • generate SLSA provenance and sign it with OIDC/cosign (verifiable)  ￼

If you ever want to go this route, you can add a “Verification” section to README:

## Verifying releases

1) Download `hp_<os>_<arch>.tar.gz` and `checksums.txt`
2) Verify checksum:
   sha256sum -c checksums.txt

(Then later you can upgrade it with cosign provenance.)

⸻

## 6) Turn on GitHub native security features (no code changes)

In repo Settings → Security → Advanced Security:

**Essential (enable these):**
 • ✅ Private vulnerability reporting
 • ✅ Dependency graph
 • ✅ Dependabot alerts
 • ✅ Dependabot security updates
 • ✅ Secret Protection
 • ✅ Push protection

**Recommended:**
 • ✅ Grouped security updates (reduces PR noise)
 • ✅ Copilot Autofix (AI-suggested fixes for CodeQL alerts)

**Optional:**
 • Dependabot version updates (auto-PRs for non-security bumps, can be noisy)
 • Automatic dependency submission (for build-time dependencies)

**Code scanning section:**
 • CodeQL analysis: Use "Advanced" setup if you have a workflow file, or "Default" for GitHub-managed
 • Protection rules: Configure severity thresholds for blocking merges

These show up in the repo's "Security" tab and are familiar to users.

⸻

## 7) Protect your main branch with rulesets

In repo Settings → Rules → Rulesets → New branch ruleset:

**Basic setup:**
 • Ruleset name: `main-protection`
 • Enforcement status: **Active**
 • Target branches: **Default** (applies to `main`)

**Recommended rules for solo/small projects:**
 • ✅ Restrict deletions
 • ✅ Block force pushes
 • ✅ Require status checks to pass → add `lint`, `test`

**Additional for teams or stricter security:**
 • ✅ Require code scanning results → CodeQL
 • ✅ Require a pull request before merging
 • ✅ Require approvals

**Leave unchecked** (usually overkill):
 • Restrict creations / Restrict updates
 • Require linear history
 • Require signed commits
 • Require deployments to succeed

Branch protection improves your OpenSSF Scorecard and prevents accidental force-pushes or deletions.

⸻

## Recommended setup for solo/small CLI projects

**Workflows to add:**
 • OpenSSF Scorecard Action + README badge
 • CodeQL workflow
 • govulncheck (can be in CI or standalone workflow)

**GitHub settings to enable:**
 • Private vulnerability reporting
 • Dependency graph
 • Dependabot alerts + security updates
 • Grouped security updates
 • Secret Protection + Push protection
 • Copilot Autofix

**Branch protection ruleset:**
 • Restrict deletions
 • Block force pushes
 • Require status checks (`lint`, `test`)
 • Require code scanning (CodeQL)

**Files to add:**
 • `SECURITY.md` with disclosure policy
 • `checksums.txt` in releases

**What to skip for solo projects:**

These add friction without much benefit when you're the only contributor:

 • *Require a pull request before merging* — forces you to create PRs for every change instead of pushing directly to main
 • *Require approvals* — no one else to approve anyway
 • *Require signed commits* — setup overhead; useful for high-security projects
 • *Require linear history* — prevents merge commits; matter of preference
 • *Dependabot version updates* — creates noisy PRs for every dependency bump, not just security fixes
 • *Restrict creations / Restrict updates* — too restrictive for normal development

This setup signals "continuously scanned and maintained" while staying practical for solo development.
