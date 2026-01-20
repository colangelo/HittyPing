# Multi-Target Mode Specification

## Overview

Add support for monitoring multiple targets simultaneously, displaying each target's latency bar on its own line with the hostname above.

## Usage

```bash
hp google.com cloudflare.com 1.1.1.1
hp -c 10 dns.google 1.1.1.1          # 10 requests per target
hp -3 -d google.com cloudflare.com   # HTTP/3 with downgrade for all
```

When multiple positional arguments are provided, hp enters multi-target mode.

## Output Format

```
HP multi-target (3 hosts)
Legend: ▁▂▃<150ms ▄▅<400ms ▆▇█>=400ms !fail

google.com [142.250.180.14]
▁▁▂▁▄▃▁▂▁▁▂▃▁▁▂▁▁

cloudflare.com [104.16.132.229]
▁▁▁▁▁▁▂▁▁▁▁▁▁▁▁▁▁

1.1.1.1
▁▁▁▁▂▁▁▁▁▁▁▁▂▁▁▁▁
```

### Header
- Shows "HP multi-target (N hosts)" instead of single target header
- Single shared legend line

### Per-Target Display
- Hostname line with resolved IP (like single-target mode)
- Bar line immediately below
- Blank line between targets for visual separation
- Each bar wraps independently at terminal width

### Final Summary (Ctrl+C)

```
Summary:
                   min    avg    max   loss
google.com          45ms   67ms  512ms   0%
cloudflare.com      12ms   18ms   42ms   0%
1.1.1.1              8ms   14ms   31ms   0%
```

## Implementation

### Architecture

1. **Target struct** - holds per-target state:
   - hostname, resolved IP
   - HTTP client (each may have different protocol state if downgrading)
   - stats (min, avg, max, count, failures)
   - current bar string

2. **Goroutine per target** - each target runs independently:
   - Sends requests at specified interval
   - Updates its own stats
   - Signals main loop to redraw

3. **Display goroutine** - handles terminal output:
   - Receives updates from target goroutines
   - Redraws affected lines using ANSI cursor positioning
   - Handles terminal resize

### Terminal Control

Use ANSI escape sequences for cursor positioning:
- `\033[<row>;0H` - move to row
- `\033[K` - clear to end of line

Each target occupies 3 lines (hostname, bar, blank), so target N starts at row `3 + (N * 3)`.

### Flags

All existing flags apply to all targets:
- `-i/--interval` - same interval for all
- `-t/--timeout` - same timeout for all
- `-c/--count` - N requests per target (total = N × targets)
- `-g/-y` thresholds - shared across all targets
- `-3/-2/-1` protocol - all targets use same starting protocol
- `-d/-D` downgrade - each target downgrades independently

### Edge Cases

- **DNS failure for one target**: Show error, continue with others
- **All targets fail**: Exit with error
- **Single target provided**: Fall back to current single-target mode (no behavior change)
- **Terminal too narrow**: Bars wrap as they do today
- **Terminal too short**: Scroll mode (don't try to update in place)

## Future Enhancements (Out of Scope)

- Per-target thresholds
- Per-target protocols
- Output to multiple files
- Named target groups from config file

## Testing

- Unit tests for Target struct and stats calculation
- Integration test with mock HTTP server
- Manual testing with 2, 5, 10 targets
