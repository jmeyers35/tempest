# Tempest Justfile

# Build the main application
build:
    go build ./...

# Run all tests
test:
    go test ./...

# Clean up module dependencies
tidy:
    go mod tidy

# Run basic example
example-basic:
    go run examples/basic/*.go

# Run payment processing example
example-payment:
    go run examples/payment-processing/*.go

# Run fault injection example
example-fault:
    go run examples/fault-injection/*.go

# Run race condition example
example-race:
    go run examples/race-condition/*.go

# Run all examples
examples: example-basic example-payment example-fault example-race

# Clean build artifacts
clean:
    go clean

# Format code
fmt:
    go fmt ./...

# Vet code for issues
vet:
    go vet ./...

# Run full checks (fmt, vet, test)
check: fmt vet test

# Show available commands
help:
    @just --list
