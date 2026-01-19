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

# Run go vet
vet:
    go vet ./...

# Run vulnerability check
vuln:
    go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Local CI: fmt check, vet, vuln, test
ci:
    @echo "Checking fmt..."
    @test -z "$(go fmt ./...)" || (echo "go fmt needed" && exit 1)
    @echo "Running vet..."
    go vet ./...
    @echo "Running vulnerability check..."
    go run golang.org/x/vuln/cmd/govulncheck@latest ./...
    @echo "Running tests..."
    go test -v ./...
    @echo "CI passed"

# Clean build artifacts
clean:
    rm -f hp
