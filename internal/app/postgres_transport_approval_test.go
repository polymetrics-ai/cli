package app

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

func TestPostgresManagedTargetAuthorizationLifetimeDefaultsAndRejectsOutOfRangeBeforePlanning(t *testing.T) {
	for _, tt := range []struct {
		name    string
		input   time.Duration
		want    time.Duration
		wantErr error
	}{
		{name: "default is one day", want: 24 * time.Hour},
		{name: "minimum one day", input: 24 * time.Hour, want: 24 * time.Hour},
		{name: "maximum two days", input: 48 * time.Hour, want: 48 * time.Hour},
		{name: "below minimum is refused", input: 24*time.Hour - time.Nanosecond, wantErr: ErrPostgresManagedTargetAuthorizationLifetime},
		{name: "above maximum is refused", input: 48*time.Hour + time.Nanosecond, wantErr: ErrPostgresManagedTargetAuthorizationLifetime},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizePostgresManagedTargetAuthorizationLifetime(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("normalizePostgresManagedTargetAuthorizationLifetime(%s) error = %T %v, want %v", tt.input, err, err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Fatalf("normalizePostgresManagedTargetAuthorizationLifetime(%s) = %s, want %s", tt.input, got, tt.want)
			}
		})
	}
}

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

func TestPostgresManagedTargetDurableAuthorizationValidationDoesNotReusePreviewSealLifetime(t *testing.T) {
	binding := postgresManagedTargetApprovalBinding{
		Domain: postgresManagedTargetBindingDomain, ConnectionID: "conn", Stream: "commits", StreamID: "stream_commits",
		Mode: synccontract.ModeIncrementalUpsert, SourceConnector: "github", SourceSchema: "schema", SourceCredential: "source_revision",
		SourceConfiguration: "source_config", Destination: "postgres", CredentialRevision: "target_revision", ConfigurationDigest: "target_config",
	}
	bindingHash, err := hashJSON(binding)
	if err != nil {
		t.Fatal(err)
	}
	planHash, err := postgresManagedTargetPlanHash(bindingHash, postgresManagedTargetAuthorizationDefaultLifetime)
	if err != nil {
		t.Fatal(err)
	}
	plan := ReversePlan{
		Mode: reversePlanModePostgresManagedTarget, TransportBindingSHA256: bindingHash,
		DestinationConnector: binding.Destination, Action: postgresManagedTargetAction(binding.Mode),
		PlanHash: planHash, AuthorizationLifetime: postgresManagedTargetAuthorizationDefaultLifetime,
		AuthorizationReference: "auth_durable_fixture",
		// A missing or expired plan seal is intentionally not enough to validate
		// a pre-token plan. The durable route instead verifies its separately
		// stored scope identity at every destination unit.
	}
	runtime := connectors.RuntimeConfig{CredentialRevision: binding.CredentialRevision, ConfigurationDigest: binding.ConfigurationDigest}
	instance := &App{registry: connectors.NewRegistry()}
	if _, err := instance.validatePostgresManagedTargetPlanWithSeal(plan, runtime, binding, false); err != nil {
		t.Fatalf("durable authorization validation = %v, want the scope-bound route to outlive the preview seal", err)
	}
	if _, err := instance.validatePostgresManagedTargetPlanWithSeal(plan, runtime, binding, true); err == nil {
		t.Fatal("pre-token plan validation accepted a missing preview seal")
	}
}
