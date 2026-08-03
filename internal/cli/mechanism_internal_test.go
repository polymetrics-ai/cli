package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	browserauthstore "polymetrics.ai/internal/browserauth/store"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/vault"
)

func TestMechanismMarker(t *testing.T) {
	cases := []struct {
		name string
		mech *connectors.MechanismSpec
		want string
	}{
		{"nil mechanism (legacy/local primitive)", nil, ""},
		{"sanctioned official_api", &connectors.MechanismSpec{Kind: connectors.MechanismOfficialAPI, SanctionedByProvider: true}, ""},
		{"unsanctioned web_session", &connectors.MechanismSpec{Kind: connectors.MechanismWebSession, SanctionedByProvider: false}, " [UNOFFICIAL]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := mechanismMarker(tc.mech); got != tc.want {
				t.Fatalf("mechanismMarker(%+v) = %q, want %q", tc.mech, got, tc.want)
			}
		})
	}
}

// fakeWebConnector is a minimal connectors.Connector + DefinitionProvider
// standing in for a future -web connector, so the enable-gate can be tested
// end-to-end without this foundations-only PR shipping an actual connector.
type fakeWebConnector struct {
	name string
	def  connectors.Definition
}

func (f fakeWebConnector) Name() string { return f.name }
func (f fakeWebConnector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: f.name, DisplayName: f.def.DisplayName, Mechanism: f.def.Mechanism}
}
func (f fakeWebConnector) Check(context.Context, connectors.RuntimeConfig) error {
	return nil
}
func (f fakeWebConnector) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{Connector: f.name}, nil
}
func (f fakeWebConnector) Read(context.Context, connectors.ReadRequest, func(connectors.Record) error) error {
	return connectors.ErrUnsupportedOperation
}
func (f fakeWebConnector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
func (f fakeWebConnector) Definition() connectors.Definition { return f.def }

func fakeWebSessionDefinition(name string) connectors.Definition {
	return connectors.Definition{
		Name:        name,
		DisplayName: "Acme Web",
		Mechanism: &connectors.MechanismSpec{
			Kind:                 connectors.MechanismWebSession,
			Label:                "Acme web session (unofficial, experimental)",
			SanctionedByProvider: false,
			ProviderTermsURL:     "https://acme.example/terms",
			OptInRequired:        true,
		},
		Risk: connectors.RiskSpec{Approval: "Using this connector may get your personal Acme account suspended or shut down."},
	}
}

func TestConnectorEnableRequiresWebSessionMechanism(t *testing.T) {
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	registry := connectors.NewEmptyRegistry()
	registry.Register(fakeWebConnector{name: "acme", def: connectors.Definition{
		Name: "acme", DisplayName: "Acme",
		Mechanism: &connectors.MechanismSpec{Kind: connectors.MechanismOfficialAPI, SanctionedByProvider: true},
	}})

	var stdout bytes.Buffer
	err := runConnectorEnable(context.Background(), root, registry, []string{"acme"}, &stdout, false)
	if err == nil || !strings.Contains(err.Error(), "does not require enabling") {
		t.Fatalf("runConnectorEnable(official_api) error = %v, want a \"does not require enabling\" complaint", err)
	}
}

func TestConnectorEnableFlow(t *testing.T) {
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	def := fakeWebSessionDefinition("acme-web")
	registry := connectors.NewEmptyRegistry()
	registry.Register(fakeWebConnector{name: "acme-web", def: def})

	ctx := context.Background()
	wantHash := browserauthstore.HashWarning(connectorRiskWarning(def))

	// Without --accept-risk: shows the warning and the hash, saves nothing.
	var stdout bytes.Buffer
	if err := runConnectorEnable(ctx, root, registry, []string{"acme-web"}, &stdout, false); err != nil {
		t.Fatalf("runConnectorEnable() error = %v", err)
	}
	out := stdout.String()
	for _, want := range []string{"UNOFFICIAL", "suspended", wantHash} {
		if !strings.Contains(out, want) {
			t.Fatalf("enable output missing %q:\n%s", want, out)
		}
	}

	v, err := vault.Open(filepath.Join(root, ".polymetrics"))
	if err != nil {
		t.Fatalf("vault.Open() error = %v", err)
	}
	store := browserauthstore.New(v)
	accepted, err := store.HasAcceptedCurrentRisk(ctx, "acme-web", "default", wantHash)
	if err != nil {
		t.Fatalf("HasAcceptedCurrentRisk() error = %v", err)
	}
	if accepted {
		t.Fatalf("risk marked accepted before --accept-risk was ever supplied")
	}

	// A mismatched hash is rejected, not silently accepted.
	stdout.Reset()
	err = runConnectorEnable(ctx, root, registry, []string{"acme-web", "--accept-risk", "0000000000000000000000000000000000000000000000000000000000000000"}, &stdout, false)
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("runConnectorEnable() with wrong hash error = %v, want a mismatch complaint", err)
	}

	// The real hash enables it.
	stdout.Reset()
	if err := runConnectorEnable(ctx, root, registry, []string{"acme-web", "--accept-risk", wantHash}, &stdout, false); err != nil {
		t.Fatalf("runConnectorEnable() with correct hash error = %v", err)
	}
	accepted, err = store.HasAcceptedCurrentRisk(ctx, "acme-web", "default", wantHash)
	if err != nil {
		t.Fatalf("HasAcceptedCurrentRisk() error = %v", err)
	}
	if !accepted {
		t.Fatalf("risk acceptance was not recorded after a matching --accept-risk")
	}
}

func TestConnectorEnableUnknownConnector(t *testing.T) {
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	registry := connectors.NewEmptyRegistry()
	var stdout bytes.Buffer
	err := runConnectorEnable(context.Background(), root, registry, []string{"does-not-exist"}, &stdout, false)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("runConnectorEnable(unknown) error = %v, want a \"not found\" complaint", err)
	}
}
