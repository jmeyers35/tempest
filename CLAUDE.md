# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Build and Run
- `go build .` - Build the main application
- `go run .` - Run the main application (prints "Hello, Tempest!")

### Examples
- `go run examples/basic/*.go` - Basic DST example with simple workflow
- `go run examples/payment-processing/*.go` - Complex payment processing saga
- `go run examples/fault-injection/*.go` - Comprehensive fault injection demo

### Testing and Development
- `go test ./...` - Run all tests in the project
- `go mod tidy` - Clean up module dependencies

## Architecture Overview

Tempest is a Temporal workflow simulator built in Go that provides deterministic testing of workflows with invariant checking.

### Core Components

**Simulator (`pkg/simulator/simulator.go`)**
- Main simulation engine that ticks time and executes workflows in a controlled environment
- Uses Temporal's test suite for workflow execution
- Supports invariant checking at each tick
- Configuration includes tick duration, random seed, and max tick limits

**Workflow Structure**
- Workflows are registered with the simulator and executed in test environments
- Time advancement is manual via ticks rather than real-time
- Invariants are checked after each tick and will panic if violated

### Key Concepts

- **Ticks**: Discrete time advancement units (configurable duration)
- **Invariants**: State checks that must remain true throughout simulation
- **Test Environment**: Uses Temporal's `testsuite.TestWorkflowEnvironment` for isolated execution

### Development Notes

- Uses a local Temporal SDK replacement (`replace go.temporal.io/sdk => /Users/jacobmeyers/dev/sdk-go`)
- Maximum simulation limit of 1,000,000 ticks to prevent infinite loops
- Workflows run in goroutines with 1-hour test timeout