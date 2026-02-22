# Changelog

All notable changes to hp (hittyping) will be documented in this file.

## [0.8.2] - 2026-02-22

### Fixed

- Cursor no longer lands on the stats line in braille mode (replaced absolute save/restore with relative cursor movement)
- Stats line truncated to terminal width to prevent wrapping-induced scroll

### Added

- Beta release channel via Homebrew tap (`brew install colangelo/tap/hp-beta`)
- `just beta` recipe to trigger beta builds from dev branch

## [0.8.1] - 2026-02-12

### Fixed

- Keypresses (Ctrl-O, Ctrl-R, etc.) no longer corrupt the bar display
  - Disabled terminal echo, canonical mode, and extended input processing (VDISCARD/VREPRINT)
  - Ctrl+C still works via ISIG
- Ctrl-Z (suspend) now properly restores terminal state and cursor
  - Bar and stats redraw correctly on `fg` resume
  - Display mutex prevents output race during suspend sequence
  - Only current line blocks are redrawn (no spurious extra line)

### Added

- Steady block cursor via DECSCUSR on terminals that support it (Kitty, Alacritty, WezTerm, GNOME Terminal, Windows Terminal, xterm)

## [0.8.0] - 2026-01-25

### Added

- `-b/--braille` flag for braille character visualization (2x density)
  - Packs 2 readings per character using braille dot patterns
  - Shows more history in the same terminal width

## [0.7.9] - 2026-01-25

### Added

- `-j/--jitter` flag to add random variation to request interval (anti-fingerprinting)

### Security

- Pin Docker base images by SHA256 hash (supply chain security)
- Move workflow permissions to job level (least privilege)

## [0.7.8] - 2026-01-25

### Added

- Docker container image published to GitHub Container Registry (ghcr.io/colangelo/hp)
- Release procedure documentation (docs/release-procedure.md)

## [0.7.7] - 2026-01-25

### Added

- `--noheader` flag to hide header line
- `-q/--quiet` flag to hide header and legend
- `-Q/--silent` flag to hide header, legend, and final stats (pure bar output)

### Changed

- **Breaking**: `-q` now means `--quiet` (hide header + legend), not `--nolegend`
- **Breaking**: Legend hidden by default; use `--legend` to show it (replaces `--nolegend`)

## [0.7.6] - 2026-01-24

### Added

- Cosign-signed release binaries for supply-chain verification
- Release verification documentation in SECURITY.md

### Changed

- Pin all GitHub Actions to commit SHAs (security hardening)
- Add restrictive `permissions: read-all` to CI workflows
- Consolidate security docs into reusable template (`docs/github-security-setup.md`)

### Removed

- `docs/security-checklist.md` (merged into `docs/github-security-setup.md`)

## [0.7.5] - 2026-01-24

### Added

- Uninstall recipe in justfile (`just uninstall`)
- OpenSSF Scorecard workflow and README badge
- CodeQL security scanning workflow
- SECURITY.md with vulnerability reporting policy
- SHA-256 checksums.txt in GitHub releases
- Branch protection ruleset for main branch
- Security checklist documentation (`docs/security-checklist.md`)

### Fixed

- Silence quic-go UDP buffer warnings that corrupted terminal display on Linux
- Remove outdated build tag reference from `--http3` help text

### Changed

- Header styling: hostname now bold, IP address highlighted for better visibility

### Security

- Enable GitHub Advanced Security features (Dependabot, secret scanning, push protection)
- Require status checks and CodeQL scanning for merges to main

## [0.7.4] - 2026-01-19

### Added

- Homebrew tap: `brew install colangelo/tap/hp`
- Scoop bucket for Windows installation
- govulncheck in CI (local justfile + GitHub Actions)
- Automated Homebrew/Scoop manifest updates in release workflow

### Fixed

- Windows build: split terminal width detection into platform-specific files
- Use portable `os.Interrupt` for signal handling

### Changed

- Improved downgrade message wording

## [0.7.3] - 2026-01-19

### Added

- MIT LICENSE file
- Platform support note in README (developed/tested on macOS)
- Windows builds (amd64, arm64) in release workflow

## [0.7.2] - 2026-01-19

### Added

- GitHub Actions CI workflow (test on push/PR)
- GitHub Actions release workflow (build on tag)
- Unit tests for `getEnvInt`, `getURLForProto`, `getBlock`
- Linux arm64 builds in release matrix

## [0.7.1] - 2026-01-19

### Changed

- Downgrade now only triggers at startup (before first successful ping)
- Pre-tests lower protocols before committing to downgrade (finds first working protocol)
- Legend no longer reprints after downgrade (cleaner output)

## [0.7.0] - 2026-01-19

### Added

- `-d/--downgrade` flag for auto-downgrade on 3 consecutive failures (secure only: HTTP/3 → HTTP/2 → HTTPS)
- `-D/--downgrade-insecure` flag for full downgrade including plain HTTP
- Protocol level tracking with dynamic client recreation on downgrade
- Visual downgrade indicator: `↓ Downgrading to HTTP/2 after 3 failures`

### Changed

- Header reprints with new protocol after downgrade
- Refactored client creation into `createClient()` helper function

## [0.6.1] - 2025-01-19

### Added

- Show resolved IP address in header: `dns.google [8.8.8.8]`
- IPv6 address support (wrapped in brackets for valid URLs)
- Early DNS validation - fail immediately if hostname cannot be resolved

### Changed

- Header now shows full name and version: `HittyPing (v0.6.1) dns.google [8.8.8.8] (HTTPS)`
- `-n` changed to `-q` for `--nolegend` (quiet mode)

## [0.6.0] - 2025-01-18

### Added

- `-1` shorthand for `--http` (HTTP/1.1)
- `-2/--http2` flag to force HTTP/2 (fails if not negotiated)
- `-3` shorthand for `--http3` (QUIC)
- `-c/--count` flag to limit number of requests (like `ping -c`)

### Changed

- Protocol flags are now mutually exclusive
- `--http` display changed from "HTTP" to "HTTP/1.1"

## [0.5.1] - 2025-01-18

### Added

- `-v/--version` flag to show version and exit
- `-h/--help` flag (provided by pflag)

### Changed

- Version output now shows former name: `hp (hittyping) version X.Y.Z`

## [0.5.0] - 2025-01-18

### Added

- HTTP/3 (QUIC) support via `--http3` flag (always included in build, ~10MB)

### Changed

- Module renamed to `github.com/ac/hp`

### Notes

- HTTP/3 requires servers that support it (e.g., Cloudflare, Google)
- First HTTP/3 request may be slower due to QUIC handshake

## [0.4.0] - 2025-01-18

### Breaking Changes

- **Renamed binary**: `hittyping` → `hp`
- **Renamed env vars**: `HITTYPING_*` → `HP_*`

### Added

- POSIX-style CLI flags via `spf13/pflag`:
  - `-i/--interval`, `-t/--timeout`, `-n/--nolegend`
  - `-m/--min`, `-g/--green`, `-y/--yellow`
- `-k/--insecure` flag to skip TLS certificate verification
- `--http` flag to use plain HTTP instead of HTTPS
- Header now shows protocol: `HP host (HTTPS)` or `HP host (HTTP)`

### Changed

- All flags now support both short (`-i`) and long (`--interval`) forms

## [0.3.1] - 2025-01-18

### Changed

- Failure symbol changed from gray `×` to red bold `!` (matches prettyping)
- Fixed last block not appearing before line wrap

## [0.3.0] - 2025-01-18

### Added

- Configurable color thresholds via flags (`-green`, `-yellow`) and env vars (`HITTYPING_GREEN`, `HITTYPING_YELLOW`)
- Minimum latency baseline (`-min` / `HITTYPING_MIN`) for smallest block scaling
- Terminal width detection - bar now wraps to next line instead of truncating
- Cursor follows the bar as it grows

### Changed

- Bar height now correlates with color zones:
  - Green zone (< green threshold): ▁▂▃
  - Yellow zone (green to yellow): ▄▅
  - Red zone (>= yellow): ▆▇█
- Legend displays actual threshold values
- Previous bar lines preserved when wrapping to new line

## [0.2.0] - 2025-01-18

### Added

- `--nolegend` flag to hide the legend line for cleaner output
- Ad-hoc code signing on macOS install to prevent Gatekeeper kills
- justfile with `build`, `install`, `run`, and `clean` recipes
- CLAUDE.md for repository guidance

### Changed

- Header now displays hostname without `https://` prefix
- Final statistics also show clean hostname

## [0.1.0] - 2025-01-18

### Added

- Initial release
- HTTPS HEAD request latency monitoring
- prettyping-style Unicode block visualization (▁▂▃▄▅▆▇█)
- Color-coded output: green (<150ms), yellow (<400ms), red (>400ms)
- Live statistics: min/avg/max latency, packet loss percentage
- Graceful Ctrl+C handling with final summary
- Configurable interval (`-i`) and timeout (`-t`) flags
- Auto-prepends `https://` to bare hostnames
