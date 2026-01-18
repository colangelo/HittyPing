# hittyping - prettyping-style HTTPS latency monitor

# Build the binary
build:
    go build -o hittyping .

# Install to /usr/local/bin
install: build
    sudo cp hittyping /usr/local/bin/

# Build and run with default target
run *ARGS: build
    ./hittyping {{ARGS}}

# Clean build artifacts
clean:
    rm -f hittyping
