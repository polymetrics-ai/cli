package app_test

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
)

func TestPlanConnectorCommandPersistsCompleteDeclaredContent(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	connector := &runnerContentPlanConnector{}
	a.Registry().Register(connector)
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "runner-content-local",
		Connector: connector.Name(),
		Config:    map[string]string{"fixture": "local"},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}

	plan, preview, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Name:       "delete_content",
		Connector:  connector.Name(),
		Credential: "runner-content-local",
		Path:       []string{"content", "delete"},
		Flags: map[string][]string{
			"token":          {"fixture-token"},
			"content":        {"complete connector content"},
			"nested-content": {"complete nested request"},
			"nested-token":   {"complete nested response token"},
		},
		Preview: true,
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand: %v", err)
	}
	if preview == nil || preview.Digest == "" {
		t.Fatalf("preview = %#v, want no-network preview with digest", preview)
	}
	if connector.dryRuns != 1 || connector.writes != 0 {
		t.Fatalf("connector calls dry-runs/writes = %d/%d, want preview only", connector.dryRuns, connector.writes)
	}
	if plan.Status != "previewed" || plan.PreviewDigest == "" || plan.ApprovalToken == "" || plan.PlanSeal == nil || !plan.ExpiresAt.After(plan.CreatedAt) {
		t.Fatalf("plan lifecycle = %#v, want previewed bounded destructive approval", plan)
	}

	want := connectors.Record{
		"token":   "fixture-token",
		"content": "complete connector content",
		"metadata": map[string]any{
			"request":  map[string]any{"content": "complete nested request"},
			"response": map[string]any{"token": "complete nested response token"},
		},
	}
	if !reflect.DeepEqual(plan.ConnectorCommandRecord, want) {
		t.Fatalf("plan record = %#v, want %#v", plan.ConnectorCommandRecord, want)
	}
	if len(plan.Sample) != 1 || !reflect.DeepEqual(plan.Sample[0], want) {
		t.Fatalf("plan sample = %#v, want complete declared content %#v", plan.Sample, want)
	}

	reopened, err := app.Open(root)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	stored, err := reopened.GetReversePlan(plan.ID)
	if err != nil {
		t.Fatalf("GetReversePlan: %v", err)
	}
	if stored.ApprovalToken != "" {
		t.Fatalf("stored approval token = %q, want omitted from persisted plan", stored.ApprovalToken)
	}
	if len(stored.Sample) != 1 || !reflect.DeepEqual(stored.Sample[0], want) {
		t.Fatalf("reloaded plan sample = %#v, want complete declared content %#v", stored.Sample, want)
	}
}

type runnerContentPlanConnector struct {
	dryRuns int
	writes  int
}

func (c *runnerContentPlanConnector) Name() string { return "runner-content-fixture" }

func (c *runnerContentPlanConnector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:         c.Name(),
		DisplayName:  "Runner content fixture",
		Capabilities: connectors.Capabilities{Write: true},
	}
}

func (c *runnerContentPlanConnector) Manifest() connectors.Manifest {
	return connectors.Manifest{
		Metadata: c.Metadata(),
		WriteActions: []connectors.WriteActionSpec{{
			Name: "delete_content", Method: http.MethodDelete, Path: "/content/{token}",
		}},
	}
}

func (c *runnerContentPlanConnector) CommandSurface() *connectors.CommandSurface {
	return &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
		Path:         "content delete",
		Intent:       "reverse_etl",
		Availability: "implemented",
		Write:        "delete_content",
		RedactFields: []string{"token", "content", "metadata"},
		Flags: []connectors.CommandSurfaceFlag{
			{Name: "token", Type: "string", MapsTo: "record.token", Required: true},
			{Name: "content", Type: "string", MapsTo: "record.content", Required: true},
			{Name: "nested-content", Type: "string", MapsTo: "record.metadata.request.content", Required: true},
			{Name: "nested-token", Type: "string", MapsTo: "record.metadata.response.token", Required: true},
		},
	}}}
}

func (*runnerContentPlanConnector) Check(context.Context, connectors.RuntimeConfig) error { return nil }

func (*runnerContentPlanConnector) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{}, connectors.ErrUnsupportedOperation
}

func (*runnerContentPlanConnector) Read(context.Context, connectors.ReadRequest, func(connectors.Record) error) error {
	return connectors.ErrUnsupportedOperation
}

func (c *runnerContentPlanConnector) DryRunWrite(_ context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	c.dryRuns++
	return connectors.WritePreview{
		RecordsStaged: len(records),
		Action:        req.Action,
		Digest:        strings.Repeat("c", 64),
		ApprovalTarget: connectors.WriteApprovalTarget{
			Connector:           c.Name(),
			Operation:           req.Action,
			Method:              http.MethodDelete,
			MutationClass:       "delete",
			TargetDigest:        strings.Repeat("d", 64),
			CredentialRevision:  req.Config.CredentialRevision,
			ConfigurationDigest: req.Config.ConfigurationDigest,
			Batchable:           true,
			Scope:               req.Config.WriteApprovalScope,
			Confirmation:        connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
		},
	}, nil
}

func (c *runnerContentPlanConnector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	c.writes++
	return connectors.WriteResult{RecordsWritten: 1}, nil
}
