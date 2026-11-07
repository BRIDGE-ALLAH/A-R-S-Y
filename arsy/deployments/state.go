package deployments

import "fmt"

type State string

const (
	StateQueued    State = "queued"
	StatePreparing State = "preparing"
	StateUploading State = "uploading"
	StateValidating State = "validating"
	StateDeploying State = "deploying"
	StateVerifying State = "verifying"
	StateSucceeded State = "succeeded"
	StateFailed    State = "failed"
	StateCancelled State = "cancelled"
)

var allowedTransitions = map[State]map[State]struct{}{
	StateQueued:    {StatePreparing: {}, StateFailed: {}, StateCancelled: {}},
	StatePreparing: {StateUploading: {}, StateFailed: {}, StateCancelled: {}},
	StateUploading: {StateValidating: {}, StateFailed: {}, StateCancelled: {}},
	StateValidating: {StateDeploying: {}, StateFailed: {}, StateCancelled: {}},
	StateDeploying: {StateVerifying: {}, StateFailed: {}, StateCancelled: {}},
	StateVerifying: {StateSucceeded: {}, StateFailed: {}, StateCancelled: {}},
}

// Transition validates and returns the next deployment state.
func Transition(current, next State) (State, error) {
	if _, ok := allowedTransitions[current][next]; !ok {
		return current, fmt.Errorf("invalid deployment state transition %q -> %q", current, next)
	}
	return next, nil
}
