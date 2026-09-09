package app

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/connectors"
)

var errReadInput234 = errors.New("selected request input refused")

type guardedReadSource234 struct {
	streamingSource
	calls int
	seen  connectors.ReadInputValidationRequest
}

func (s *guardedReadSource234) ValidateReadInputs(_ context.Context, req connectors.ReadInputValidationRequest) error {
	s.calls++
	s.seen = req
	if req.Config["page_size"] != "2" {
		return errReadInput234
	}
	return nil
}

func TestRunETLReadInputPreSecret234(t *testing.T) {
	for _, valid := range []bool{false, true} {
		name := "invalid"
		if valid {
			name = "overlay_healthy"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			root := t.TempDir()
			if err := InitProject(root); err != nil {
				t.Fatal(err)
			}
			a, err := Open(root)
			if err != nil {
				t.Fatal(err)
			}
			source := &guardedReadSource234{streamingSource: streamingSource{total: 5}}
			destination := &batchDestination{}
			registry := connectors.NewRegistry()
			if err := registry.Register(source); err != nil {
				t.Fatal(err)
			}
			if err := registry.Register(destination); err != nil {
				t.Fatal(err)
			}
			a.registry = registry
			if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: source.Name(), Config: map[string]string{"page_size": "invalid"}}); err != nil {
				t.Fatal(err)
			}
			if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "dest", Connector: destination.Name(), Config: map[string]string{"path": filepath.Join(root, "out")}}); err != nil {
				t.Fatal(err)
			}
			overlay := map[string]string{}
			if valid {
				overlay["page_size"] = "2"
			}
			if _, err := a.CreateConnection(ctx, CreateConnectionRequest{Name: "source_to_dest", Source: EndpointConfig{Connector: source.Name(), Credential: "source", Config: overlay}, Destination: EndpointConfig{Connector: destination.Name(), Credential: "dest"}, Streams: map[string]StreamConfig{"records": {SyncMode: "full_refresh_overwrite", PrimaryKey: []string{"id"}, DestinationTable: "records"}}}); err != nil {
				t.Fatal(err)
			}
			beforeRuns := len(a.state.Runs)
			if !valid {
				// An accidental source or destination vault read is a reached failure,
				// rather than a mock that silently reports successful non-access.
				a.vault = nil
				defer func() {
					if value := recover(); value != nil {
						t.Fatalf("invalid input reached vault before validation: %v", value)
					}
				}()
			}
			run, err := a.RunETL(ctx, RunETLRequest{Connection: "source_to_dest", Stream: "records", BatchSize: 2})
			if source.calls != 1 || source.seen.Stream != "records" {
				t.Fatalf("selected validation calls=%d request=%+v", source.calls, source.seen)
			}
			if valid {
				if err != nil || run.RecordsRead != 5 || run.RecordsLoaded != 5 || len(destination.batches) != 3 {
					t.Fatalf("healthy returned rows/batches: %+v %v %v", run, err, destination.batches)
				}
				if source.seen.Config["page_size"] != "2" {
					t.Fatal("endpoint override did not reach validation")
				}
			} else if !errors.Is(err, errReadInput234) || len(a.state.Runs) != beforeRuns || len(destination.batches) != 0 {
				t.Fatalf("invalid frontier: err=%v runs=%d batches=%v", err, len(a.state.Runs), destination.batches)
			}
		})
	}
}
