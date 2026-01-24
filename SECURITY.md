# Security Policy

## Supported Versions

Only the latest release is supported with security fixes.

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < latest | :x:               |

## Reporting a Vulnerability

**Please DO NOT open a public issue for security reports.**

Use [GitHub Private Vulnerability Reporting](../../security/advisories/new) or email: security@colangelo.dev

Include:
- Version or commit hash
- OS and architecture
- Steps to reproduce
- Expected vs actual behavior

You can expect an initial response within 48 hours. If confirmed, a fix will be prioritized and credited in the release notes (unless you prefer to remain anonymous).

## What This Project Does / Does Not Do

**Does:**
- Perform outbound HTTP/HTTPS/HTTP3 requests to user-specified URLs
- Display latency statistics in the terminal

**Does NOT:**
- Execute shell commands
- Collect telemetry or analytics
- Phone home or make requests to any URL other than the user-specified target
- Store any data persistently
- Require elevated privileges

## Verifying Releases

All release binaries are signed with [Sigstore cosign](https://docs.sigstore.dev/) using GitHub Actions OIDC.

Each release includes:
- `hp-<os>-<arch>` - the binary
- `hp-<os>-<arch>.sig` - cosign signature
- `hp-<os>-<arch>.pem` - signing certificate
- `checksums.txt` - SHA256 checksums
- `checksums.txt.sig` / `checksums.txt.pem` - signed checksums

### Verify with cosign (recommended)

```bash
# Install cosign: https://docs.sigstore.dev/cosign/system_config/installation/

cosign verify-blob \
  --signature hp-linux-amd64.sig \
  --certificate hp-linux-amd64.pem \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp 'github.com/colangelo/HittyPing' \
  hp-linux-amd64
```

Or download and verify automatically for your platform:

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]') && \
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/') && \
BIN="hp-${OS}-${ARCH}" && \
BASE_URL="https://github.com/colangelo/HittyPing/releases/latest/download" && \
curl -sLO "${BASE_URL}/${BIN}" && \
curl -sLO "${BASE_URL}/${BIN}.sig" && \
curl -sLO "${BASE_URL}/${BIN}.pem" && \
cosign verify-blob \
  --signature "${BIN}.sig" \
  --certificate "${BIN}.pem" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp 'github.com/colangelo/HittyPing' \
  "${BIN}"
```

### Verify checksums only

```bash
# Download the binary and checksums
curl -LO https://github.com/colangelo/HittyPing/releases/latest/download/hp-darwin-arm64
curl -LO https://github.com/colangelo/HittyPing/releases/latest/download/checksums.txt

# Verify
sha256sum -c checksums.txt --ignore-missing
# or on macOS:
shasum -a 256 -c checksums.txt --ignore-missing
```

### Why verify?

Cosign verification ensures:
1. The binary was built by GitHub Actions (not a compromised maintainer)
2. The binary hasn't been tampered with since release
3. You're running exactly what was built from the source code
