# hp - prettyping-style HTTP(S) latency monitor

# Default: list available commands
default:
    @just --list

# Build the binary (no HTTP/3, ~8MB)
build:
    go build -o hp .

# Build with HTTP/3 support (~20MB)
build-http3:
    go build -tags http3 -o hp .

# Install to /usr/local/bin (with ad-hoc signing for macOS)
install: build
    sudo cp hp /usr/local/bin/
    sudo codesign --force --sign - /usr/local/bin/hp

# Install with HTTP/3 support
install-http3: build-http3
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

# Local CI: fmt check, vet, test
ci:
    @echo "Checking fmt..."
    @test -z "$(go fmt ./...)" || (echo "go fmt needed" && exit 1)
    @echo "Running vet..."
    go vet ./...
    @echo "Running tests..."
    go test -v ./...
    @echo "CI passed"

# Clean build artifacts
clean:
    rm -f hp
