## 1. Data Model

- [x] 1.1 Add `period` struct to main.go (fields: `up bool`, `start time.Time`, `count int`)
- [x] 1.2 Add period tracking fields to `stats` struct: `periods []period`, `currentPeriod *period`

## 2. Core Tracking Logic

- [x] 2.1 Add `recordPeriod(s *stats, up bool)` helper that handles period creation and transitions
- [x] 2.2 Call `recordPeriod` in the main loop failure branch (after line 281)
- [x] 2.3 Call `recordPeriod` in the main loop success branch (after line 345)
- [x] 2.4 Add `closePeriods(s *stats)` helper that closes the active period by appending it to `s.periods`
- [x] 2.5 Call `closePeriods` before `printFinal` in Ctrl+C handler (line ~208) and count-exit path (line ~373)

## 3. Duration Formatting

- [x] 3.1 Implement `fmtDuration(d time.Duration) string` with compact format (7s, 2m13s, 1h02m)

## 4. Timeline Display

- [x] 4.1 Add timeline rendering to `printFinal()`: print "timeline:" header and each period line with start time, UP/DOWN label (color-coded), duration, and count
- [x] 4.2 Mark the last period with "active" instead of duration
- [x] 4.3 Only show timeline when `s.failures > 0`
- [x] 4.4 Implement truncation: show first 5 + last 5 with "... N more ..." when periods exceed 20

## 5. Tests

- [x] 5.1 Add table-driven tests for `fmtDuration` covering seconds, minutes+seconds, hours+minutes edge cases
- [x] 5.2 Add tests for `recordPeriod` verifying period creation, continuation, and transitions
- [x] 5.3 Build and run all tests to verify no regressions
