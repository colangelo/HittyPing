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

# Fuzz the input parsers (30s each)
fuzz:
    go test -run '^$' -fuzz '^FuzzTruncateToWidth$' -fuzztime 30s .
    go test -run '^$' -fuzz '^FuzzHostname$' -fuzztime 30s .

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

# Trigger beta build from dev branch (requires gh CLI)
beta:
    command gh workflow run beta.yml --ref dev
    @echo "Beta workflow triggered. Monitor: command gh run watch"

# Download and verify latest release for current platform
verify-release:
    #!/usr/bin/env bash
    set -euo pipefail
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
    BIN="hp-${OS}-${ARCH}"
    BASE_URL="https://github.com/colangelo/HittyPing/releases/latest/download"
    # just runs recipes in the repo root: download elsewhere, never into the checkout
    TMP="${TMPDIR:-/tmp}"
    DIR=$(mktemp -d "${TMP%/}/hp-verify-release.XXXXXX")
    cd "$DIR"
    echo "Downloading ${BIN} into ${DIR}..."
    curl -fsSLO "${BASE_URL}/${BIN}"
    ID=(--certificate-oidc-issuer https://token.actions.githubusercontent.com --certificate-identity-regexp 'github.com/colangelo/HittyPing')
    if curl -fsSLO "${BASE_URL}/${BIN}.sigstore.json" 2>/dev/null; then
      echo "Verifying Sigstore bundle..."
      cosign verify-blob --bundle "${BIN}.sigstore.json" "${ID[@]}" "${BIN}"
    else
      echo "No Sigstore bundle (release before v0.8.11): verifying .sig/.pem..."
      curl -fsSLO "${BASE_URL}/${BIN}.sig"
      curl -fsSLO "${BASE_URL}/${BIN}.pem"
      cosign verify-blob --signature "${BIN}.sig" --certificate "${BIN}.pem" "${ID[@]}" "${BIN}"
    fi
    chmod +x "${BIN}"
    echo "Verified! Binary: ${DIR}/${BIN}"

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
    # Signatures go to a temp dir that is removed even when verification fails
    TMP="${TMPDIR:-/tmp}"
    DIR=$(mktemp -d "${TMP%/}/hp-verify-installed.XXXXXX")
    trap 'rm -rf "$DIR"' EXIT
    echo "Downloading signatures for ${BIN}..."
    ID=(--certificate-oidc-issuer https://token.actions.githubusercontent.com --certificate-identity-regexp 'github.com/colangelo/HittyPing')
    if curl -fsSL -o "${DIR}/${BIN}.sigstore.json" "${BASE_URL}/${BIN}.sigstore.json" 2>/dev/null; then
      echo "Verifying Sigstore bundle..."
      cosign verify-blob --bundle "${DIR}/${BIN}.sigstore.json" "${ID[@]}" "${HP_PATH}"
    else
      echo "No Sigstore bundle (release before v0.8.11): verifying .sig/.pem..."
      curl -fsSL -o "${DIR}/${BIN}.sig" "${BASE_URL}/${BIN}.sig"
      curl -fsSL -o "${DIR}/${BIN}.pem" "${BASE_URL}/${BIN}.pem"
      cosign verify-blob --signature "${DIR}/${BIN}.sig" --certificate "${DIR}/${BIN}.pem" "${ID[@]}" "${HP_PATH}"
    fi
    echo "Verified! ${HP_PATH} matches signed release v${VERSION}"

# Check that verify-release and verify-installed leave the checkout untouched (needs network, cosign, an installed hp)
test-verify:
    #!/usr/bin/env bash
    set -euo pipefail
    before=$(git status --porcelain --ignored)
    out=$(just verify-release 2>&1) || { echo "$out"; echo "FAIL: verify-release exited non-zero"; exit 1; }
    bin=$(printf '%s\n' "$out" | sed -n 's/^Verified! Binary: //p')
    [[ -x "$bin" && "$bin" != "$PWD"/* ]] || { echo "$out"; echo "FAIL: binary '$bin' missing or inside the checkout"; exit 1; }
    "$bin" --skill | grep -q '^name: hp$' || { echo "FAIL: downloaded binary does not print the skill"; exit 1; }
    just verify-installed >/dev/null 2>&1 || { echo "FAIL: verify-installed exited non-zero"; exit 1; }
    after=$(git status --porcelain --ignored)
    [[ "$before" == "$after" ]] || { diff <(echo "$before") <(echo "$after"); echo "FAIL: the checkout changed"; exit 1; }
    echo "PASS: both recipes verified and the checkout is unchanged (binary: $bin)"
