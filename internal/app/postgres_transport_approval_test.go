package app

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

func TestPostgresManagedTargetMissingAndConsumedApprovalErrorsAreTyped(t *testing.T) {
	a := &App{state: state{ReversePlans: []ReversePlan{{
		ID: "rplan_consumed", Mode: reversePlanModePostgresManagedTarget, Status: "executed",
	}}}}
	before := a.state
	if _, err := a.authorizePostgresManagedTargetTransport(context.Background(), Connection{}, "issues", synctransport.DestinationApproval{}, connectors.RuntimeConfig{}); !errors.Is(err, ErrPostgresManagedTargetApprovalRequired) {
		t.Fatalf("missing approval error = %T %v, want ErrPostgresManagedTargetApprovalRequired", err, err)
	}
	_, err := a.authorizePostgresManagedTargetTransport(context.Background(), Connection{}, "issues", synctransport.DestinationApproval{
		PlanID: "rplan_consumed", ApprovalToken: "opaque-replay-fixture",
	}, connectors.RuntimeConfig{})
	var replay *AuthorizationTokenReplayError
	if !errors.As(err, &replay) {
		t.Fatalf("consumed approval error = %T %v, want AuthorizationTokenReplayError", err, err)
	}
	if !reflect.DeepEqual(a.state, before) {
		t.Fatalf("approval refusals changed App state: before=%#v after=%#v", before, a.state)
	}
}

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
