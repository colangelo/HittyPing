# hittyping - prettyping-style HTTPS latency monitor

# Default: list available commands
default:
    @just --list

# Build the binary
build:
    go build -o hittyping .

# Install to /usr/local/bin (with ad-hoc signing for macOS)
install: build
    sudo cp hittyping /usr/local/bin/
    sudo codesign --force --sign - /usr/local/bin/hittyping

# Build and run with default target
run *ARGS: build
    ./hittyping {{ARGS}}

# Clean build artifacts
clean:
    rm -f hittyping
