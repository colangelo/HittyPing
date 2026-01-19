# Roadmap

Planned features and improvements for hp.

## Completed

### v0.6.1 - Display & Validation Improvements (DONE)

- [x] Show resolved IP in header: `dns.google [8.8.8.8]`
- [x] IPv6 address support
- [x] Early DNS validation (fail if hostname unresolvable)
- [x] Full name and version in header
- [x] `-q` for quiet mode (was `-n`)

### v0.6.0 - Protocol Shortcuts & Count (DONE)

- [x] `-1` shorthand for `--http` (HTTP/1.1)
- [x] `-2/--http2` flag to force HTTP/2 (fail if not negotiated)
- [x] `-3` shorthand for `--http3` (QUIC)
- [x] `-c/--count` flag to limit number of requests (like `ping -c`)

### v0.5.1 - CLI Polish (DONE)

- [x] `-v/--version` flag
- [x] Version output includes former name: `hp (hittyping) version X.Y.Z`

### v0.5.0 - HTTP/3 Support (DONE)

- [x] HTTP/3 (QUIC) support via `--http3` flag
- [x] Build tag `http3` to optionally include quic-go dependency
- [x] Default build remains small (~7.6MB), HTTP/3 build ~10MB
- [x] `just build-http3` and `just install-http3` recipes

### v0.4.0 - Rename & Protocol Options (DONE)

- [x] Rename binary: `hittyping` → `hp`
- [x] Rename env vars: `HITTYPING_*` → `HP_*`
- [x] Migrate to `spf13/pflag` for POSIX-style flags
- [x] `-k/--insecure` flag to skip TLS certificate verification
- [x] `--http` flag to use plain HTTP instead of HTTPS
- [x] Header shows protocol in use: `HP host (HTTPS)`

## Planned

## Ideas / Under Consideration

- [ ] `-j/--jitter` flag to add random variation to interval (anti-fingerprinting)
- [ ] DNS resolution timing breakdown (separate from HTTP RTT)
- [ ] TCP connection timing vs TLS handshake vs HTTP response
- [ ] JSON output mode for scripting
- [ ] Configuration file support (~/.config/hp.toml)
- [ ] Multiple targets in parallel

---

## HTTP/3 Implementation Notes

### The quic-go Dependency

HTTP/3 requires `github.com/quic-go/quic-go` - the reference Go implementation of QUIC.

**Pros:**
- 0-RTT connection establishment (faster first request)
- Better performance on lossy/mobile networks
- Built-in TLS 1.3
- Well-maintained (used by Caddy, Traefik, Cloudflare)
- Pure Go (no CGO required)

**Cons:**
- **Binary size**: +10-15MB (current ~8MB → ~20MB with HTTP/3)
- **Compile time**: Noticeably longer
- **Transitive dependencies**: Pulls in crypto, x509, and other packages
- **Server support**: Not all servers support HTTP/3 yet

### Proposed Implementation

Use Go build tags to make HTTP/3 optional:

```go
// http3.go
//go:build http3

package main

import "github.com/quic-go/quic-go/http3"
// HTTP/3 client implementation
```

```go
// http3_stub.go
//go:build !http3

package main

// Stub that returns error "HTTP/3 not compiled in, rebuild with -tags http3"
```

**Build commands:**
```bash
# Default build (no HTTP/3, small binary)
go build -o hp .

# With HTTP/3 support (larger binary)
go build -tags http3 -o hp .
```

**justfile recipes:**
```just
build:
    go build -o hp .

build-http3:
    go build -tags http3 -o hp .
```

This approach keeps the default binary small and dependency-free while allowing users who need HTTP/3 to opt-in.
