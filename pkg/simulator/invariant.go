package simulator

// Invariant represents an invariant throughout a Workflow's execution.
type Invariant struct {
	Name  string
	Check InvariantCheck
}

// InvariantCheck represents an invariant in the Workflow's state.
// It should return true if the invariant is upheld.
type InvariantCheck func() bool
