# Tasks: Add Unit Tests and CI/Release Workflows

## Phase 1: Unit Tests

### 1.1 Create Test File Structure

- [ ] 1.1.1 Create `main_test.go` with package declaration and imports

### 1.2 Test getEnvInt Function

- [ ] 1.2.1 Test returns default when env var not set
- [ ] 1.2.2 Test returns parsed value when env var is valid integer
- [ ] 1.2.3 Test returns default when env var is invalid (non-numeric)

### 1.3 Test getURLForProto Function

- [ ] 1.3.1 Test returns http:// for protoHTTP1
- [ ] 1.3.2 Test returns https:// for protoHTTPS, protoHTTP2, protoHTTP3

### 1.4 Test getBlock Function

- [ ] 1.4.1 Test green zone blocks (0-2) for latencies below greenThreshold
- [ ] 1.4.2 Test yellow zone blocks (3-4) for latencies between green and yellow thresholds
- [ ] 1.4.3 Test red zone blocks (5-7) for latencies at or above yellowThreshold
- [ ] 1.4.4 Test edge cases: minLatency boundary, threshold boundaries

### 1.5 Verify Tests Pass

- [ ] 1.5.1 Run `go test -v ./...` and confirm all tests pass

## Phase 2: CI Workflow

### 2.1 Create CI Workflow

- [ ] 2.1.1 Create `.github/workflows/ci.yml`
- [ ] 2.1.2 Configure triggers for push and pull_request events
- [ ] 2.1.3 Set up Go (latest stable) environment
- [ ] 2.1.4 Add test execution step

### 2.2 Verify CI

- [ ] 2.2.1 Push to dev branch and verify workflow runs

## Phase 3: Release Workflow

### 3.1 Create Release Workflow

- [ ] 3.1.1 Create `.github/workflows/release.yml`
- [ ] 3.1.2 Configure trigger for v* tags
- [ ] 3.1.3 Define build matrix (darwin/linux × amd64/arm64 × default/http3)
- [ ] 3.1.4 Add test step before build
- [ ] 3.1.5 Add build steps for default and http3 variants
- [ ] 3.1.6 Configure `softprops/action-gh-release` for artifact upload

### 3.2 Verify Release

- [ ] 3.2.1 Create test tag and verify workflow completes
- [ ] 3.2.2 Verify binaries appear on GitHub Releases page
- [ ] 3.2.3 Clean up test tag/release after verification
