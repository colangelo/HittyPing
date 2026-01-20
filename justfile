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
