## ADDED Requirements

### Requirement: Period tracking
The system SHALL track alternating UP and DOWN periods during monitoring. A period starts when the connection state changes (success→failure or failure→success) or on the first request. Each period SHALL record whether it is UP or DOWN, the start time, and the number of requests in that period.

#### Scenario: First request succeeds
- **WHEN** the first request completes successfully
- **THEN** a new UP period is created with start time of the request and count of 1

#### Scenario: First request fails
- **WHEN** the first request fails
- **THEN** a new DOWN period is created with start time of the request and count of 1

#### Scenario: Consecutive successes
- **WHEN** a request succeeds and the current period is UP
- **THEN** the current period's count is incremented by 1

#### Scenario: Consecutive failures
- **WHEN** a request fails and the current period is DOWN
- **THEN** the current period's count is incremented by 1

#### Scenario: Transition from UP to DOWN
- **WHEN** a request fails and the current period is UP
- **THEN** the current UP period is closed and appended to the completed periods list
- **AND** a new DOWN period is created with start time of the request and count of 1

#### Scenario: Transition from DOWN to UP
- **WHEN** a request succeeds and the current period is DOWN
- **THEN** the current DOWN period is closed and appended to the completed periods list
- **AND** a new UP period is created with start time of the request and count of 1

### Requirement: Timeline summary display
The system SHALL display a timeline section in the final summary when there were any failures during the session. The timeline SHALL NOT be displayed when all requests succeeded.

#### Scenario: Summary with failures shows timeline
- **WHEN** the session ends (Ctrl+C or count limit) and there were failures
- **THEN** the summary includes a "timeline:" section after the existing stats
- **AND** each period is shown on its own line with: start time (HH:MM:SS), UP or DOWN label, duration, and request count
- **AND** UP periods are displayed in green, DOWN periods in red
- **AND** the last period shows "active" instead of duration since it was ongoing at exit

#### Scenario: Clean session hides timeline
- **WHEN** the session ends and there were zero failures
- **THEN** no timeline section is displayed

#### Scenario: Silent mode hides timeline
- **WHEN** the `--silent` flag is set
- **THEN** no timeline section is displayed (along with the rest of the summary)

### Requirement: Compact duration formatting
The system SHALL format durations in a compact human-readable form for the timeline display.

#### Scenario: Seconds only
- **WHEN** a duration is less than 60 seconds
- **THEN** it is formatted as "{N}s" (e.g., "7s", "45s")

#### Scenario: Minutes and seconds
- **WHEN** a duration is 60 seconds or more but less than 1 hour
- **THEN** it is formatted as "{M}m{SS}s" (e.g., "2m13s", "15m00s")

#### Scenario: Hours and minutes
- **WHEN** a duration is 1 hour or more
- **THEN** it is formatted as "{H}h{MM}m" (e.g., "1h02m", "3h45m")

### Requirement: Timeline truncation for flappy connections
The system SHALL truncate the timeline display when the number of periods exceeds a reasonable threshold to prevent excessive output.

#### Scenario: Many periods truncated
- **WHEN** the total number of periods exceeds 20
- **THEN** the first 5 and last 5 periods are shown with a "... N more ..." line in between
