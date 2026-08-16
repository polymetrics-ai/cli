package app

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
)

// TestCreateConnectionPersistsValidatedTransformPlan exercises the production
// connection-construction path. In particular, it proves the filename is not
// what reaches durable state: only the normalized closed plan and its hash do.
func TestCreateConnectionPersistsValidatedTransformPlan(t *testing.T) {
	ctx := context.Background()
	a, source := newTransformPlanApp(t)
	plan, err := database.ParseTransformPlanV1([]byte(`{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"},{"expr":{"upper":"status"},"target":"status","type":"string"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	created, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name: "typed_events", Source: EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"},
		Streams: map[string]StreamConfig{"events": {
			SyncMode: "full_refresh_overwrite", PrimaryKey: []string{"id"}, CursorField: "updated_at",
			TransformPlan: string(plan.NormalizedJSON()), TransformPlanHash: plan.Hash(),
		}},
	})
	if err != nil {
		t.Fatalf("CreateConnection() error = %v", err)
	}
	stored := created.Streams["events"]
	if got, want := stored.TransformPlan, string(plan.NormalizedJSON()); got != want {
		t.Fatalf("stored transform plan = %q, want normalized %q", got, want)
	}
	if got, want := stored.TransformPlanHash, plan.Hash(); got != want {
		t.Fatalf("stored transform hash = %q, want %q", got, want)
	}
}

// TestCreateConnectionRefusesTransformHashDriftBeforePersistence proves a
// stale preview/approval binding cannot create a connection or reach runtime
// source I/O.
func TestCreateConnectionRefusesTransformHashDriftBeforePersistence(t *testing.T) {
	ctx := context.Background()
	a, source := newTransformPlanApp(t)
	_, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name: "drifted_events", Source: EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"},
		Streams: map[string]StreamConfig{"events": {
			SyncMode: "full_refresh_overwrite", PrimaryKey: []string{"id"}, CursorField: "updated_at",
			TransformPlan:     `{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"}]}`,
			TransformPlanHash: "stale",
		}},
	})
	if !errors.Is(err, database.ErrTransformPlanInvalid) {
		t.Fatalf("CreateConnection() error = %T %v, want ErrTransformPlanInvalid", err, err)
	}
	if got := a.ListConnections(); len(got) != 0 {
		t.Fatalf("CreateConnection() persisted a rejected transform: %#v", got)
	}
}

// TestCreateConnectionAllowsAbsentTransformPlan is the optional-field edge:
// existing connections retain their legacy, untransformed behavior.
func TestCreateConnectionAllowsAbsentTransformPlan(t *testing.T) {
	ctx := context.Background()
	a, source := newTransformPlanApp(t)
	created, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name: "legacy_events", Source: EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"},
		Streams:     map[string]StreamConfig{"events": {SyncMode: "full_refresh_overwrite", PrimaryKey: []string{"id"}, CursorField: "updated_at"}},
	})
	if err != nil {
		t.Fatalf("CreateConnection() without transform error = %v", err)
	}
	if stream := created.Streams["events"]; stream.TransformPlan != "" || stream.TransformPlanHash != "" {
		t.Fatalf("absent transform stored as %#v", stream)
	}
}

type transformPlanSource struct{ *scriptedSyncSource }

func (s *transformPlanSource) Catalog(ctx context.Context, _ connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: s.Name(), Streams: []connectors.Stream{{
		Name: "events", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"},
		Schema: []byte(`{"type":"object","required":["id","updated_at","status"],"properties":{"id":{"type":"integer"},"updated_at":{"type":"string","format":"date-time"},"status":{"type":"string"}}}`),
	}}}, nil
}

func newTransformPlanApp(t *testing.T) (*App, *transformPlanSource) {
	t.Helper()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	source := &transformPlanSource{scriptedSyncSource: newScriptedSyncSource("typed_source", nil)}
	registry := connectors.NewRegistry()
	registry.Register(source)
	a.registry = registry
	ctx := context.Background()
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: source.Name()}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "warehouse", Connector: "warehouse", Config: map[string]string{"path": filepath.Join(root, "warehouse")}}); err != nil {
		t.Fatal(err)
	}
	return a, source
}
