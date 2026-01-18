# hp - prettyping-style HTTP(S) latency monitor

# Default: list available commands
default:
    @just --list

# Build the binary
build:
    go build -o hp .

# Install to /usr/local/bin (with ad-hoc signing for macOS)
install: build
    sudo cp hp /usr/local/bin/
    sudo codesign --force --sign - /usr/local/bin/hp

# Build and run with default target
run *ARGS: build
    ./hp {{ARGS}}

# Clean build artifacts
clean:
    rm -f hp
