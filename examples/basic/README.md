# Basic Tempest Example

This example demonstrates the fundamentals of using Tempest for deterministic simulation testing of Temporal workflows.

## What it shows

- Simple workflow with state mutation
- Basic invariant checking
- Deterministic simulation with seeded randomness
- Event tracking and reporting

## Files

- `workflow.go` - A simple workflow that modifies account balance
- `main.go` - Tempest simulation setup and execution

## How to run

```bash
go run examples/basic/main.go
```

## Expected output

The simulation will run the workflow with deterministic timing and check that the balance invariant holds. Since the workflow decreases the balance below zero, the invariant will fail and the simulation will panic - this demonstrates that Tempest correctly catches business logic violations.

## Key concepts

- **Deterministic execution**: Same seed produces identical results
- **Invariant checking**: Business rules validated at each simulation step  
- **Event tracking**: Complete history of simulation events for debugging
- **Reproducibility**: Random calls are tracked for full reproducibility

## Next steps

- See `examples/payment-processing/` for a more complex workflow example
- See `examples/fault-injection/` for advanced fault injection patterns