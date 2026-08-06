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
	if strings.Contains(string(encoded), "binding") || strings.Contains(string(encoded), firstRuntime.CoordinationIdentity.AuthCohortKey()) {
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
	if runtime.CredentialRevision == runtime.CoordinationIdentity.AuthCohortKey() {
		t.Fatal("approval revision was reused as coordination identity")
	}
}
