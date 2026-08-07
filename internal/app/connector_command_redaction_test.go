package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
)

// sealedSentinel is an obviously fake stand-in for caller-sealed secret
// material. It must never look like a real credential.
const sealedSentinel = "SEALED-SENTINEL-VALUE"

// TestPlanConnectorCommandRedactsDeclaredWriteActionFields pins that a
// connector-command reverse-ETL plan honours the resolved write action's
// redact_fields, the way the source-table plan path already does.
//
// The plan's Sample is the operator-facing copy: it is persisted to
// state.json and rendered by `--json`. A write action declaring
// redact_fields:["encrypted_value"] must not have that value reproduced there.
//
// ConnectorCommandRecord withholds the same fields outright: the key is absent
// at rest, and the operator re-supplies it from the same command flag at
// preview/approve, where the existing plan-hash check re-binds it.
func TestPlanConnectorCommandRedactsDeclaredWriteActionFields(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	connector := &redactingPlanConnector{}
	a.Registry().Register(connector)
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "redacting-local",
		Connector: connector.Name(),
		Config:    map[string]string{"fixture": "local"},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}

	plan, _, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Name:       "set_secret",
		Connector:  connector.Name(),
		Credential: "redacting-local",
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

	if got := plan.RedactFields; len(got) != 1 || got[0] != "encrypted_value" {
		t.Fatalf("plan.RedactFields = %#v, want [encrypted_value] from the write action", got)
	}
	if len(plan.Sample) != 1 {
		t.Fatalf("plan.Sample = %#v, want exactly one sample record", plan.Sample)
	}
	// The declared field is redacted, and only it: the plan stays usable.
	if got := plan.Sample[0]["encrypted_value"]; got != "redacted" {
		t.Fatalf("sample encrypted_value = %#v, want %q", got, "redacted")
	}
	if got := plan.Sample[0]["secret_name"]; got != "DEPLOY_KEY" {
		t.Fatalf("sample secret_name = %#v, want DEPLOY_KEY (only declared fields are redacted)", got)
	}
	if got := plan.Sample[0]["key_id"]; got != "568250167242549743" {
		t.Fatalf("sample key_id = %#v, want it preserved (only declared fields are redacted)", got)
	}
	// The executable copy withholds the declared field entirely: absent, not
	// a placeholder that execute would happily dispatch.
	if _, present := plan.ConnectorCommandRecord["encrypted_value"]; present {
		t.Fatalf("ConnectorCommandRecord = %#v, want encrypted_value withheld entirely", plan.ConnectorCommandRecord)
	}
	if got := plan.ConnectorCommandRecord["secret_name"]; got != "DEPLOY_KEY" {
		t.Fatalf("ConnectorCommandRecord secret_name = %#v, want undeclared fields kept", got)
	}

	reopened, err := app.Open(root)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	stored, err := reopened.GetReversePlan(plan.ID)
	if err != nil {
		t.Fatalf("GetReversePlan: %v", err)
	}
	if len(stored.Sample) != 1 || stored.Sample[0]["encrypted_value"] != "redacted" {
		t.Fatalf("reloaded sample = %#v, want encrypted_value redacted in the persisted plan", stored.Sample)
	}

	// Assert against the bytes on disk, not just the decoded struct: an
	// in-memory-only check would not catch the sample being written raw.
	raw, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatalf("read state.json: %v", err)
	}
	var state struct {
		ReversePlans []struct {
			ID     string              `json:"id"`
			Sample []connectors.Record `json:"sample"`
		} `json:"reverse_plans"`
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		t.Fatalf("decode state.json: %v", err)
	}
	found := false
	for _, p := range state.ReversePlans {
		if p.ID != plan.ID {
			continue
		}
		found = true
		sample, err := json.Marshal(p.Sample)
		if err != nil {
			t.Fatalf("marshal persisted sample: %v", err)
		}
		if strings.Contains(string(sample), sealedSentinel) {
			t.Fatalf("persisted sample bytes contain the sealed sentinel: %s", sample)
		}
	}
	if !found {
		t.Fatalf("plan %s not found in persisted state", plan.ID)
	}

	// What the CLI emits for --json: safeReversePlanForOutput nils the
	// executable record and re-runs this same redaction over the sample.
	output := plan
	output.ConnectorCommandRecord = nil
	output.Sample = app.RedactReversePlanRecords(output.Sample, output.RedactFields)
	emitted, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("marshal plan for output: %v", err)
	}
	if strings.Contains(string(emitted), sealedSentinel) {
		t.Fatalf("emitted plan JSON contains the sealed sentinel: %s", emitted)
	}
	if !strings.Contains(string(emitted), "DEPLOY_KEY") || !strings.Contains(string(emitted), "568250167242549743") {
		t.Fatalf("emitted plan JSON dropped non-sensitive fields: %s", emitted)
	}
}

type redactingPlanConnector struct{}

func (c *redactingPlanConnector) Name() string { return "redacting-secret-fixture" }

func (c *redactingPlanConnector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:         c.Name(),
		DisplayName:  "Redacting secret fixture",
		Capabilities: connectors.Capabilities{Write: true},
	}
}

func (c *redactingPlanConnector) Manifest() connectors.Manifest {
	return connectors.Manifest{
		Metadata: c.Metadata(),
		WriteActions: []connectors.WriteActionSpec{{
			Name:         "set_secret",
			Method:       http.MethodPut,
			Path:         "/secrets/{secret_name}",
			RedactFields: []string{"encrypted_value"},
		}},
	}
}

func (c *redactingPlanConnector) CommandSurface() *connectors.CommandSurface {
	return &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
		Path:         "secret set",
		Intent:       "reverse_etl",
		Availability: "implemented",
		Write:        "set_secret",
		Flags: []connectors.CommandSurfaceFlag{
			{Name: "secret-name", Type: "string", MapsTo: "record.secret_name", Required: true},
			{Name: "encrypted-value", Type: "string", MapsTo: "record.encrypted_value", Required: true},
			{Name: "key-id", Type: "string", MapsTo: "record.key_id", Required: true},
		},
	}}}
}

func (*redactingPlanConnector) Check(context.Context, connectors.RuntimeConfig) error { return nil }

func (*redactingPlanConnector) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{}, connectors.ErrUnsupportedOperation
}

func (*redactingPlanConnector) Read(context.Context, connectors.ReadRequest, func(connectors.Record) error) error {
	return connectors.ErrUnsupportedOperation
}

func (c *redactingPlanConnector) DryRunWrite(_ context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	return connectors.WritePreview{
		RecordsStaged: len(records),
		Action:        req.Action,
		Digest:        strings.Repeat("e", 64),
	}, nil
}

func (c *redactingPlanConnector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{RecordsWritten: 1}, nil
}
