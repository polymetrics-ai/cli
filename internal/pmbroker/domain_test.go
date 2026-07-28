package pmbroker

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRuntimeModePolicy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		selection RuntimeSelection
		wantErr   error
	}{
		{
			name: "remote production write allowed",
			selection: RuntimeSelection{
				Mode:        RuntimeModeRemote,
				Environment: EnvironmentTypeProduction,
				Operation:   RuntimeOperationWrite,
			},
		},
		{
			name: "hybrid requires policy binding",
			selection: RuntimeSelection{
				Mode:        RuntimeModeHybrid,
				Environment: EnvironmentTypeDevelopment,
				Operation:   RuntimeOperationRead,
			},
			wantErr: ErrHybridPolicyRequired,
		},
		{
			name: "non-canonical hybrid cannot bypass policy binding",
			selection: RuntimeSelection{
				Mode:        RuntimeMode("hybrid\n"),
				Environment: EnvironmentTypeDevelopment,
				Operation:   RuntimeOperationRead,
			},
			wantErr: ErrInvalidRuntimeMode,
		},
		{
			name: "hybrid development read with policy allowed",
			selection: RuntimeSelection{
				Mode:            RuntimeModeHybrid,
				Environment:     EnvironmentTypeDevelopment,
				Operation:       RuntimeOperationRead,
				PolicyBindingID: "policy_ci_fixture",
			},
		},
		{
			name: "invalid policy binding refused even outside hybrid",
			selection: RuntimeSelection{
				Mode:            RuntimeModeRemote,
				Environment:     EnvironmentTypeDevelopment,
				Operation:       RuntimeOperationRead,
				PolicyBindingID: "bad\npolicy",
			},
			wantErr: ErrInvalidPolicyBindingID,
		},
		{
			name: "local production write refused",
			selection: RuntimeSelection{
				Mode:        RuntimeModeLocal,
				Environment: EnvironmentTypeProduction,
				Operation:   RuntimeOperationWrite,
			},
			wantErr: ErrProductionLocalFallbackForbidden,
		},
		{
			name: "hybrid production schedule refused even with policy",
			selection: RuntimeSelection{
				Mode:            RuntimeModeHybrid,
				Environment:     EnvironmentTypeProduction,
				Operation:       RuntimeOperationScheduledJob,
				PolicyBindingID: "policy_ci_fixture",
			},
			wantErr: ErrProductionLocalFallbackForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.selection.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestContextStateValidationAndResolution(t *testing.T) {
	t.Parallel()

	ctx := testContext("prod")
	if err := ctx.Validate(); err != nil {
		t.Fatalf("Context.Validate() error = %v", err)
	}
	whitespaceName := testContext("prod ")
	if err := whitespaceName.Validate(); !errors.Is(err, ErrInvalidContextName) {
		t.Fatalf("Context.Validate() whitespace name error = %v, want %v", err, ErrInvalidContextName)
	}
	state := UserState{Version: CurrentStateVersion, ActiveContext: "prod", Contexts: []Context{ctx, testContext("dev")}}
	if err := state.Validate(); err != nil {
		t.Fatalf("UserState.Validate() error = %v", err)
	}

	tests := []struct {
		name    string
		req     ResolveRequest
		want    string
		source  ResolveSource
		wantErr error
	}{
		{name: "explicit wins", req: ResolveRequest{State: state, ExplicitContext: "dev", AllowLegacyLocal: true}, want: "dev", source: ResolveSourceExplicit},
		{name: "approval-bound requirement wins after explicit", req: ResolveRequest{State: state, ApprovalBoundContext: "dev"}, want: "dev", source: ResolveSourceApprovalBound},
		{name: "project required wins before active", req: ResolveRequest{State: state, ProjectRequiredContext: "dev"}, want: "dev", source: ResolveSourceProjectRequired},
		{name: "active wins before legacy", req: ResolveRequest{State: state, AllowLegacyLocal: true}, want: "prod", source: ResolveSourceActiveUser},
		{name: "legacy synthesized when no state", req: ResolveRequest{AllowLegacyLocal: true}, want: LegacyLocalContextName, source: ResolveSourceLegacyLocal},
		{name: "missing explicit stops safely", req: ResolveRequest{State: state, ExplicitContext: "missing"}, wantErr: ErrContextNotFound},
		{name: "required mismatch stops safely", req: ResolveRequest{State: state, ExplicitContext: "prod", ProjectRequiredContext: "dev"}, wantErr: ErrContextMismatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := ResolveContext(tt.req)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ResolveContext() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if resolved.Context.Name != tt.want || resolved.Source != tt.source {
				t.Fatalf("ResolveContext() = (%s,%s), want (%s,%s)", resolved.Context.Name, resolved.Source, tt.want, tt.source)
			}
		})
	}
}

func TestStoreRejectsUnknownSecretFields(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "contexts.json")
	store := Store{Path: path}
	state := UserState{Version: CurrentStateVersion, ActiveContext: "prod", Contexts: []Context{testContext("prod")}}
	if err := store.Save(state); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.ActiveContext != "prod" || len(loaded.Contexts) != 1 {
		t.Fatalf("loaded state = %#v, want prod context", loaded)
	}

	poisoned := `{"version":1,"active_context":"prod","contexts":[],"token":"not-allowed"}`
	if err := os.WriteFile(path, []byte(poisoned), 0o600); err != nil {
		t.Fatalf("write poisoned store: %v", err)
	}
	if _, err := store.Load(); !errors.Is(err, ErrUnsafeState) {
		t.Fatalf("Load() poisoned error = %v, want %v", err, ErrUnsafeState)
	}

	trailing := `{"version":1,"contexts":[]} {"version":1}`
	if err := os.WriteFile(path, []byte(trailing), 0o600); err != nil {
		t.Fatalf("write trailing store: %v", err)
	}
	if _, err := store.Load(); !errors.Is(err, ErrUnsafeState) {
		t.Fatalf("Load() trailing error = %v, want %v", err, ErrUnsafeState)
	}

	if err := store.Save(state); err != nil {
		t.Fatalf("restore safe state: %v", err)
	}
	if err := os.Chmod(path, 0o666); err != nil {
		t.Fatalf("chmod unsafe store: %v", err)
	}
	if _, err := store.Load(); !errors.Is(err, ErrUnsafeState) {
		t.Fatalf("Load() unsafe mode error = %v, want %v", err, ErrUnsafeState)
	}
}

func TestIncompatibleContractVersionError(t *testing.T) {
	t.Parallel()

	err := NewIncompatibleContractVersionError("2.0", "corr_test_0123456789abcdef")
	if err.HTTPStatus != 426 {
		t.Fatalf("HTTPStatus = %d, want 426", err.HTTPStatus)
	}
	if err.Error.Code != ErrorCodeIncompatibleContractVersion {
		t.Fatalf("code = %q, want %q", err.Error.Code, ErrorCodeIncompatibleContractVersion)
	}
	if len(err.Error.SupportedVersions) != 1 || err.Error.SupportedVersions[0] != ContractVersion1 {
		t.Fatalf("supported versions = %v, want [1.0]", err.Error.SupportedVersions)
	}
	if err.UnsafeRequestedVersion != "" {
		t.Fatalf("unsafe requested version leaked: %q", err.UnsafeRequestedVersion)
	}
}

func testContext(name string) Context {
	envType := EnvironmentTypeProduction
	if name != "prod" {
		envType = EnvironmentTypeDevelopment
	}
	return Context{
		Name: name,
		Organization: Organization{
			ID:          "org_0123456789abcdef",
			DisplayName: "Acme Organization",
		},
		Workspace: Workspace{
			ID:             "wks_0123456789abcdef",
			OrganizationID: "org_0123456789abcdef",
			DisplayName:    "Analytics Workspace",
		},
		Environment: Environment{
			ID:             "env_0123456789abcdef",
			WorkspaceID:    "wks_0123456789abcdef",
			OrganizationID: "org_0123456789abcdef",
			DisplayName:    name + " Environment",
			Type:           envType,
		},
		BrokerProfile: BrokerProfile{
			ID:             "bpf_0123456789abcdef",
			OrganizationID: "org_0123456789abcdef",
			WorkspaceID:    "wks_0123456789abcdef",
			EnvironmentID:  "env_0123456789abcdef",
			DisplayName:    "Pilot Broker Profile",
		},
		Runtime: RuntimeModeSelection{Mode: DefaultRuntimeMode(envType)},
	}
}
