# Payment Processing Example

This example demonstrates comprehensive testing of a complex, multi-step payment processing workflow using Tempest's deterministic simulation testing.

## What it shows

- Multi-step workflow with compensation logic (Saga pattern)
- Activity-level fault injection with realistic failure modes
- Business-specific invariants and state validation
- Complex error handling and retry policies
- Activity call tracking and analysis

## Business Logic

The payment workflow implements a complete payment processing saga:

1. **Customer Validation** - Verify customer and payment method
2. **Fund Reservation** - Reserve funds (authorization)
3. **Charge Processing** - Process the actual charge with retries
4. **Notification** - Send customer and merchant notifications (best effort)
5. **Settlement** - Update merchant account

If any critical step fails, compensation logic releases reservations and handles rollback.

## Files

- `workflow.go` - Complex payment processing workflow with saga pattern
- `activities.go` - Payment-related activities (validation, charging, notifications)
- `main.go` - Tempest simulation with activity-level fault injection

## How to run

```bash
go run examples/payment-processing/main.go
```

## Expected behavior

The simulation tests the payment workflow under various failure conditions:

- Customer validation failures (5% rate)
- Payment processor unavailability (15% rate) 
- Network timeouts and delays
- Notification service failures

Business invariants ensure:
- Valid payment status transitions
- Proper compensation when charges fail
- No double-charging scenarios

## Key concepts

- **Activity-level fault injection**: Failures injected into specific activities
- **Realistic failure modes**: Domain-specific errors (insufficient funds, API down)
- **Business invariants**: Domain logic validation, not infrastructure testing
- **Saga pattern testing**: Compensation logic under various failure scenarios
- **Comprehensive reporting**: Activity call analysis and fault injection summaries

## Fault injection patterns

This example demonstrates several fault injection patterns:

```go
// Realistic failure rates
sim.RegisterActivityWithBehavior(activities.ValidateCustomer, simulator.ActivityBehavior{
    FailureRate: 0.05, // 5% validation failures
    CustomErrors: []error{
        errors.New("customer service unavailable"),
        errors.New("invalid customer credentials"),
    },
})

// Timeout simulation
sim.RegisterActivityWithBehavior(activities.ProcessCharge, simulator.ActivityBehavior{
    FailureRate: 0.08, // 8% charge failures
    TimeoutRate: 0.05, // 5% timeout rate
    MinLatency:  200 * time.Millisecond,
    MaxLatency:  1 * time.Second,
})
```

## Business invariants

The example shows how to write business-focused invariants:

```go
sim.RegisterInvariant(workflow.ProcessPayment, simulator.Invariant{
    Name: "payment status is valid",
    Check: func() bool {
        // Validate state transitions make business sense
        return isValidPaymentStatus(workflow.State.PaymentStatus)
    },
})
```

This focuses testing on **business logic correctness** rather than infrastructure behavior.