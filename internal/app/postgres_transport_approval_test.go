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

func TestDeclarativeTypedDestinationAuthorization_ReusesWithoutPostEffectFailure(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	const planID = "rplan_reusable_authorization"
	if _, err := a.updateState(func(current state) (state, error) {
		current.ReversePlans = append(current.ReversePlans, ReversePlan{
			ID: planID, Mode: reversePlanModeDeclarativeTypedDestinationTransport,
			Status: reversePlanStatusApprovalConsumptionUncertain,
		})
		return current, nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := a.markDeclarativeTypedDestinationPlanExecuted(planID); err != nil {
		t.Fatalf("first completion marker = %v", err)
	}
	if err := a.markDeclarativeTypedDestinationPlanExecuted(planID); err != nil {
		t.Fatalf("reused authorization post-effect completion marker = %v", err)
	}
}

func TestDeclarativeTypedDestinationAuthorization_RejectsChangedOrForeignBindingBeforeIO(t *testing.T) {
	plan := ReversePlan{TransportBindingSHA256: "approved-binding"}
	if err := validateDeclarativeTypedDestinationPlanBinding(plan, "changed-binding"); err == nil {
		t.Fatal("changed or foreign transport binding was accepted")
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

// The transform hash is part of the operator-approved write shape. A plan
// issued for one normalized mapping must become stale before any source or
// destination I/O when the connection changes to another mapping.
func TestPostgresManagedTargetApprovalBindingChangesWithTransformPlanHash(t *testing.T) {
	identity := postgresManagedTargetApprovalBinding{
		Domain: postgresManagedTargetBindingDomain, ConnectionID: "conn", Stream: "events", StreamID: "stream_events",
		Mode: synccontract.ModeFullOverwrite, SourceConnector: "postgres", SourceSchema: "source-schema",
		SourceCredential: "source-revision", SourceConfiguration: "source-config", Destination: "postgres",
		CredentialRevision: "destination-revision", ConfigurationDigest: "destination-config",
	}
	identityMapping, err := hashJSON(identity)
	if err != nil {
		t.Fatal(err)
	}
	transformed := identity
	transformed.TransformPlanHash = "sha256:realistic-projection-and-filter"
	transformedMapping, err := hashJSON(transformed)
	if err != nil {
		t.Fatal(err)
	}
	if identityMapping == transformedMapping {
		t.Fatalf("approval binding hash = %q for distinct transform hashes, want plan/preview/approval scope to change", identityMapping)
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
