# Security Policy

## Supported Versions

Only the latest release is supported with security fixes.

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < latest | :x:               |

## Reporting a Vulnerability

**Please DO NOT open a public issue for security reports.**

Email: security@colangelo.dev

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

Each GitHub Release includes a `checksums.txt` file with SHA-256 hashes.

```bash
# Download the binary and checksums
curl -LO https://github.com/colangelo/HittyPing/releases/latest/download/hp-darwin-arm64
curl -LO https://github.com/colangelo/HittyPing/releases/latest/download/checksums.txt

# Verify
sha256sum -c checksums.txt --ignore-missing
# or on macOS:
shasum -a 256 -c checksums.txt --ignore-missing
```
