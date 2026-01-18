# Changelog

All notable changes to hp (formerly hittyping) will be documented in this file.

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
