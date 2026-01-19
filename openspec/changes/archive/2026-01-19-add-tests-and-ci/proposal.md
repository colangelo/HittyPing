# Change: Add Unit Tests and CI/Release Workflows

## Why

The hp project currently has no automated testing or continuous integration. This creates risk during development and makes releases a manual process. Adding tests ensures code correctness, while CI/CD automates quality checks and cross-platform binary releases.

## What Changes

- Add unit tests for pure functions (`getEnvInt`, `getURLForProto`, `getBlock`)
- Add GitHub Actions workflow for CI (test on push/PR)
- Add GitHub Actions workflow for automated releases (build cross-platform binaries on tags)

## Impact

- Affected specs: None existing (creates new `testing` and `ci-release` capabilities)
- Affected code:
  - `main_test.go` (new)
  - `.github/workflows/ci.yml` (new)
  - `.github/workflows/release.yml` (new)
- No breaking changes to existing functionality
