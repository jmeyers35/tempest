# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Build and Run (Justfile Commands)
- `just build` - Build the main application (`go build ./...`)
- `just test` - Run all tests (`go test ./...`)
- `just tidy` - Clean up module dependencies (`go mod tidy`)
- `just check` - Run format, vet, and test (`go fmt ./... && go vet ./... && go test ./...`)
- `just fmt` - Format code (`go fmt ./...`)
- `just vet` - Vet code for issues (`go vet ./...`)

### Direct Go Commands
- `go build ./...` - Build all packages
- `go run .` - Run the main application (if it exists)
- `go test ./...` - Run all tests in the project
- `go mod tidy` - Clean up module dependencies

### Examples
- `just example-basic` or `go run examples/basic/*.go` - Basic DST example with simple workflow
- `just example-payment` or `go run examples/payment-processing/*.go` - Complex payment processing saga
- `just example-fault` or `go run examples/fault-injection/*.go` - Comprehensive fault injection demo
- `just example-race` or `go run examples/race-condition/*.go` - Race condition testing example
- `just examples` - Run all examples

## Architecture Overview

Tempest is a Temporal workflow simulator built in Go that provides deterministic testing of workflows with invariant checking.

### Core Components

**Simulator (`pkg/simulator/simulator.go`)**
- Main simulation engine that provides deterministic testing of Temporal workflows
- Uses Temporal's `testsuite.TestWorkflowEnvironment` for isolated execution
- Supports comprehensive fault injection and invariant checking
- Configuration includes tick duration, random seed, failure probabilities, and fault injection rules

**Core Files:**
- `pkg/simulator/simulator.go` - Main simulator implementation
- `pkg/simulator/events.go` - Event system for deterministic execution
- `pkg/simulator/event_queue.go` - Priority queue for event ordering
- `pkg/simulator/invariant.go` - Invariant checking system

### Key Concepts

- **Deterministic Testing**: Reproducible execution using seeded randomness
- **Fault Injection**: Activity-level failures (timeouts, errors, network issues)
- **Business Invariants**: Domain-specific validation rules checked continuously
- **Event-Driven Simulation**: Priority-based event scheduling for consistent execution
- **Activity Behaviors**: Configurable failure rates, latencies, and custom errors

### Development Notes

- Uses a local Temporal SDK replacement (`replace go.temporal.io/sdk => ./sdk-go`)
- Maximum simulation limit of 1,000,000 ticks to prevent infinite loops
- Supports multiple examples: basic, payment-processing, fault-injection, race-condition
- Built with Justfile for command management