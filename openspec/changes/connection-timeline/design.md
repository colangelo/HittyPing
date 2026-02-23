## Context

hp currently tracks total failures (`s.failures`) as a single counter. When a request fails, the counter increments and a red `!` block is appended. The final summary shows total requests, ok, failed, and loss percentage — but no temporal information about when failures occurred or how they clustered.

The `stats` struct (main.go:86-99) holds all monitoring state. The main loop (main.go:278-383) processes each request result. `printFinal()` (main.go:679-701) renders the exit summary.

## Goals / Non-Goals

**Goals:**
- Track alternating UP (successful) and DOWN (failure) periods with start times and request counts
- Display a color-coded timeline in the final summary showing the connectivity pattern
- Make intermittent outage diagnosis trivial ("drops every ~2min after recovery")
- Zero visual noise when there are no failures

**Non-Goals:**
- No new CLI flags — timeline appears automatically when relevant
- No live/inline timeline display during monitoring — summary only
- No persistence or export of timeline data
- No per-request timestamp tracking — only period boundaries

## Decisions

### 1. Period-based model over event log

Track `[]period` (alternating up/down) rather than individual request timestamps. This is O(transitions) not O(requests), keeps memory bounded for long sessions, and directly represents the pattern users care about.

Alternative: Per-request timestamps — rejected because it's wasteful (most sessions are thousands of requests) and requires post-processing to find patterns.

### 2. Inline state tracking in main loop

Add period transition logic directly in the main loop (after line 280 for failures, after line 344 for successes) rather than a separate goroutine or observer. The main loop already has the success/fail distinction and runs sequentially — no need for additional complexity.

### 3. Close active period before printFinal

Before calling `printFinal()` (at Ctrl+C handler line 208 and count-exit line 373), close the current period by appending it to `s.periods`. This ensures the last period appears in the timeline. The last period gets a special "active" label since it was ongoing at exit.

### 4. Timeline only shown when failures > 0

If all requests succeed, there's only one UP period — showing "timeline: UP 5m (300 ok)" adds no value. Only print the timeline section when `s.failures > 0`, keeping output clean for healthy targets.

### 5. Compact duration formatting

Use a custom `fmtDuration()` returning strings like "7s", "2m13s", "1h02m" rather than Go's default `Duration.String()` which produces "2m13.000000s". The compact format is more readable in the fixed-width timeline layout.

## Risks / Trade-offs

- [Minimal memory overhead] → Each period is ~25 bytes (bool + Time + int). Even with thousands of transitions, this is negligible. No mitigation needed.
- [Timeline output length for very flaky connections] → A connection flapping every second could produce hundreds of periods. → Mitigation: Cap displayed periods (e.g., first 5 + last 5 with "... N more ..." in between) if list exceeds a threshold.
- [Time zone display] → Use local time (time.Now()) which matches user expectations. Users running across time zones can set TZ env var.
