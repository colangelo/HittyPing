# Roadmap

Planned features and improvements for hp.

## Planned

### Multi-Target Mode

Monitor multiple hosts simultaneously with independent timing per target.

```sh
hp google.com cloudflare.com 1.1.1.1
```

**Design**: Stacked rows with labels above bars (see `docs/multi-target-spec.md`)

```txt
HP multi-target (3 hosts)
Legend: ▁▂▃<150ms ▄▅<400ms ▆▇█>=400ms !fail

google.com [142.250.180.14]
▁▁▂▁▄▃▁▂▁▁▂▃▁▁▂▁▁

cloudflare.com [104.16.132.229]
▁▁▁▁▁▁▂▁▁▁▁▁▁▁▁▁▁

1.1.1.1
▁▁▁▁▂▁▁▁▁▁▁▁▂▁▁▁▁
```

**Implementation approach**: Raw ANSI for initial version (keeps it simple). Could migrate to Bubble Tea later if interactive features are added.

---

## Ideas / Under Consideration

### Near-term Enhancements

- [ ] `-j/--jitter` flag to add random variation to interval (anti-fingerprinting)
- [ ] DNS resolution timing breakdown (separate from HTTP RTT)
- [ ] TCP connection timing vs TLS handshake vs HTTP response
- [ ] JSON output mode for scripting
- [ ] Configuration file support (~/.config/hp.toml)

### TUI Evolution (Bubble Tea)

If hp evolves toward interactive features, consider migrating to [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework:

**Benefits**:

- Clean Elm architecture (Model-Update-View)
- Built-in resize handling, input, mouse support
- Composable components via lipgloss styling
- Easier to maintain and test

**Potential features enabled by Bubble Tea**:

- [ ] Interactive mode (pause/resume, keyboard shortcuts)
- [ ] Target selection and drill-down
- [ ] Scrollable history buffer
- [ ] Live dashboard with stats panels
- [ ] Mouse support for target selection
- [ ] Split-pane views (targets + detail)

**Trade-off**: Adds ~2-4MB to binary size. Worth it only if interactive features are desired.

---

## Completed

### v0.7.4 - Windows Build Fix

- [x] Platform-specific terminal width detection (terminal_unix.go, terminal_windows.go)
- [x] Portable signal handling with os.Interrupt
- [x] Homebrew tap (`brew install colangelo/tap/hp`)
- [x] Scoop bucket for Windows
- [x] govulncheck in CI (local + GHA)
- [x] Automated Homebrew/Scoop manifest updates in release workflow
- [x] Improved downgrade message wording

### v0.7.3 - Release Improvements

- [x] MIT LICENSE file
- [x] Platform support note in README
- [x] Windows builds (amd64, arm64) in release workflow

### v0.7.2 - CI & Testing

- [x] GitHub Actions CI workflow (test on push/PR)
- [x] GitHub Actions release workflow (build on tag)
- [x] Unit tests for core functions
- [x] Linux arm64 builds in release matrix

### v0.7.1 - Downgrade Refinements

- [x] Downgrade only triggers at startup (before first successful ping)
- [x] Pre-tests lower protocols before committing (finds first working protocol)
- [x] Legend no longer reprints after downgrade (cleaner output)

### v0.7.0 - Protocol Downgrade

- [x] `-d/--downgrade` flag for auto-downgrade on 3 consecutive failures (secure only)
- [x] `-D/--downgrade-insecure` flag for full downgrade including plain HTTP
- [x] Fallback chain: HTTP/3 → HTTP/2 → HTTPS → HTTP (with -D)
- [x] Visual downgrade indicator message
- [x] Header reprints with new protocol after downgrade

### v0.6.1 - Display & Validation Improvements

- [x] Show resolved IP in header: `dns.google [8.8.8.8]`
- [x] IPv6 address support
- [x] Early DNS validation (fail if hostname unresolvable)
- [x] Full name and version in header
- [x] `-q` for quiet mode (was `-n`)

### v0.6.0 - Protocol Shortcuts & Count

- [x] `-1` shorthand for `--http` (HTTP/1.1)
- [x] `-2/--http2` flag to force HTTP/2 (fail if not negotiated)
- [x] `-3` shorthand for `--http3` (QUIC)
- [x] `-c/--count` flag to limit number of requests (like `ping -c`)

### v0.5.1 - CLI Polish

- [x] `-v/--version` flag
- [x] Version output includes former name: `hp (hittyping) version X.Y.Z`

### v0.5.0 - HTTP/3 Support

- [x] HTTP/3 (QUIC) support via `--http3` flag
- [x] HTTP/3 always included in standard build (~10MB)

### v0.4.0 - Rename & Protocol Options

- [x] Rename binary: `hittyping` → `hp`
- [x] Rename env vars: `HITTYPING_*` → `HP_*`
- [x] Migrate to `spf13/pflag` for POSIX-style flags
- [x] `-k/--insecure` flag to skip TLS certificate verification
- [x] `--http` flag to use plain HTTP instead of HTTPS
- [x] Header shows protocol in use: `HP host (HTTPS)`
