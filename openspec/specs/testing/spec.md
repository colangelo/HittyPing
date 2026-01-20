# testing Specification

## Purpose
TBD - created by archiving change add-tests-and-ci. Update Purpose after archive.
## Requirements
### Requirement: Unit Tests for Pure Functions

The project SHALL have unit tests for all pure functions that can be tested in isolation.

#### Scenario: getEnvInt returns default when env var not set
- **WHEN** the environment variable is not set
- **THEN** getEnvInt returns the provided default value

#### Scenario: getEnvInt returns parsed value for valid integer
- **WHEN** the environment variable contains a valid integer string
- **THEN** getEnvInt returns the parsed integer value

#### Scenario: getEnvInt returns default for invalid input
- **WHEN** the environment variable contains a non-numeric string
- **THEN** getEnvInt returns the provided default value

### Requirement: URL Protocol Selection Tests

The project SHALL have tests verifying correct URL scheme selection based on protocol level.

#### Scenario: HTTP/1.1 uses http scheme
- **WHEN** protocol level is protoHTTP1
- **THEN** getURLForProto returns URL with http:// scheme

#### Scenario: HTTPS protocols use https scheme
- **WHEN** protocol level is protoHTTPS, protoHTTP2, or protoHTTP3
- **THEN** getURLForProto returns URL with https:// scheme

### Requirement: Latency Block Visualization Tests

The project SHALL have tests verifying correct block character and color selection based on latency.

#### Scenario: Green zone visualization
- **WHEN** latency is below greenThreshold
- **THEN** getBlock returns a green-colored block character (index 0-2)

#### Scenario: Yellow zone visualization
- **WHEN** latency is between greenThreshold and yellowThreshold
- **THEN** getBlock returns a yellow-colored block character (index 3-4)

#### Scenario: Red zone visualization
- **WHEN** latency is at or above yellowThreshold
- **THEN** getBlock returns a red-colored block character (index 5-7)

#### Scenario: Threshold boundary handling
- **WHEN** latency equals exactly a threshold value
- **THEN** getBlock returns the correct zone color for that boundary

