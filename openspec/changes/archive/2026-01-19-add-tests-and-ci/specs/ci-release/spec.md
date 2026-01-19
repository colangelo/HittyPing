# CI/Release Capability

## ADDED Requirements

### Requirement: Continuous Integration

The project SHALL run automated tests on every code change to ensure quality.

#### Scenario: Tests run on push
- **WHEN** code is pushed to any branch
- **THEN** GitHub Actions runs the test suite

#### Scenario: Tests run on pull request
- **WHEN** a pull request is opened or updated
- **THEN** GitHub Actions runs the test suite

#### Scenario: CI uses correct Go version
- **WHEN** CI workflow executes
- **THEN** it uses the latest stable Go version

### Requirement: Automated Release Builds

The project SHALL automatically build and publish cross-platform binaries when a version tag is pushed.

#### Scenario: Release triggered by version tag
- **WHEN** a tag matching pattern v* is pushed
- **THEN** the release workflow executes

#### Scenario: Tests pass before release
- **WHEN** release workflow starts
- **THEN** tests run and must pass before building binaries

#### Scenario: Cross-platform binaries built
- **WHEN** release workflow builds binaries
- **THEN** it produces binaries for darwin-amd64, darwin-arm64, linux-amd64

#### Scenario: HTTP/3 variants built
- **WHEN** release workflow builds binaries
- **THEN** it produces both default and http3 variants for each platform

#### Scenario: Binaries uploaded to GitHub Release
- **WHEN** all binaries are built successfully
- **THEN** they are uploaded as assets to the GitHub Release for that tag

### Requirement: Binary Naming Convention

Release binaries SHALL follow a consistent naming convention for easy identification.

#### Scenario: Default binary naming
- **WHEN** a default (non-http3) binary is built
- **THEN** it is named hp-{os}-{arch} (e.g., hp-darwin-arm64)

#### Scenario: HTTP/3 binary naming
- **WHEN** an http3 variant binary is built
- **THEN** it is named hp-http3-{os}-{arch} (e.g., hp-http3-darwin-arm64)
