package app_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
)

func TestCredentialCoordination_ExplicitLinkSharesOnlyOpaqueIdentity(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	instance, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	first, err := instance.AddCredential(ctx, app.AddCredentialRequest{
		Name:           "sample-shared",
		Connector:      "sample",
		ProviderFamily: "provider-fixture",
		AuthProfile:    "service-profile",
	})
	if err != nil {
		t.Fatalf("AddCredential(first) error = %v", err)
	}
	if first.ProviderFamily != "provider-fixture" || first.AuthProfile != "service-profile" {
		t.Fatal("credential did not retain declared coordination metadata")
	}
	if _, err := instance.AddCredential(ctx, app.AddCredentialRequest{
		Name:           "faker-shared",
		Connector:      "faker",
		ProviderFamily: "provider-fixture",
		AuthProfile:    "service-profile",
	}); err != nil {
		t.Fatalf("AddCredential(second) error = %v", err)
	}
	if _, err := instance.LinkCredential("faker-shared", "sample-shared"); err != nil {
		t.Fatalf("LinkCredential() error = %v", err)
	}

	_, firstRuntime, err := instance.ResolveConnectorCredential(ctx, "sample", "sample-shared", nil)
	if err != nil {
		t.Fatalf("ResolveConnectorCredential(first) error = %v", err)
	}
	_, linkedRuntime, err := instance.ResolveConnectorCredential(ctx, "faker", "faker-shared", nil)
	if err != nil {
		t.Fatalf("ResolveConnectorCredential(linked) error = %v", err)
	}
	if firstRuntime.CoordinationIdentity.AuthCohortKey() == "" {
		t.Fatal("resolved credential lacks an opaque auth cohort key")
	}
	if firstRuntime.CoordinationIdentity.AuthCohortKey() != linkedRuntime.CoordinationIdentity.AuthCohortKey() {
		t.Fatal("explicitly linked compatible credentials do not share auth cohort identity")
	}

	scope := connectors.RateLimitScope{
		PolicyID: "core-rest-v1",
		Kind:     connectors.RateScopeKindAccount,
		Subject:  "account-fixture-001",
	}
	firstRate, err := firstRuntime.CoordinationIdentity.RateScopeKey(scope)
	if err != nil {
		t.Fatalf("first RateScopeKey() error = %v", err)
	}
	linkedRate, err := linkedRuntime.CoordinationIdentity.RateScopeKey(scope)
	if err != nil {
		t.Fatalf("linked RateScopeKey() error = %v", err)
	}
	if firstRate != linkedRate {
		t.Fatal("linked compatible credentials do not share a compatible rate scope")
	}
	differentSubject, err := linkedRuntime.CoordinationIdentity.RateScopeKey(connectors.RateLimitScope{
		PolicyID: "core-rest-v1",
		Kind:     connectors.RateScopeKindAccount,
		Subject:  "account-fixture-002",
	})
	if err != nil {
		t.Fatalf("RateScopeKey(different subject) error = %v", err)
	}
	if linkedRate == differentSubject {
		t.Fatal("different declared rate subject shared a budget")
	}
	if _, err := instance.AddCredential(ctx, app.AddCredentialRequest{
		Name:           "faker-linked-on-create",
		Connector:      "faker",
		ProviderFamily: "provider-fixture",
		AuthProfile:    "service-profile",
		LinkCredential: "sample-shared",
	}); err != nil {
		t.Fatalf("AddCredential(linked on create) error = %v", err)
	}
	_, createdLinkedRuntime, err := instance.ResolveConnectorCredential(ctx, "faker", "faker-linked-on-create", nil)
	if err != nil {
		t.Fatalf("ResolveConnectorCredential(linked on create) error = %v", err)
	}
	if createdLinkedRuntime.CoordinationIdentity.AuthCohortKey() != firstRuntime.CoordinationIdentity.AuthCohortKey() {
		t.Fatal("credential explicitly linked during creation does not share the auth cohort")
	}

	if _, err := instance.AddCredential(ctx, app.AddCredentialRequest{
		Name:           "faker-unlinked",
		Connector:      "faker",
		ProviderFamily: "provider-fixture",
		AuthProfile:    "service-profile",
	}); err != nil {
		t.Fatalf("AddCredential(unlinked) error = %v", err)
	}
	_, unlinkedRuntime, err := instance.ResolveConnectorCredential(ctx, "faker", "faker-unlinked", nil)
	if err != nil {
		t.Fatalf("ResolveConnectorCredential(unlinked) error = %v", err)
	}
	if linkedRuntime.CoordinationIdentity.AuthCohortKey() == unlinkedRuntime.CoordinationIdentity.AuthCohortKey() {
		t.Fatal("copied but unlinked credentials unexpectedly share a cohort")
	}

	if _, err := instance.AddCredential(ctx, app.AddCredentialRequest{
		Name:           "faker-incompatible",
		Connector:      "faker",
		ProviderFamily: "other-provider",
		AuthProfile:    "service-profile",
	}); err != nil {
		t.Fatalf("AddCredential(incompatible) error = %v", err)
	}
	if _, err := instance.LinkCredential("faker-incompatible", "sample-shared"); err == nil {
		t.Fatal("LinkCredential() accepted incompatible provider family")
	} else {
		if !strings.Contains(err.Error(), "provider family") {
			t.Fatalf("LinkCredential() error does not identify the failed constraint: %v", err)
		}
		if strings.Contains(err.Error(), "other-provider") || strings.Contains(err.Error(), "provider-fixture") {
			t.Fatal("LinkCredential() error echoed a declared metadata value")
		}
	}

	inspected, err := instance.InspectCredential("sample-shared")
	if err != nil {
		t.Fatalf("InspectCredential() error = %v", err)
	}
	encoded, err := json.Marshal(inspected)
	if err != nil {
		t.Fatalf("marshal inspected credential: %v", err)
	}
	stateBytes, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatalf("read protected coordination state: %v", err)
	}
	var protectedState struct {
		CredentialBindings map[string]struct {
			BindingID string `json:"binding_id"`
		} `json:"credential_bindings"`
	}
	if err := json.Unmarshal(stateBytes, &protectedState); err != nil {
		t.Fatalf("decode protected coordination state: %v", err)
	}
	bindingID := protectedState.CredentialBindings[first.ID].BindingID
	if bindingID == "" {
		t.Fatal("test setup did not persist a protected binding")
	}
	if strings.Contains(string(encoded), bindingID) || strings.Contains(string(encoded), string(firstRuntime.CoordinationIdentity.AuthCohortKey())) {
		t.Fatal("ordinary credential inspection exposed protected coordination identity material")
	}
}

func TestCredentialCoordination_MigratesLegacyMetadataWithoutChangingApprovalLifetime(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	instance, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, err := instance.AddCredential(ctx, app.AddCredentialRequest{Name: "legacy-sample", Connector: "sample"}); err != nil {
		t.Fatalf("AddCredential() error = %v", err)
	}

	statePath := filepath.Join(root, ".polymetrics", "state", "state.json")
	stateBytes, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	var legacy map[string]any
	if err := json.Unmarshal(stateBytes, &legacy); err != nil {
		t.Fatalf("decode state: %v", err)
	}
	delete(legacy, "credential_bindings")
	delete(legacy, "coordination_salt")
	credentials, ok := legacy["credentials"].([]any)
	if !ok || len(credentials) != 1 {
		t.Fatal("test setup did not contain one credential")
	}
	credential, ok := credentials[0].(map[string]any)
	if !ok {
		t.Fatal("test setup credential has unexpected representation")
	}
	delete(credential, "provider_family")
	delete(credential, "auth_profile")
	legacyBytes, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("encode legacy state: %v", err)
	}
	if err := os.WriteFile(statePath, legacyBytes, 0o600); err != nil {
		t.Fatalf("write legacy state: %v", err)
	}

	migrated, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open(legacy state) error = %v", err)
	}
	meta, err := migrated.InspectCredential("legacy-sample")
	if err != nil {
		t.Fatalf("InspectCredential() error = %v", err)
	}
	if meta.ProviderFamily != "sample" || meta.AuthProfile == "" {
		t.Fatal("legacy credential did not receive isolated default coordination metadata")
	}
	_, runtime, err := migrated.ResolveConnectorCredential(ctx, "sample", "legacy-sample", nil)
	if err != nil {
		t.Fatalf("ResolveConnectorCredential() error = %v", err)
	}
	if runtime.CoordinationIdentity.AuthCohortKey() == "" {
		t.Fatal("migrated credential has no opaque coordination identity")
	}
	if runtime.CredentialRevision == string(runtime.CoordinationIdentity.AuthCohortKey()) {
		t.Fatal("approval revision was reused as coordination identity")
	}
}

func TestCredentialCoordination_EmptyProjectOpenDoesNotRewriteState(t *testing.T) {
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	statePath := filepath.Join(root, ".polymetrics", "state", "state.json")
	readState := func() struct {
		Revision           uint64                     `json:"revision"`
		CredentialBindings map[string]json.RawMessage `json:"credential_bindings"`
	} {
		t.Helper()
		stateBytes, err := os.ReadFile(statePath)
		if err != nil {
			t.Fatalf("read state: %v", err)
		}
		var persisted struct {
			Revision           uint64                     `json:"revision"`
			CredentialBindings map[string]json.RawMessage `json:"credential_bindings"`
		}
		if err := json.Unmarshal(stateBytes, &persisted); err != nil {
			t.Fatalf("decode state: %v", err)
		}
		return persisted
	}

	initial := readState()
	if initial.CredentialBindings == nil {
		t.Fatal("initial state did not serialize the empty credential binding map")
	}
	for attempt := 0; attempt < 2; attempt++ {
		if _, err := app.Open(root); err != nil {
			t.Fatalf("Open() error = %v", err)
		}
		if got := readState().Revision; got != initial.Revision {
			t.Fatalf("Open() rewrote empty project state revision = %d, want %d", got, initial.Revision)
		}
	}
}

func TestCredentialCoordination_RejectsInvalidDeclarationsBeforePersistence(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	instance, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	for _, test := range []struct {
		name      string
		request   app.AddCredentialRequest
		fieldName string
		input     string
	}{
		{
			name: "provider family",
			request: app.AddCredentialRequest{
				Name:           "sample-invalid-family",
				Connector:      "sample",
				ProviderFamily: "invalid family",
			},
			fieldName: "provider family",
			input:     "invalid family",
		},
		{
			name: "auth profile",
			request: app.AddCredentialRequest{
				Name:        "sample-invalid-profile",
				Connector:   "sample",
				AuthProfile: "invalid profile",
			},
			fieldName: "auth profile",
			input:     "invalid profile",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := instance.AddCredential(ctx, test.request); err == nil {
				t.Fatal("AddCredential() accepted invalid coordination declaration")
			} else {
				if !strings.Contains(err.Error(), test.fieldName) || !strings.Contains(err.Error(), "constraint") {
					t.Fatalf("AddCredential() error did not name the field and constraint: %v", err)
				}
				if strings.Contains(err.Error(), test.input) {
					t.Fatal("AddCredential() error echoed the rejected declaration")
				}
			}
		})
	}
	if credentials := instance.ListCredentials(); len(credentials) != 0 {
		t.Fatalf("invalid declarations persisted credentials: %v", credentials)
	}
}

func TestCredentialCoordination_CrossConnectorLinksRequireExplicitDeclarations(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	instance, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if _, err := instance.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "sample-undeclared",
		Connector: "sample",
	}); err != nil {
		t.Fatalf("AddCredential(undeclared) error = %v", err)
	}
	if _, err := instance.AddCredential(ctx, app.AddCredentialRequest{
		Name:           "sample-declared",
		Connector:      "sample",
		ProviderFamily: "sample",
		AuthProfile:    "default",
	}); err != nil {
		t.Fatalf("AddCredential(declared) error = %v", err)
	}
	if _, err := instance.LinkCredential("sample-undeclared", "sample-declared"); err != nil {
		t.Fatalf("LinkCredential(same connector) error = %v", err)
	}
	if _, err := instance.AddCredential(ctx, app.AddCredentialRequest{
		Name:           "faker-explicit",
		Connector:      "faker",
		ProviderFamily: "sample",
		AuthProfile:    "default",
	}); err != nil {
		t.Fatalf("AddCredential(cross connector) error = %v", err)
	}

	requireExplicitDeclarations := func(err error) {
		t.Helper()
		if err == nil {
			t.Fatal("cross-connector link accepted an undeclared credential")
		}
		if !strings.Contains(err.Error(), "explicitly declared") {
			t.Fatalf("cross-connector link error = %v, want explicit declaration failure", err)
		}
	}
	_, err = instance.LinkCredential("faker-explicit", "sample-undeclared")
	requireExplicitDeclarations(err)
	_, err = instance.LinkCredential("sample-undeclared", "faker-explicit")
	requireExplicitDeclarations(err)
	_, err = instance.LinkCredential("faker-explicit", "sample-declared")
	requireExplicitDeclarations(err)
	_, err = instance.AddCredential(ctx, app.AddCredentialRequest{
		Name:           "faker-linked-on-create",
		Connector:      "faker",
		ProviderFamily: "sample",
		AuthProfile:    "default",
		LinkCredential: "sample-undeclared",
	})
	requireExplicitDeclarations(err)
	_, err = instance.AddCredential(ctx, app.AddCredentialRequest{
		Name:           "faker-linked-to-declared-on-create",
		Connector:      "faker",
		ProviderFamily: "sample",
		AuthProfile:    "default",
		LinkCredential: "sample-declared",
	})
	requireExplicitDeclarations(err)

	_, undeclaredRuntime, err := instance.ResolveConnectorCredential(ctx, "sample", "sample-undeclared", nil)
	if err != nil {
		t.Fatalf("ResolveConnectorCredential(undeclared) error = %v", err)
	}
	_, explicitRuntime, err := instance.ResolveConnectorCredential(ctx, "faker", "faker-explicit", nil)
	if err != nil {
		t.Fatalf("ResolveConnectorCredential(explicit) error = %v", err)
	}
	if undeclaredRuntime.CoordinationIdentity.AuthCohortKey() == explicitRuntime.CoordinationIdentity.AuthCohortKey() {
		t.Fatal("undeclared credential shared a cross-connector auth cohort")
	}
}

func TestCredentialCoordination_MigrationIsolatesUnverifiedCrossConnectorBindings(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	instance, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	for _, request := range []app.AddCredentialRequest{
		{
			Name:           "sample-shared",
			Connector:      "sample",
			ProviderFamily: "provider-fixture",
			AuthProfile:    "service-profile",
		},
		{
			Name:           "faker-shared",
			Connector:      "faker",
			ProviderFamily: "provider-fixture",
			AuthProfile:    "service-profile",
		},
	} {
		if _, err := instance.AddCredential(ctx, request); err != nil {
			t.Fatalf("AddCredential(%s) error = %v", request.Name, err)
		}
	}
	if _, err := instance.LinkCredential("faker-shared", "sample-shared"); err != nil {
		t.Fatalf("LinkCredential() error = %v", err)
	}

	statePath := filepath.Join(root, ".polymetrics", "state", "state.json")
	stateBytes, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	var persisted map[string]any
	if err := json.Unmarshal(stateBytes, &persisted); err != nil {
		t.Fatalf("decode state: %v", err)
	}
	bindings, ok := persisted["credential_bindings"].(map[string]any)
	if !ok {
		t.Fatal("test setup did not persist credential bindings")
	}
	for _, rawBinding := range bindings {
		binding, ok := rawBinding.(map[string]any)
		if !ok {
			t.Fatal("test setup binding has unexpected representation")
		}
		delete(binding, "provider_family_declared")
		delete(binding, "auth_profile_declared")
		delete(binding, "declaration_provenance_recorded")
	}
	legacyBytes, err := json.Marshal(persisted)
	if err != nil {
		t.Fatalf("encode unverified state: %v", err)
	}
	if err := os.WriteFile(statePath, legacyBytes, 0o600); err != nil {
		t.Fatalf("write unverified state: %v", err)
	}

	migrated, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open(unverified state) error = %v", err)
	}
	_, sampleRuntime, err := migrated.ResolveConnectorCredential(ctx, "sample", "sample-shared", nil)
	if err != nil {
		t.Fatalf("ResolveConnectorCredential(sample) error = %v", err)
	}
	_, fakerRuntime, err := migrated.ResolveConnectorCredential(ctx, "faker", "faker-shared", nil)
	if err != nil {
		t.Fatalf("ResolveConnectorCredential(faker) error = %v", err)
	}
	if sampleRuntime.CoordinationIdentity.AuthCohortKey() == fakerRuntime.CoordinationIdentity.AuthCohortKey() {
		t.Fatal("migration retained an unverified cross-connector auth cohort")
	}
	if _, err := migrated.LinkCredential("faker-shared", "sample-shared"); err == nil {
		t.Fatal("migration accepted a cross-connector link without declaration provenance")
	}
}
