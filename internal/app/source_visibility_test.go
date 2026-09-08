package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestindex"
)

func TestSourceVisibilityAppPublicBoundaries162(t *testing.T) {
	entries := []connectors.LazyRegistryEntry{}
	for _, entry := range manifestindex.GeneratedEntries() {
		if entry.Connector == "asana" || entry.Connector == "vercel" || entry.Connector == "github" {
			entries = append(entries, connectors.LazyRegistryEntry{Metadata: entry.Metadata, SourceVisibility: entry.SourceVisibility})
		}
	}
	loads := 0
	registry, err := connectors.NewLazyRegistryWithEntries(entries, func(ctx context.Context, name string) (connectors.Connector, error) {
		loads++
		b, e := engine.Load(defs.FS, name)
		if e != nil {
			return nil, e
		}
		return engine.New(b, nil), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	instance := &App{registry: registry}
	for _, tc := range []struct{ connector, id, lane, code string }{
		{"asana", "asana.rest.removeCustomFieldSettingForGoal", "direct_write", "source_mapping_unproven"},
		{"asana", "asana.rest.removeCustomFieldSettingForGoal", "direct_read", "source_lane_not_applicable"},
		{"vercel", "vercel.rest.createWebhook", "sync_transport", "missing_foundation"},
	} {
		selection := connectors.SourceCellSelection{Source: connectors.SourceOperationKey{Connector: tc.connector, Inventory: "primary", ID: tc.id}, Lane: connectors.SourceLane(tc.lane)}
		_, err = instance.PreflightConnectorSource(t.Context(), selection)
		var typed *connectors.SourceSelectionError
		if !errors.As(err, &typed) || typed.Code != tc.code || typed.Selection != selection || loads != 0 {
			t.Fatalf("actual App preflight: %v loads=%d", err, loads)
		}
		_, err = PreflightConnectorSource(t.Context(), registry, selection)
		if !errors.As(err, &typed) || typed.Code != tc.code || loads != 0 {
			t.Fatalf("before-Open entry: %v loads=%d", err, loads)
		}
	}
	catalog, err := instance.ConnectorSources(t.Context(), "asana")
	if err != nil || len(catalog.Operations) != 249 || loads != 0 {
		t.Fatalf("App discovery: %v loads=%d", err, loads)
	}
	selection := connectors.SourceCellSelection{Source: connectors.SourceOperationKey{Connector: "asana", Inventory: "other", ID: "asana.rest.removeCustomFieldSettingForGoal"}, Lane: "direct_write"}
	_, err = instance.PreflightConnectorSource(t.Context(), selection)
	var input *connectors.SourceSelectionInputError
	if !errors.As(err, &input) || loads != 0 {
		t.Fatalf("App unknown tuple: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = instance.PreflightConnectorSource(ctx, selection)
	if !errors.Is(err, context.Canceled) || loads != 0 {
		t.Fatalf("App cancellation: %v", err)
	}
	// Same real resolver can fire and reach ordinary command validation; it was
	// not an unconditional zero counter or a substitute unavailable executor.
	_, _, err = instance.PlanConnectorCommand(t.Context(), PlanConnectorCommandRequest{Connector: "github", Path: []string{"label", "delete"}})
	if err == nil || loads != 1 {
		t.Fatalf("ordinary real command boundary not reached: %v loads=%d", err, loads)
	}
}

func TestSourceVisibilityAppMalformedCause163(t *testing.T) {
	raw := `{"schema_version":1,"operations":!}`
	h := sha256.Sum256([]byte(raw))
	a := connectors.SourceVisibilityArtifact{SchemaVersion: 1, Connector: "asana", Coverage: "in_cohort", Bytes: len(raw), Payload: raw, SHA256: hex.EncodeToString(h[:])}
	loads := 0
	r, err := connectors.NewLazyRegistryWithEntries([]connectors.LazyRegistryEntry{{Metadata: connectors.Metadata{Name: "asana"}, SourceVisibility: a}}, func(context.Context, string) (connectors.Connector, error) {
		loads++
		return nil, errors.New("unexpected execution")
	})
	if err != nil {
		t.Fatal(err)
	}
	instance := &App{registry: r}
	_, err = instance.PreflightConnectorSource(t.Context(), connectors.SourceCellSelection{Source: connectors.SourceOperationKey{Connector: "asana", Inventory: "primary", ID: "asana.rest.addCustomFieldSettingForGoal"}, Lane: "binary_download"})
	var data *connectors.SourceVisibilityDataError
	var syntax *json.SyntaxError
	if !errors.As(err, &data) || !errors.As(err, &syntax) || loads != 0 {
		t.Fatalf("actual decode cause lost/crossed boundary: %v loads=%d", err, loads)
	}
}
