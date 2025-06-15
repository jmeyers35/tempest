# Tempest

**Deterministic Simulation Testing for Temporal Workflows**

Tempest is a powerful testing framework that enables comprehensive, deterministic testing of Temporal workflows through controlled simulation and fault injection. Build confidence in your distributed systems by testing thousands of scenarios that would be impossible to reproduce in traditional testing environments.

## Why Tempest?

Testing distributed workflows is notoriously difficult. Traditional approaches suffer from:

- **Flaky tests** due to timing and race conditions
- **Limited fault coverage** - can't easily simulate rare failure modes  
- **Non-reproducible failures** that are hard to debug
- **Insufficient business logic validation** under adverse conditions

Tempest solves these problems by providing:

✅ **Deterministic execution** - same seed, same results, every time  
✅ **Comprehensive fault injection** - test realistic failure scenarios  
✅ **Business invariant validation** - ensure correctness under all conditions  
✅ **Reproducible results** - debug issues with confidence  

## Quick Start

### Installation

```bash
go get github.com/jmeyers35/tempest
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/jmeyers35/tempest/pkg/simulator"
    "go.temporal.io/sdk/workflow"
)

// Your existing Temporal workflow
func PaymentWorkflow(ctx workflow.Context, amount int) error {
    // Standard Temporal workflow code
    return workflow.ExecuteActivity(ctx, ProcessPayment, amount).Get(ctx, nil)
}

func ProcessPayment(ctx context.Context, amount int) error {
    // Your payment processing logic
    return nil
}

func main() {
    // Configure the simulator
    config := simulator.DefaultConfig()
    config.Seed = 12345
    config.TickIncrement = 1 * time.Minute
    config.FailureProbability = 0.1

    sim := simulator.New(config)

    // Register your workflow and activities
    sim.RegisterWorkflow(PaymentWorkflow)
    sim.RegisterActivity(ProcessPayment)

    // Add business invariants
    sim.RegisterInvariant(PaymentWorkflow, simulator.Invariant{
        Name: "payment amount must be positive",
        Check: func() bool {
            return amount > 0
        },
    })

    // Run the simulation
    result := sim.Run(PaymentWorkflow, 100)
    
    fmt.Printf("Simulation completed: %+v\n", result)
}
```

## Key Features

### 🎯 Deterministic Testing
- **Reproducible results** using seeded randomness
- **Manual time advancement** through configurable ticks
- **Event-driven simulation** with priority-based scheduling

### 💥 Comprehensive Fault Injection  
- **Activity-level failures** - timeouts, errors, custom exceptions
- **Network simulation** - delays, partitions, connectivity issues
- **Configurable probabilities** - model real-world failure rates
- **Manual fault injection** - target specific scenarios

### 🔍 Business Logic Validation
- **Invariant checking** - validate business rules at each simulation step
- **State mutation tracking** - monitor workflow state changes
- **Activity call monitoring** - track invocations, results, and latencies
- **Complete audit trail** - full event history for debugging

### 🏗️ Temporal Integration
- Uses Temporal's `TestWorkflowEnvironment` for isolation
- **Zero code changes** to existing workflows
- **Activity mocking** and behavior configuration
- **Retry policy support** - test timeout and retry scenarios

## Examples

### Basic Workflow Testing

See [`examples/basic/`](examples/basic/) for a complete introductory example.

### Complex Saga Pattern

The [`examples/payment-processing/`](examples/payment-processing/) demonstrates testing a complex payment processing saga with multiple steps, compensation logic, and business invariants.

### Advanced Fault Injection  

The [`examples/fault-injection/`](examples/fault-injection/) shows comprehensive fault injection patterns and reporting.

## Configuration

### Simulator Configuration

```go
config := simulator.SimulatorConfig{
    Seed:              12345,                // Reproducibility seed
    TickIncrement:     1 * time.Minute,     // Time progression rate
    MaxTicks:          1000,                // Simulation limit
    FailureProbability: 0.1,                // 10% random failure rate
    EnableEventJitter:  true,               // Add timing variation
    FaultConfigs:      []FaultConfig{...},  // Fault injection rules
}
```

### Activity Behavior Configuration

```go
sim.RegisterActivityWithBehavior(ProcessPayment, simulator.ActivityBehavior{
    Name:         "ProcessPayment",
    FailureRate:  0.1,                     // 10% failure rate
    TimeoutRate:  0.05,                    // 5% timeout rate
    MinLatency:   50 * time.Millisecond,   // Minimum response time
    MaxLatency:   200 * time.Millisecond,  // Maximum response time
    CustomErrors: []error{                 // Domain-specific errors
        errors.New("insufficient funds"),
        errors.New("invalid payment method"),
    },
})
```

### Business Invariants

```go
sim.RegisterInvariant(PaymentWorkflow, simulator.Invariant{
    Name: "payment state consistency",
    Check: func() bool {
        // Validate business rules
        return paymentState.IsConsistent()
    },
})
```

## Development

### Building and Testing

```bash
# Build the project
go build .

# Run the main application
go run .

# Run all tests
go test ./...

# Clean up dependencies
go mod tidy
```

### Running Examples

```bash
# Basic example
go run examples/basic/*.go

# Payment processing saga
go run examples/payment-processing/*.go

# Fault injection demo
go run examples/fault-injection/*.go
```

## How It Works

Tempest creates a controlled simulation environment where:

1. **Time is deterministic** - advanced manually through discrete ticks
2. **Events are ordered** - priority queue ensures consistent execution
3. **Randomness is seeded** - same seed produces identical results  
4. **Faults are injected** - realistic failure scenarios are simulated
5. **Invariants are checked** - business rules are validated continuously

This approach enables testing scenarios that would be nearly impossible to reproduce reliably in traditional testing environments.

## Use Cases

### Development and Testing
- **Unit testing** workflows with comprehensive coverage
- **Integration testing** with realistic fault conditions
- **Regression testing** with deterministic results

### Production Readiness
- **Chaos engineering** - validate resilience before deployment
- **Capacity planning** - understand behavior under load and failures
- **Incident prevention** - identify edge cases before they occur

### Debugging and Analysis
- **Root cause analysis** - reproduce issues deterministically
- **Performance testing** - measure latency and throughput
- **Behavioral analysis** - understand workflow execution patterns

## Best Practices

### Test Design
- **Start simple** - begin with basic scenarios before adding complexity
- **Use realistic seeds** - different seeds reveal different failure modes
- **Define clear invariants** - specify business rules that must always hold
- **Model real failure rates** - use production data to configure fault injection

### Debugging
- **Use consistent seeds** - reproduce issues reliably during debugging
- **Enable detailed logging** - leverage event history for analysis
- **Isolate failures** - test individual activities and combinations
- **Validate incrementally** - add complexity gradually

### Production Use
- **Automate testing** - integrate into CI/CD pipelines
- **Monitor coverage** - ensure all code paths are tested
- **Update scenarios** - evolve tests as business logic changes
- **Document findings** - share insights with the team

---

**Built with ❤️ for the Temporal community**

For questions, issues, or feature requests, please [open an issue](https://github.com/jmeyers35/tempest/issues) or start a [discussion](https://github.com/jmeyers35/tempest/discussions).
