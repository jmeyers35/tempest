# Race Condition Demo

This example demonstrates how Tempest can catch subtle race condition bugs that are difficult to find with traditional unit tests or integration tests.

## The Bug

The workflow implements an inventory management system where multiple concurrent orders can be processed. The bug is a classic "check-then-act" race condition:

1. Order workflow checks if items are available in inventory
2. If available, it reserves the items
3. **BUG**: Between the check and reservation, another concurrent workflow might have already reserved the same items

This leads to **overselling** - more items being reserved than are actually available.

## Why Traditional Tests Miss This

- **Unit tests**: Test individual activities in isolation, missing the concurrency aspect
- **Integration tests**: Usually run single workflows sequentially, not exposing race conditions
- **Temporal test server**: Executes activities synchronously by default, masking timing issues

## How Tempest Catches It

1. **Deterministic simulation**: Multiple workflows run concurrently in controlled time
2. **Invariant checking**: After each tick, we verify that total reservations never exceed available inventory
3. **Reproducible**: Same seed produces same results, making bugs debuggable

## Running the Demo

```bash
go run examples/race-condition/*.go
```

The demo will show the race condition being detected by invariant violations.