package payday

import "slices"

// validTransitions defines the allowed state transitions for payday events.
var validTransitions = map[Status][]Status{
	StatusScheduled: {StatusPending},
	StatusPending:   {StatusConfirmed, StatusDeferred, StatusSkipped},
	StatusDeferred:  {StatusPending},
}

// ValidTransition reports whether moving from one status to another is allowed.
func ValidTransition(from, to Status) bool {
	targets, ok := validTransitions[from]
	if !ok {
		return false
	}
	return slices.Contains(targets, to)
}
