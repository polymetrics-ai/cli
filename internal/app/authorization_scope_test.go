package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
)

func TestAuthorizationScopeIdentityIsContentFree(t *testing.T) {
	ctx := context.Background()
	a, plan := setupApprovedReversePlan(t, ctx)

	before, err := a.AuthorizationScopeForReversePlan(ctx, plan.ID)
	if err != nil {
		t.Fatalf("AuthorizationScopeForReversePlan(before) error = %v", err)
	}
	beforeIdentity, err := app.AuthorizationScopeIdentity(before)
	if err != nil {
		t.Fatalf("AuthorizationScopeIdentity(before) error = %v", err)
	}

	rows, err := a.QueryTable(ctx, app.QueryTableRequest{Table: "sample_customers", Limit: 10})
	if err != nil {
		t.Fatalf("QueryTable() error = %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("QueryTable() returned no source rows to vary")
	}
	beforeCount := len(rows)
	rows[0]["email"] = "changed-content@example.test"
	rows = append(rows, connectors.Record{
		"id":         "new-record",
		"name":       "New record",
		"email":      "new-record@example.test",
		"updated_at": time.Now().UTC().Format(time.RFC3339Nano),
	})
	if err := writeWarehouseRows(t, a, "sample_customers", rows); err != nil {
		t.Fatalf("writeWarehouseRows() error = %v", err)
	}
	if len(rows) != beforeCount+1 {
		t.Fatalf("changed payload record count = %d, want %d", len(rows), beforeCount+1)
	}

	after, err := a.AuthorizationScopeForReversePlan(ctx, plan.ID)
	if err != nil {
		t.Fatalf("AuthorizationScopeForReversePlan(after) error = %v", err)
	}
	afterIdentity, err := app.AuthorizationScopeIdentity(after)
	if err != nil {
		t.Fatalf("AuthorizationScopeIdentity(after) error = %v", err)
	}
	if afterIdentity != beforeIdentity {
		t.Fatalf("scope identity changed with record content/count/timestamp: before=%s after=%s", beforeIdentity, afterIdentity)
	}
}

func TestAuthorizationScopeIdentityChangesForEveryBoundProperty(t *testing.T) {
	expiresAt := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	base := app.AuthorizationScope{
		SourceConnection:              "source-connection",
		DestinationConnection:         "destination-connection",
		DestinationCredentialRevision: "credential-revision-one",
		StreamTables: []app.AuthorizationStreamTable{{
			Stream: "customers", SourceTable: "warehouse_customers", DestinationTable: "customer_records",
		}},
		FieldMappings:                  map[string]string{"email": "email", "id": "external_id"},
		WriteAction:                    "upsert",
		DestinationConfigurationDigest: "configuration-digest-one",
		EnabledOperations:              []string{"upsert"},
		ConfirmationPolicy:             connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
		ExpiresAt:                      expiresAt,
	}
	want, err := app.AuthorizationScopeIdentity(base)
	if err != nil {
		t.Fatalf("AuthorizationScopeIdentity(base) error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*app.AuthorizationScope)
	}{
		{"source connection", func(scope *app.AuthorizationScope) { scope.SourceConnection = "other-source" }},
		{"destination connection", func(scope *app.AuthorizationScope) { scope.DestinationConnection = "other-destination" }},
		{"credential revision", func(scope *app.AuthorizationScope) { scope.DestinationCredentialRevision = "credential-revision-two" }},
		{"stream table set", func(scope *app.AuthorizationScope) { scope.StreamTables[0].SourceTable = "warehouse_orders" }},
		{"field mappings", func(scope *app.AuthorizationScope) { scope.FieldMappings["email"] = "contact_email" }},
		{"write action", func(scope *app.AuthorizationScope) { scope.WriteAction = "replace" }},
		{"destination configuration", func(scope *app.AuthorizationScope) { scope.DestinationConfigurationDigest = "configuration-digest-two" }},
		{"enabled operations", func(scope *app.AuthorizationScope) { scope.EnabledOperations = []string{"upsert", "delete"} }},
		{"confirmation policy", func(scope *app.AuthorizationScope) { scope.ConfirmationPolicy = connectors.WriteConfirmation{} }},
		{"expiry", func(scope *app.AuthorizationScope) { scope.ExpiresAt = scope.ExpiresAt.Add(time.Minute) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := cloneAuthorizationScope(base)
			tt.mutate(&scope)
			got, err := app.AuthorizationScopeIdentity(scope)
			if err != nil {
				t.Fatalf("AuthorizationScopeIdentity() error = %v", err)
			}
			if got == want {
				t.Fatalf("scope identity = %s after changing %s, want inequality", got, tt.name)
			}
		})
	}
}

func TestRunReverseETLAllowsIdenticalAuthorizationScopeWithoutToken(t *testing.T) {
	ctx := context.Background()
	a, plan, writes := setupAuthorizedReversePlan(t, ctx)
	if got := *writes; got != 1 {
		t.Fatalf("first approved provider writes = %d, want 1", got)
	}

	rows, err := a.QueryTable(ctx, app.QueryTableRequest{Table: "repo_deletes", Limit: 10})
	if err != nil {
		t.Fatalf("QueryTable() error = %v", err)
	}
	rows[0]["id"] = "repo-content-changed"
	rows = append(rows, connectors.Record{
		"id":         "repo-added-later",
		"updated_at": time.Now().UTC().Format(time.RFC3339Nano),
	})
	if err := writeWarehouseRows(t, a, "repo_deletes", rows); err != nil {
		t.Fatalf("writeWarehouseRows() error = %v", err)
	}

	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID})
	if err != nil {
		t.Fatalf("RunReverseETL(unattended) error = %v", err)
	}
	if run.Status != "completed" || run.RecordsSucceeded != 2 {
		t.Fatalf("unattended reverse run = %+v, want two completed records", run)
	}
	if got := *writes; got != 3 {
		t.Fatalf("provider writes after unattended run = %d, want 3; scope authorization must dispatch the changed payload", got)
	}
}

func TestRunReverseETLRefusesInvalidAuthorizationBeforeProviderWrite(t *testing.T) {
	ctx := context.Background()

	t.Run("changed scope", func(t *testing.T) {
		a, plan, writes := setupAuthorizedReversePlan(t, ctx)
		before := *writes
		mutateStoredReversePlan(t, a.ProjectDir(), plan.ID, func(stored map[string]any) {
			stored["mappings"] = map[string]any{"id": "changed_destination_id"}
		})

		_, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID})
		var changed *app.AuthorizationScopeChangedError
		if !errors.As(err, &changed) {
			t.Fatalf("RunReverseETL(changed scope) error = %T %v, want AuthorizationScopeChangedError", err, err)
		}
		if changed.Property != "field_mappings" {
			t.Fatalf("changed property = %q, want field_mappings", changed.Property)
		}
		if got := *writes; got != before {
			t.Fatalf("provider writes after changed-scope refusal = %d, want %d", got, before)
		}
	})

	t.Run("revocation", func(t *testing.T) {
		a, plan, writes := setupAuthorizedReversePlan(t, ctx)
		before := *writes
		reference := authorizationReferenceForPlan(t, a, plan.ID)
		if err := a.RevokeAuthorization(reference); err != nil {
			t.Fatalf("RevokeAuthorization() error = %v", err)
		}

		_, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID})
		var revoked *app.AuthorizationRevokedError
		if !errors.As(err, &revoked) {
			t.Fatalf("RunReverseETL(revoked) error = %T %v, want AuthorizationRevokedError", err, err)
		}
		if got := *writes; got != before {
			t.Fatalf("provider writes after revocation = %d, want %d", got, before)
		}
	})

	t.Run("expiry", func(t *testing.T) {
		a, plan, writes := setupAuthorizedReversePlan(t, ctx)
		before := *writes
		reference := authorizationReferenceForPlan(t, a, plan.ID)
		mutateStoredAuthorization(t, a.ProjectDir(), reference, func(stored map[string]any) {
			scope, ok := stored["scope"].(map[string]any)
			if !ok {
				t.Fatalf("authorization scope has type %T", stored["scope"])
			}
			scope["expires_at"] = time.Now().UTC().Add(-time.Minute).Format(time.RFC3339Nano)
		})

		_, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID})
		var expired *app.AuthorizationExpiredError
		if !errors.As(err, &expired) {
			t.Fatalf("RunReverseETL(expired) error = %T %v, want AuthorizationExpiredError", err, err)
		}
		if got := *writes; got != before {
			t.Fatalf("provider writes after expiry = %d, want %d", got, before)
		}
	})
}

func TestRunReverseETLConsumesAuthorizationTokenOnceAndStoresNoMaterial(t *testing.T) {
	ctx := context.Background()
	a, plan, writes := setupAuthorizedReversePlan(t, ctx)
	if got := *writes; got != 1 {
		t.Fatalf("first proceed provider writes = %d, want 1", got)
	}
	stored := a.ListAuthorizations()
	if len(stored) != 1 {
		t.Fatalf("ListAuthorizations() = %#v, want one durable authorization", stored)
	}

	_, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken})
	var replay *app.AuthorizationTokenReplayError
	if !errors.As(err, &replay) {
		t.Fatalf("RunReverseETL(token replay) error = %T %v, want AuthorizationTokenReplayError", err, err)
	}
	if got := *writes; got != 1 {
		t.Fatalf("provider writes after token replay = %d, want 1", got)
	}
	if got := len(a.ListAuthorizations()); got != 1 {
		t.Fatalf("durable authorizations after token replay = %d, want 1", got)
	}

	stateBytes, err := os.ReadFile(filepath.Join(a.ProjectDir(), "state", "state.json"))
	if err != nil {
		t.Fatalf("ReadFile(state) error = %v", err)
	}
	serialized, err := json.Marshal(stored)
	if err != nil {
		t.Fatalf("Marshal(ListAuthorizations) error = %v", err)
	}
	for _, forbidden := range []string{plan.ApprovalToken, "sample-token"} {
		if bytes.Contains(stateBytes, []byte(forbidden)) {
			t.Fatalf("state.json contains protected material %q", forbidden)
		}
		if bytes.Contains(serialized, []byte(forbidden)) {
			t.Fatalf("authorization output contains protected material %q: %s", forbidden, serialized)
		}
	}
	for _, key := range []string{`"destination_credential":`, `"destination_config":`} {
		if bytes.Contains(serialized, []byte(key)) {
			t.Fatalf("authorization output contains raw protected field %s: %s", key, serialized)
		}
	}
	if stored[0].Reference == "" || stored[0].ScopeIdentity == "" {
		t.Fatalf("authorization record lacks safe reference/identity: %+v", stored[0])
	}
}

// setupAuthorizedReversePlan uses a loopback GitHub endpoint because it is the
// only hermetic fixture that exercises plan, preview, destructive confirmation,
// and the real provider request path. The caller observes every send via writes.
func setupAuthorizedReversePlan(t *testing.T, ctx context.Context) (*app.App, app.ReversePlan, *int) {
	t.Helper()
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writes++
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	a, plan := setupGitHubGenericDestructivePlan(t, ctx, server.URL)
	var err error
	plan, _, err = a.PreviewReversePlan(ctx, plan.ID, nil)
	if err != nil {
		t.Fatalf("PreviewReversePlan() error = %v", err)
	}
	if _, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID: plan.ID, ApprovalToken: plan.ApprovalToken,
		Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}); err != nil {
		t.Fatalf("RunReverseETL(first proceed) error = %v", err)
	}
	return a, plan, &writes
}

func authorizationReferenceForPlan(t *testing.T, a *app.App, planID string) string {
	t.Helper()
	for _, plan := range a.ListReversePlans() {
		if plan.ID == planID && plan.AuthorizationReference != "" {
			return plan.AuthorizationReference
		}
	}
	t.Fatalf("reverse plan %q has no durable authorization reference", planID)
	return ""
}

func mutateStoredAuthorization(t *testing.T, projectDir, reference string, mutate func(map[string]any)) {
	t.Helper()
	path := filepath.Join(projectDir, "state", "state.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(state) error = %v", err)
	}
	var state map[string]any
	if err := json.Unmarshal(raw, &state); err != nil {
		t.Fatalf("Unmarshal(state) error = %v", err)
	}
	authorizations, ok := state["authorizations"].([]any)
	if !ok {
		t.Fatalf("state authorizations has type %T", state["authorizations"])
	}
	for _, item := range authorizations {
		authorization, ok := item.(map[string]any)
		if !ok || authorization["reference"] != reference {
			continue
		}
		mutate(authorization)
		encoded, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			t.Fatalf("Marshal(state) error = %v", err)
		}
		if err := os.WriteFile(path, append(encoded, '\n'), 0o600); err != nil {
			t.Fatalf("WriteFile(state) error = %v", err)
		}
		return
	}
	t.Fatalf("authorization %q not found in state", reference)
}

func cloneAuthorizationScope(scope app.AuthorizationScope) app.AuthorizationScope {
	cloned := scope
	cloned.StreamTables = append([]app.AuthorizationStreamTable(nil), scope.StreamTables...)
	cloned.FieldMappings = make(map[string]string, len(scope.FieldMappings))
	for source, destination := range scope.FieldMappings {
		cloned.FieldMappings[source] = destination
	}
	cloned.EnabledOperations = append([]string(nil), scope.EnabledOperations...)
	return cloned
}
