package app

import (
	"errors"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

func TestPostgresManagedTargetStaleApprovalIsTypedAndStateFree(t *testing.T) {
	a := &App{state: state{WorkspaceID: "workspace-before"}}
	before := a.state
	plan := ReversePlan{
		TransportBindingSHA256: "stale-binding",
		DestinationConnector:   "destination",
		Action:                 postgresManagedTargetAction(synccontract.ModeIncrementalUpsert),
	}
	binding := postgresManagedTargetApprovalBinding{
		Domain: postgresManagedTargetBindingDomain, Mode: synccontract.ModeIncrementalUpsert, Destination: "destination",
	}

	if _, err := a.validatePostgresManagedTargetPlan(plan, connectors.RuntimeConfig{}, binding); !errors.Is(err, ErrPostgresManagedTargetApprovalStale) {
		t.Fatalf("stale approval error = %v, want ErrPostgresManagedTargetApprovalStale", err)
	}
	if !reflect.DeepEqual(a.state, before) {
		t.Fatalf("stale approval changed App state: before=%#v after=%#v", before, a.state)
	}
}
