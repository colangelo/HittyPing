# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

hp (formerly hittyping) is a prettyping-style HTTP(S) latency monitor written in Go. It visualizes response times using Unicode block characters with color coding.

## Build Commands

```bash
just build    # Build binary
just install  # Build and install to /usr/local/bin
just run      # Build and run with default target
just clean    # Remove binary
```

Or directly with Go:

```bash
go build -o hp .
```

## Usage

```bash
hp                                    # Default: https://1.1.1.1
hp dns.nextdns.io                     # Custom target (https:// auto-added)
hp -i 500ms dns.nextdns.io            # 500ms interval (or --interval)
hp -t 3s cloudflare.com               # 3 second timeout (or --timeout)
hp -n dns.nextdns.io                  # Hide legend (or --nolegend)
hp -k https://self-signed.example     # Skip TLS verification (or --insecure)
hp --http example.com                 # Use plain HTTP instead of HTTPS
hp -g 100 -y 200 8.8.8.8              # Custom thresholds (or --green, --yellow)
```

## Flags

| Short | Long | Env Var | Default | Description |
|-------|------|---------|---------|-------------|
| `-i` | `--interval` | | 1s | Request interval |
| `-t` | `--timeout` | | 5s | Request timeout |
| `-n` | `--nolegend` | | false | Hide legend line |
| `-m` | `--min` | `HP_MIN` | 0 | Min latency baseline (ms) |
| `-g` | `--green` | `HP_GREEN` | 150 | Green threshold (ms) |
| `-y` | `--yellow` | `HP_YELLOW` | 400 | Yellow threshold (ms) |
| `-k` | `--insecure` | | false | Skip TLS verification |
| | `--http` | | false | Use plain HTTP |

## Architecture

Go application using `spf13/pflag` for POSIX-style CLI flags.

Key functions:
- `measureRTT()` - HEAD request timing
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
