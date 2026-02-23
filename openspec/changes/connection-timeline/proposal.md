## Why

When diagnosing intermittent connectivity issues (VPN drops, flaky proxies, unstable links), knowing *when* outages happen and how long the connection stays up between them is critical. Currently hp only reports total loss percentage — it discards the temporal pattern. Users need to see "it drops every ~2 minutes after recovery" at a glance.

## What Changes

- Track alternating UP/DOWN periods during the monitoring session, recording start time and request count for each
- Display a color-coded timeline in the final summary (on Ctrl+C or `-c` count completion)
- Timeline only appears when there were failures — zero noise for clean sessions
- Add a compact duration formatter for human-readable period lengths

## Capabilities

### New Capabilities
- `connection-timeline`: Track alternating UP/DOWN connectivity periods and display a timeline summary showing start times, durations, and request counts for each period

### Modified Capabilities

## Impact

- `main.go`: New `period` struct, new fields on `stats`, main loop state tracking, `printFinal()` timeline output, `fmtDuration()` helper
- `main_test.go`: Tests for `fmtDuration()` and period tracking logic
- No new dependencies, no flag changes, no breaking changes
