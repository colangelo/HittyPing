# hp - prettyping-style HTTP(S) latency monitor

# Default: list available commands
default:
    @just --list

# Build the binary (~10MB, includes HTTP/3)
build:
    go build -o hp .

# Install to /usr/local/bin (with ad-hoc signing for macOS)
install: build
    sudo cp hp /usr/local/bin/
    sudo codesign --force --sign - /usr/local/bin/hp

# Uninstall from /usr/local/bin
uninstall:
    sudo rm -f /usr/local/bin/hp

# Build and run with default target
run *ARGS: build
    ./hp {{ARGS}}

# Run tests
test:
    go test -v ./...

# Format code
fmt:
    go fmt ./...

# Run golangci-lint
lint:
    go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run

# Run vulnerability check
vuln:
    go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Local CI: lint, vuln, test
ci:
    @echo "Running lint..."
    go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run
    @echo "Running vulnerability check..."
    go run golang.org/x/vuln/cmd/govulncheck@latest ./...
    @echo "Running tests..."
    go test -v ./...
    @echo "CI passed"

# Clean build artifacts
clean:
    rm -f hp

# Download and verify latest release for current platform
verify-release:
    #!/usr/bin/env bash
    set -euo pipefail
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
    BIN="hp-${OS}-${ARCH}"
    BASE_URL="https://github.com/colangelo/HittyPing/releases/latest/download"
    echo "Downloading ${BIN}..."
    curl -sLO "${BASE_URL}/${BIN}"
    curl -sLO "${BASE_URL}/${BIN}.sig"
    curl -sLO "${BASE_URL}/${BIN}.pem"
    echo "Verifying signature..."
    cosign verify-blob \
      --signature "${BIN}.sig" \
      --certificate "${BIN}.pem" \
      --certificate-oidc-issuer https://token.actions.githubusercontent.com \
      --certificate-identity-regexp 'github.com/colangelo/HittyPing' \
      "${BIN}"
    echo "Verified! Binary: ${BIN}"

# Bump version (major, minor, or patch)
bump PART="patch":
    #!/usr/bin/env bash
    set -euo pipefail
    CURRENT=$(grep 'const version = ' main.go | sed 's/.*"\(.*\)".*/\1/')
    IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"
    case "{{PART}}" in
        major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
        minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
        patch) PATCH=$((PATCH + 1)) ;;
        *) echo "Usage: just bump [major|minor|patch]"; exit 1 ;;
    esac
    NEW="${MAJOR}.${MINOR}.${PATCH}"
    sed -i '' "s/const version = \"$CURRENT\"/const version = \"$NEW\"/" main.go
    echo "Bumped version: $CURRENT → $NEW"

# Verify installed hp binary (from brew/scoop) against release signatures
verify-installed:
    #!/usr/bin/env bash
    set -euo pipefail
    HP_PATH=$(which hp 2>/dev/null || whence hp 2>/dev/null) || { echo "hp not found in PATH"; exit 1; }
    VERSION=$(${HP_PATH} --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+')
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
    BIN="hp-${OS}-${ARCH}"
    BASE_URL="https://github.com/colangelo/HittyPing/releases/download/v${VERSION}"
    echo "Found: ${HP_PATH} (v${VERSION})"
    echo "Downloading signatures for ${BIN}..."
    curl -sLO "${BASE_URL}/${BIN}.sig"
    curl -sLO "${BASE_URL}/${BIN}.pem"
    echo "Verifying..."
    cosign verify-blob \
      --signature "${BIN}.sig" \
      --certificate "${BIN}.pem" \
      --certificate-oidc-issuer https://token.actions.githubusercontent.com \
      --certificate-identity-regexp 'github.com/colangelo/HittyPing' \
      "${HP_PATH}"
    rm -f "${BIN}.sig" "${BIN}.pem"
    echo "Verified! ${HP_PATH} matches signed release v${VERSION}"
