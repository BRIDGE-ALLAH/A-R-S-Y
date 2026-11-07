package deployments

import "testing"

func TestDeploymentStateTransitions(t *testing.T) {
	valid := [][2]State{
		{StateQueued, StatePreparing},
		{StatePreparing, StateUploading},
		{StateUploading, StateValidating},
		{StateValidating, StateDeploying},
		{StateDeploying, StateVerifying},
		{StateVerifying, StateSucceeded},
		{StateQueued, StateCancelled},
		{StatePreparing, StateFailed},
	}
	for _, pair := range valid {
		got, err := Transition(pair[0], pair[1])
		if err != nil || got != pair[1] {
			t.Fatalf("Transition(%q, %q) = %q, %v", pair[0], pair[1], got, err)
		}
	}
}

func TestDeploymentStateTransitionsRejectInvalidEdges(t *testing.T) {
	invalid := [][2]State{
		{StateQueued, StateSucceeded},
		{StatePreparing, StateSucceeded},
		{StateSucceeded, StateQueued},
		{StateFailed, StatePreparing},
		{StateCancelled, StateDeploying},
		{"unknown", StateQueued},
	}
	for _, pair := range invalid {
		got, err := Transition(pair[0], pair[1])
		if err == nil || got != pair[0] {
			t.Fatalf("Transition(%q, %q) = %q, %v; want rejection", pair[0], pair[1], got, err)
		}
	}
}
