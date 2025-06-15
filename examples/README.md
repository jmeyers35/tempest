# Tempest Examples

This directory contains examples demonstrating Tempest's deterministic simulation testing capabilities for Temporal workflows.

## Getting Started

If you're new to Tempest, start with the basic example:

```bash
go run examples/basic/main.go
```

## Examples

### [Basic](./basic/)
**Start here** - Demonstrates fundamental Tempest concepts:
- Simple workflow simulation
- Deterministic execution with seeded randomness
- Basic invariant checking
- Event tracking and reporting

### [Payment Processing](./payment-processing/) 
**Complex workflows** - Shows real-world payment processing saga:
- Multi-step workflow with compensation logic
- Activity-level fault injection
- Business-specific invariants
- Comprehensive error handling

### [Fault Injection](./fault-injection/)
**Advanced testing** - Comprehensive fault injection patterns:
- Multiple fault types (errors, timeouts, network issues)
- Configurable fault probabilities
- Manual fault injection for targeted testing
- Detailed fault analysis and reporting

## Key Concepts

### Deterministic Simulation Testing (DST)
Tempest leverages Temporal's built-in determinism to enable comprehensive testing:
- **Reproducible**: Same seed produces identical results
- **Exhaustive**: Test thousands of scenarios automatically  
- **Business-focused**: Test workflow logic, not infrastructure
- **Realistic**: Domain-specific faults and failure modes

### Activity-Level Fault Injection
Test how workflows handle realistic failures:
```go
sim.RegisterActivityWithBehavior(activity, simulator.ActivityBehavior{
    FailureRate: 0.1,  // 10% failure rate
    TimeoutRate: 0.05, // 5% timeout rate
    CustomErrors: []error{errors.New("service unavailable")},
})
```

### Business Invariants
Validate domain logic correctness:
```go
sim.RegisterInvariant(workflow, simulator.Invariant{
    Name: "money is never lost",
    Check: func() bool {
        return totalDebits == totalCredits
    },
})
```

## Running Examples

Each example is self-contained and can be run independently:

```bash
# Basic concepts
go run examples/basic/main.go

# Complex workflow testing  
go run examples/payment-processing/main.go

# Advanced fault injection
go run examples/fault-injection/main.go
```

## Next Steps

After exploring these examples:

1. **Adapt for your workflows** - Use these patterns with your own Temporal workflows
2. **Add business invariants** - Define domain-specific validation rules
3. **Configure fault injection** - Set realistic failure rates for your environment
4. **Integrate with CI/CD** - Run DST as part of your testing pipeline

## Learn More

- [Tempest Documentation](../../README.md)
- [Temporal Samples](https://github.com/temporalio/samples-go) - More Temporal workflow patterns
- [TigerBeetle](https://github.com/tigerbeetle/tigerbeetle) - Inspiration for deterministic testing