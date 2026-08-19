package app_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
)

const tokenSentinel = "BEARER-SENTINEL-VALUE"

func stateBytes(t *testing.T, root string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatalf("read state.json: %v", err)
	}
	return string(raw)
}

func withholdingApp(t *testing.T, ctx context.Context, connector connectors.Connector) (*app.App, string) {
	t.Helper()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	a.Registry().Register(connector)
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "withholding-local",
		Connector: connector.Name(),
		Config:    map[string]string{"fixture": "local"},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}
	return a, root
}

// TestWithheldFieldNeverReachesStateJSON asserts against the raw persisted
// bytes, not a sub-slice of the in-memory plan. The sample-only assertion in
// the previous round is exactly how the at-rest half survived undetected.
func TestWithheldFieldNeverReachesStateJSON(t *testing.T) {
	ctx := context.Background()
	a, root := withholdingApp(t, ctx, &withholdingConnector{})

	plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Name:       "set_secret",
		Connector:  "withholding-fixture",
		Credential: "withholding-local",
		Path:       []string{"secret", "set"},
		Flags: map[string][]string{
			"secret-name":     {"DEPLOY_KEY"},
			"encrypted-value": {sealedSentinel},
			"key-id":          {"568250167242549743"},
		},
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand: %v", err)
	}

	if raw := stateBytes(t, root); strings.Contains(raw, sealedSentinel) {
		t.Fatalf("state.json contains the withheld sealed value")
	}
	if _, present := plan.ConnectorCommandRecord["encrypted_value"]; present {
		t.Fatalf("ConnectorCommandRecord = %#v, want encrypted_value absent", plan.ConnectorCommandRecord)
	}
	if raw := stateBytes(t, root); !strings.Contains(raw, "DEPLOY_KEY") {
		t.Fatalf("state.json lost the non-sensitive fields the plan needs")
	}
}

// TestWithheldRequiredFieldReplaysEndToEnd is the OAuth shape: the withheld
// field is a REQUIRED body field carrying a live bearer token. The four GitHub
// OAuth endpoints are still disallowed on this branch, so this is a fixture.
func TestWithheldRequiredFieldReplaysEndToEnd(t *testing.T) {
	ctx := context.Background()
	connector := &withholdingConnector{}
	a, root := withholdingApp(t, ctx, connector)

	plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Name:       "revoke_token",
		Connector:  "withholding-fixture",
		Credential: "withholding-local",
		Path:       []string{"oauth", "revoke"},
		Flags: map[string][]string{
			"client-id":    {"Iv1.fixture"},
			"access-token": {tokenSentinel},
		},
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand: %v", err)
	}
	if raw := stateBytes(t, root); strings.Contains(raw, tokenSentinel) {
		t.Fatalf("state.json contains the withheld bearer token")
	}

	// Without re-supply the plan will not preview, and the error names the
	// field and the flag that supplies it.
	_, _, err = a.PreviewConnectorCommandPlan(ctx, plan.ID, nil)
	if err == nil {
		t.Fatal("PreviewConnectorCommandPlan with no re-supply = nil error, want a withheld-field error")
	}
	if !strings.Contains(err.Error(), "access_token") || !strings.Contains(err.Error(), "--access-token") {
		t.Fatalf("error = %v, want it to name the withheld field and its flag", err)
	}

	resupply := map[string][]string{"client-id": {"Iv1.fixture"}, "access-token": {tokenSentinel}}
	_, preview, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, resupply)
	if err != nil {
		t.Fatalf("PreviewConnectorCommandPlan with re-supply: %v", err)
	}
	if preview.Digest == "" {
		t.Fatalf("preview = %#v, want a digest", preview)
	}
	if raw := stateBytes(t, root); strings.Contains(raw, tokenSentinel) {
		t.Fatalf("previewing re-persisted the withheld bearer token to state.json")
	}

	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		WithheldFlags: resupply,
	})
	if err != nil {
		t.Fatalf("RunReverseETL with re-supply: %v", err)
	}
	if run.RecordsSucceeded != 1 {
		t.Fatalf("run = %#v, want the reconstituted record dispatched", run)
	}
	if got := connector.lastWritten["access_token"]; got != tokenSentinel {
		t.Fatalf("dispatched access_token = %#v, want the re-supplied value", got)
	}
	if raw := stateBytes(t, root); strings.Contains(raw, tokenSentinel) {
		t.Fatalf("executing persisted the withheld bearer token to state.json")
	}
}

// TestWrongResuppliedValueFailsHashCheck is the negative control that the
// re-supply path is bound, not merely merged.
func TestWrongResuppliedValueFailsHashCheck(t *testing.T) {
	ctx := context.Background()
	connector := &withholdingConnector{}
	a, _ := withholdingApp(t, ctx, connector)

	plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Name:       "revoke_token",
		Connector:  "withholding-fixture",
		Credential: "withholding-local",
		Path:       []string{"oauth", "revoke"},
		Flags: map[string][]string{
			"client-id":    {"Iv1.fixture"},
			"access-token": {tokenSentinel},
		},
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand: %v", err)
	}
	wrong := map[string][]string{"client-id": {"Iv1.fixture"}, "access-token": {"WRONG-SENTINEL-VALUE"}}
	if _, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, wrong); err == nil {
		t.Fatal("preview with a wrong re-supplied value = nil error, want the plan-hash check to fail")
	}
	if connector.writes != 0 {
		t.Fatalf("connector writes = %d, want no dispatch on a failed hash check", connector.writes)
	}
}

// TestDeclaredButUnsuppliedFieldIsNotOwedBack pins the difference between
// "withheld" and "never supplied". A declared sensitive field the operator did
// not pass was never in the record and was never removed, so it must not become
// a re-supply precondition: the plan hash was computed without it, and no later
// value could ever satisfy the plan.
func TestDeclaredButUnsuppliedFieldIsNotOwedBack(t *testing.T) {
	ctx := context.Background()
	connector := &withholdingConnector{}
	a, _ := withholdingApp(t, ctx, connector)

	for _, tc := range []struct {
		name string
		path []string
	}{
		// The declared field has an optional flag the operator omitted.
		{name: "optional_flag_omitted", path: []string{"hook", "add"}},
		// The declared field has no flag on the command at all.
		{name: "no_flag_maps_to_the_field", path: []string{"hook", "ping"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
				Name:       "hook",
				Connector:  "withholding-fixture",
				Credential: "withholding-local",
				Path:       tc.path,
				Flags:      map[string][]string{"hook-name": {"deploy"}},
			})
			if err != nil {
				t.Fatalf("PlanConnectorCommand: %v", err)
			}
			if len(plan.RedactFields) == 0 {
				t.Fatalf("plan.RedactFields = %#v, want the declared sensitive field", plan.RedactFields)
			}
			if len(plan.WithheldFields) != 0 {
				t.Fatalf("plan.WithheldFields = %#v, want nothing withheld when nothing was supplied", plan.WithheldFields)
			}
			if _, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, nil); err != nil {
				t.Fatalf("PreviewConnectorCommandPlan with no re-supply: %v", err)
			}
			run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken})
			if err != nil {
				t.Fatalf("RunReverseETL with no re-supply: %v", err)
			}
			if run.RecordsSucceeded != 1 {
				t.Fatalf("run = %#v, want a normal dispatch", run)
			}
		})
	}
}

// TestUndeclaredFieldsRoundTripUnchanged is the negative control that the
// mechanism only withholds what a write action declares.
func TestUndeclaredFieldsRoundTripUnchanged(t *testing.T) {
	ctx := context.Background()
	connector := &withholdingConnector{}
	a, root := withholdingApp(t, ctx, connector)

	plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Name:       "add_label",
		Connector:  "withholding-fixture",
		Credential: "withholding-local",
		Path:       []string{"label", "add"},
		Flags:      map[string][]string{"label-name": {"bug"}},
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand: %v", err)
	}
	if len(plan.RedactFields) != 0 {
		t.Fatalf("plan.RedactFields = %#v, want none declared", plan.RedactFields)
	}
	if got := plan.ConnectorCommandRecord["label_name"]; got != "bug" {
		t.Fatalf("ConnectorCommandRecord label_name = %#v, want it kept", got)
	}
	if raw := stateBytes(t, root); !strings.Contains(raw, "bug") {
		t.Fatalf("state.json dropped an undeclared field")
	}
	// It still executes with no re-supply at all.
	if _, _, err := a.PreviewConnectorCommandPlan(ctx, plan.ID, nil); err != nil {
		t.Fatalf("PreviewConnectorCommandPlan: %v", err)
	}
	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken})
	if err != nil {
		t.Fatalf("RunReverseETL: %v", err)
	}
	if run.RecordsSucceeded != 1 {
		t.Fatalf("run = %#v, want a normal dispatch", run)
	}
}

type withholdingConnector struct {
	writes      int
	lastWritten connectors.Record
}

func (c *withholdingConnector) Name() string { return "withholding-fixture" }

func (c *withholdingConnector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:         c.Name(),
		DisplayName:  "Withholding fixture",
		Capabilities: connectors.Capabilities{Write: true},
	}
}

func (c *withholdingConnector) Manifest() connectors.Manifest {
	return connectors.Manifest{
		Metadata: c.Metadata(),
		WriteActions: []connectors.WriteActionSpec{
			{
				Name:         "set_secret",
				Method:       http.MethodPut,
				Path:         "/secrets/{secret_name}",
				RedactFields: []string{"encrypted_value"},
			},
			{
				Name:           "revoke_token",
				Method:         http.MethodPatch,
				Path:           "/applications/{client_id}/token",
				RequiredFields: []string{"client_id", "access_token"},
				RedactFields:   []string{"access_token"},
			},
			{
				Name:   "add_label",
				Method: http.MethodPost,
				Path:   "/labels",
			},
			{
				Name:         "add_hook",
				Method:       http.MethodPost,
				Path:         "/hooks",
				RedactFields: []string{"url"},
			},
			{
				Name:         "ping_hook",
				Method:       http.MethodPost,
				Path:         "/hooks/ping",
				RedactFields: []string{"url"},
			},
		},
	}
}

func (c *withholdingConnector) CommandSurface() *connectors.CommandSurface {
	return &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{
		{
			Path: "secret set", Intent: "reverse_etl", Availability: "implemented", Write: "set_secret",
			Flags: []connectors.CommandSurfaceFlag{
				{Name: "secret-name", Type: "string", MapsTo: "record.secret_name", Required: true},
				{Name: "encrypted-value", Type: "string", MapsTo: "record.encrypted_value", Required: true},
				{Name: "key-id", Type: "string", MapsTo: "record.key_id", Required: true},
			},
		},
		{
			Path: "oauth revoke", Intent: "reverse_etl", Availability: "implemented", Write: "revoke_token",
			Flags: []connectors.CommandSurfaceFlag{
				{Name: "client-id", Type: "string", MapsTo: "record.client_id", Required: true},
				{Name: "access-token", Type: "string", MapsTo: "record.access_token", Required: true},
			},
		},
		{
			Path: "label add", Intent: "reverse_etl", Availability: "implemented", Write: "add_label",
			Flags: []connectors.CommandSurfaceFlag{
				{Name: "label-name", Type: "string", MapsTo: "record.label_name", Required: true},
			},
		},
		{
			Path: "hook add", Intent: "reverse_etl", Availability: "implemented", Write: "add_hook",
			Flags: []connectors.CommandSurfaceFlag{
				{Name: "hook-name", Type: "string", MapsTo: "record.hook_name", Required: true},
				{Name: "url", Type: "string", MapsTo: "record.url"},
			},
		},
		{
			Path: "hook ping", Intent: "reverse_etl", Availability: "implemented", Write: "ping_hook",
			Flags: []connectors.CommandSurfaceFlag{
				{Name: "hook-name", Type: "string", MapsTo: "record.hook_name", Required: true},
			},
		},
	}}
}

func (*withholdingConnector) Check(context.Context, connectors.RuntimeConfig) error { return nil }

func (*withholdingConnector) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{}, connectors.ErrUnsupportedOperation
}

func (*withholdingConnector) Read(context.Context, connectors.ReadRequest, func(connectors.Record) error) error {
	return connectors.ErrUnsupportedOperation
}

func (c *withholdingConnector) DryRunWrite(_ context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	return connectors.WritePreview{
		RecordsStaged: len(records),
		Action:        req.Action,
		Digest:        strings.Repeat("f", 64),
	}, nil
}

func (c *withholdingConnector) Write(_ context.Context, _ connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	c.writes++
	if len(records) > 0 {
		c.lastWritten = records[0]
	}
	return connectors.WriteResult{RecordsWritten: len(records)}, nil
}
