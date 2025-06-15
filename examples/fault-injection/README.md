# Fault Injection Example

This example demonstrates Tempest's comprehensive fault injection capabilities for testing workflow resilience under various failure conditions.

## What it shows

- Multiple types of fault injection (activity errors, timeouts, network issues)
- Configurable fault probabilities and custom error messages
- Manual fault injection for specific test scenarios
- Comprehensive fault tracking and analysis
- Integration with business logic testing

## Fault injection types

### Activity-level faults
- **Error injection**: Simulate service unavailability, business rule violations
- **Timeout injection**: Test timeout handling and retry policies
- **Latency simulation**: Variable response times for performance testing

### Infrastructure faults  
- **Network delays**: Simulate network partition and connectivity issues
- **Service outages**: Database unavailability, external API failures
- **Resource exhaustion**: Memory limits, connection pool exhaustion

## Files

- `main.go` - Comprehensive fault injection demonstration using the payment workflow

## How to run

```bash
go run examples/fault-injection/main.go
```

## Configuration examples

### Probabilistic fault injection
```go
config.FaultConfigs = []simulator.FaultConfig{
    {
        Type:        simulator.FaultActivityError,
        Probability: 0.05, // 5% of activities fail
    },
    {
        Type:        simulator.FaultNetworkDelay,
        Probability: 0.1,  // 10% network delays
    },
}
```

### Activity-specific behavior
```go
sim.RegisterActivityWithBehavior(activities.ProcessCharge, simulator.ActivityBehavior{
    Name:        "ProcessCharge",
    FailureRate: 0.08, // 8% failure rate
    TimeoutRate: 0.05, // 5% timeout rate
    CustomErrors: []error{
        errors.New("payment processor unavailable"),
        errors.New("card declined"),
        errors.New("fraud detection triggered"),
    },
})
```

### Manual fault injection
```go
// Inject specific faults for targeted testing
sim.InjectActivityFault("bank_api_call", errors.New("bank API unavailable"))
sim.InjectNetworkDelay("database_connection", 2*time.Second)
```

## Expected output

The simulation will show:
- Detailed fault injection statistics
- Activity call success/failure rates
- Average latencies and timeout occurrences
- Business invariant validation results
- Complete fault history for debugging

## Key concepts

- **Realistic failure modes**: Domain-specific errors that workflows must handle
- **Configurable probabilities**: Control fault injection rates for different scenarios
- **Deterministic reproduction**: Same seed produces identical fault patterns
- **Business logic focus**: Test workflow resilience, not infrastructure reliability
- **Comprehensive reporting**: Detailed analysis of fault injection effectiveness

## Testing strategies

This example demonstrates several fault injection strategies:

1. **Probabilistic testing**: Random faults based on configured probabilities
2. **Targeted testing**: Manual injection of specific failure scenarios  
3. **Comprehensive coverage**: Multiple fault types in single simulation
4. **Business validation**: Invariants ensure business logic handles faults correctly

The goal is to build confidence that workflows handle production failures gracefully.