# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

hp (formerly hittyping) is a prettyping-style HTTP(S) latency monitor written in Go. It visualizes response times using Unicode block characters with color coding.

## Build Commands

```bash
just build         # Build binary (~10MB, includes HTTP/3)
just install       # Build and install to /usr/local/bin
just run           # Build and run with default target
just test          # Run tests
just fmt           # Format code
just lint          # Run golangci-lint
just vuln          # Run vulnerability check
just ci            # Local CI (lint, vuln, test)
just clean         # Remove binary
```

Or directly with Go:

```bash
go build -o hp .
```

Or with Docker:

```bash
docker run --rm ghcr.io/colangelo/hp cloudflare.com
```

## Usage

```bash
hp                                    # Default: https://1.1.1.1
hp dns.nextdns.io                     # Custom target (https:// auto-added)
hp -c 10 dns.nextdns.io               # Send 10 requests then exit
hp -i 500ms dns.nextdns.io            # 500ms interval (or --interval)
hp -t 3s cloudflare.com               # 3 second timeout (or --timeout)
hp -q dns.nextdns.io                  # Quiet mode (hide header and legend)
hp -Q dns.nextdns.io                  # Silent mode (pure bar output)
hp -k https://self-signed.example     # Skip TLS verification (or --insecure)
hp -1 example.com                     # Use plain HTTP/1.1 (or --http)
hp -2 cloudflare.com                  # Force HTTP/2 (or --http2)
hp -3 cloudflare.com                  # Use HTTP/3 (or --http3)
hp -3 -d example.com                  # HTTP/3 with auto-downgrade on failures
hp -3 -D example.com                  # Auto-downgrade including plain HTTP
hp -g 100 -y 200 8.8.8.8              # Custom thresholds (or --green, --yellow)
```

## Flags

| Short | Long | Env Var | Default | Description |
|-------|------|---------|---------|-------------|
| `-c` | `--count` | | 0 | Number of requests (0 = unlimited) |
| `-i` | `--interval` | | 1s | Request interval |
| `-t` | `--timeout` | | 5s | Request timeout |
| | `--legend` | | false | Show legend line |
| | `--noheader` | | false | Hide header line |
| `-q` | `--quiet` | | false | Hide header and legend |
| `-Q` | `--silent` | | false | Hide header, legend, and final stats |
| `-m` | `--min` | `HP_MIN` | 0 | Min latency baseline (ms) |
| `-g` | `--green` | `HP_GREEN` | 150 | Green threshold (ms) |
| `-y` | `--yellow` | `HP_YELLOW` | 400 | Yellow threshold (ms) |
| `-k` | `--insecure` | | false | Skip TLS verification |
| `-1` | `--http` | | false | Use plain HTTP/1.1 |
| `-2` | `--http2` | | false | Force HTTP/2 (fail if not negotiated) |
| `-3` | `--http3` | | false | Use HTTP/3 (QUIC) |
| `-d` | `--downgrade` | | false | Auto-downgrade on 3 failures (secure only) |
| `-D` | `--downgrade-insecure` | | false | Auto-downgrade including plain HTTP |
| `-v` | `--version` | | | Show version and exit |
| `-h` | `--help` | | | Show help and exit |

## Architecture

Go application using `spf13/pflag` for POSIX-style CLI flags. Uses latest stable Go (currently 1.25).

Key functions:

- `measureRTT()` - HEAD request timing
- `createClient()` - Creates HTTP client for given protocol level
- `getURLForProto()` - Returns URL with appropriate scheme for protocol
- `getBlock()` - Maps latency to Unicode block + color
- `printDisplay()` - Live bar and stats rendering with ANSI cursor control
- `printFinal()` - Summary on Ctrl+C

## Visual Output

```
HP dns.nextdns.io (HTTPS)
Legend: ▁▂▃<150ms ▄▅<400ms ▆▇█>=400ms !fail
▁▁▂▁▂▃▁▁
0/8 ( 0%) lost; 98/127/203ms; last: 102ms
```

- Green (▁▂▃): < green threshold
- Yellow (▄▅): < yellow threshold
- Red (▆▇█): >= yellow threshold
- Red bold (!): Failed request

## Git Workflow

**Branch protection is enabled on `main`.** Direct pushes are blocked.

**IMPORTANT: Never delete the `dev` branch.** It is the main development branch.

### Making changes to main

1. Create a feature branch: `git checkout -b fix/description`
2. Make changes and commit
3. Push branch: `git push origin fix/description`
4. Create PR: `gh pr create --base main`
5. Wait for CI (lint, test, CodeQL) to pass
6. Merge PR: `gh pr merge --merge --delete-branch` (OK to delete feature branches)

### Releasing a new version

1. Update `const version` in `main.go` (use `just bump [major|minor|patch]`)
2. Update `CHANGELOG.md` and `ROADMAP.md`
3. Merge dev to main via PR: `gh pr create --base main --head dev`
4. Merge PR: `gh pr merge --merge` (**DO NOT use --delete-branch for dev**)
5. Checkout main, pull, create and push tag: `git tag -a vX.Y.Z -m "message" && git push origin vX.Y.Z`
6. Tag push triggers release workflow (builds, signs with cosign, updates Homebrew/Scoop)
7. Optionally set custom title: `gh release edit vX.Y.Z --title "vX.Y.Z - Title"`

### CI Requirements

PRs to main require:
- `lint` - golangci-lint
- `test` - go test
- `CodeQL` - security scanning
