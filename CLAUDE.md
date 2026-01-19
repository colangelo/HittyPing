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
just vet           # Run go vet
just ci            # Local CI (fmt check, vet, test)
just clean         # Remove binary
```

Or directly with Go:

```bash
go build -o hp .
```

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
| `-q` | `--nolegend` | | false | Quiet mode (hide legend) |
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
