# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

hp (formerly hittyping) is a prettyping-style HTTP(S) latency monitor written in Go. It visualizes response times using Unicode block characters with color coding.

## Build Commands

```bash
just build         # Build binary (~7.6MB)
just build-http3   # Build with HTTP/3 support (~10MB)
just install       # Build and install to /usr/local/bin
just install-http3 # Build with HTTP/3 and install
just run           # Build and run with default target
just clean         # Remove binary
```

Or directly with Go:

```bash
go build -o hp .                # Default build
go build -tags http3 -o hp .    # With HTTP/3 support
```

## Beads workflow

Check ./BD_WORKFLOW.md

## MCP Agent Mail: coordination for multi-agent workflows

What it is
- A mail-like layer that lets coding agents coordinate asynchronously via MCP tools and resources.
- Provides identities, inbox/outbox, searchable threads, and advisory file reservations, with human-auditable artifacts in Git.

Why it's useful
- Prevents agents from stepping on each other with explicit file reservations (leases) for files/globs.
- Keeps communication out of your token budget by storing messages in a per-project archive.
- Offers quick reads (`resource://inbox/...`, `resource://thread/...`) and macros that bundle common flows.

How to use effectively
1) Same repository
   - Register an identity: call `ensure_project`, then `register_agent` using this repo's absolute path as `project_key`.
   - Reserve files before you edit: `file_reservation_paths(project_key, agent_name, ["src/**"], ttl_seconds=3600, exclusive=true)` to signal intent and avoid conflict.
   - Communicate with threads: use `send_message(..., thread_id="FEAT-123")`; check inbox with `fetch_inbox` and acknowledge with `acknowledge_message`.
   - Read fast: `resource://inbox/{Agent}?project=<abs-path>&limit=20` or `resource://thread/{id}?project=<abs-path>&include_bodies=true`.
   - Tip: set `AGENT_NAME` in your environment so the pre-commit guard can block commits that conflict with others' active exclusive file reservations.

2) Across different repos in one project (e.g., Next.js frontend + FastAPI backend)
   - Option A (single project bus): register both sides under the same `project_key` (shared key/path). Keep reservation patterns specific (e.g., `frontend/**` vs `backend/**`).
   - Option B (separate projects): each repo has its own `project_key`; use `macro_contact_handshake` or `request_contact`/`respond_contact` to link agents, then message directly. Keep a shared `thread_id` (e.g., ticket key) across repos for clean summaries/audits.

Macros vs granular tools
- Prefer macros when you want speed or are on a smaller model: `macro_start_session`, `macro_prepare_thread`, `macro_file_reservation_cycle`, `macro_contact_handshake`.
- Use granular tools when you need control: `register_agent`, `file_reservation_paths`, `send_message`, `fetch_inbox`, `acknowledge_message`.

Common pitfalls
- "from_agent not registered": always `register_agent` in the correct `project_key` first.
- "FILE_RESERVATION_CONFLICT": adjust patterns, wait for expiry, or use a non-exclusive reservation when appropriate.
- Auth errors: if JWT+JWKS is enabled, include a bearer token with a `kid` that matches server JWKS; static bearer is used only when JWT is disabled.

## Usage

```bash
hp                                    # Default: https://1.1.1.1
hp dns.nextdns.io                     # Custom target (https:// auto-added)
hp -c 10 dns.nextdns.io               # Send 10 requests then exit
hp -i 500ms dns.nextdns.io            # 500ms interval (or --interval)
hp -t 3s cloudflare.com               # 3 second timeout (or --timeout)
hp -q dns.nextdns.io                  # Quiet mode (hide legend)
hp -k https://self-signed.example     # Skip TLS verification (or --insecure)
hp -1 example.com                     # Use plain HTTP/1.1 (or --http)
hp -2 cloudflare.com                  # Force HTTP/2 (or --http2)
hp -3 cloudflare.com                  # Use HTTP/3 (or --http3) - requires http3 build
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
| `-q` | `--nolegend` | | false | Quiet mode (hide legend) |
| `-m` | `--min` | `HP_MIN` | 0 | Min latency baseline (ms) |
| `-g` | `--green` | `HP_GREEN` | 150 | Green threshold (ms) |
| `-y` | `--yellow` | `HP_YELLOW` | 400 | Yellow threshold (ms) |
| `-k` | `--insecure` | | false | Skip TLS verification |
| `-1` | `--http` | | false | Use plain HTTP/1.1 |
| `-2` | `--http2` | | false | Force HTTP/2 (fail if not negotiated) |
| `-3` | `--http3` | | false | Use HTTP/3 (requires http3 build tag) |
| `-d` | `--downgrade` | | false | Auto-downgrade on 3 failures (secure only) |
| `-D` | `--downgrade-insecure` | | false | Auto-downgrade including plain HTTP |
| `-v` | `--version` | | | Show version and exit |
| `-h` | `--help` | | | Show help and exit |

## Architecture

Go application using `spf13/pflag` for POSIX-style CLI flags.

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
