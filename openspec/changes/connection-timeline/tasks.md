## 1. Data Model

- [ ] 1.1 Add `period` struct to main.go (fields: `up bool`, `start time.Time`, `count int`)
- [ ] 1.2 Add period tracking fields to `stats` struct: `periods []period`, `currentPeriod *period`

## 2. Core Tracking Logic

- [ ] 2.1 Add `recordPeriod(s *stats, up bool)` helper that handles period creation and transitions
- [ ] 2.2 Call `recordPeriod` in the main loop failure branch (after line 281)
- [ ] 2.3 Call `recordPeriod` in the main loop success branch (after line 345)
- [ ] 2.4 Add `closePeriods(s *stats)` helper that closes the active period by appending it to `s.periods`
- [ ] 2.5 Call `closePeriods` before `printFinal` in Ctrl+C handler (line ~208) and count-exit path (line ~373)

## 3. Duration Formatting

- [ ] 3.1 Implement `fmtDuration(d time.Duration) string` with compact format (7s, 2m13s, 1h02m)

## 4. Timeline Display

- [ ] 4.1 Add timeline rendering to `printFinal()`: print "timeline:" header and each period line with start time, UP/DOWN label (color-coded), duration, and count
- [ ] 4.2 Mark the last period with "active" instead of duration
- [ ] 4.3 Only show timeline when `s.failures > 0`
- [ ] 4.4 Implement truncation: show first 5 + last 5 with "... N more ..." when periods exceed 20

## 5. Tests

- [ ] 5.1 Add table-driven tests for `fmtDuration` covering seconds, minutes+seconds, hours+minutes edge cases
- [ ] 5.2 Add tests for `recordPeriod` verifying period creation, continuation, and transitions
- [ ] 5.3 Build and run all tests to verify no regressions
