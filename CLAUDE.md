# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

hittyping is a prettyping-style HTTPS latency monitor written in Go. It visualizes response times using Unicode block characters with color coding.

## Build Commands

```bash
just build    # Build binary
just install  # Build and install to /usr/local/bin
just run      # Build and run with default target
just clean    # Remove binary
```

Or directly with Go:

```bash
go build -o hittyping .
```

## Usage

```bash
hittyping                              # Default: https://1.1.1.1
hittyping dns.nextdns.io               # Custom target (https:// auto-added)
hittyping -i 500ms dns.nextdns.io      # 500ms interval
hittyping -t 3s cloudflare.com         # 3 second timeout
hittyping --nolegend dns.nextdns.io    # Hide legend line
```

## Architecture

Single-file Go application (`main.go`) with no external dependencies:

- `measureRTT()` - HEAD request timing
- `getBlock()` - Maps latency to Unicode block + color
- `printDisplay()` - Live bar and stats rendering with ANSI cursor control
- `printFinal()` - Summary on Ctrl+C

## Visual Output

```
HITTYPING dns.nextdns.io
Legend: ▁▂▃<150ms ▄▅<400ms ▆▇█>400ms ×fail
▁▁▂▁▂▃▁▁
0/8 ( 0%) lost; 98/127/203ms; last: 102ms
```

- Green (▁▂▃): <150ms
- Yellow (▄▅): <400ms
- Red (▆▇█): >400ms
- Gray (×): Failed request
