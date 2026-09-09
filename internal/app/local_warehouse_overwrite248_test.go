package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"
)

// This uses the actual embedded source, registry, App coordinator, warehouse
// stage and materializer. Failure before HTTP is not source-fault coverage.
func TestLocalWarehouseCanonicalOverwrite248(t *testing.T) {
	for _, outcome := range []string{"replace", "empty", "later404", "laterMalformed", "cancel"} {
		t.Run(outcome, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var phase atomic.Int32
			var reads atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reads.Add(1)
				if r.Method != "GET" || r.URL.Path != "/groups" || r.URL.Query().Get("per_page") != "50" || r.Header.Get("Authorization") != "Bearer local-overwrite-fixture" {
					t.Errorf("unexpected method/path/query or fixture auth: %s %s", r.Method, r.URL.RequestURI())
					w.WriteHeader(400)
					return
				}
				if phase.Load() == 0 {
					_, _ = w.Write([]byte(`[{"id":900}]`))
					return
				}
				if outcome == "empty" {
					_, _ = w.Write([]byte(`[]`))
					return
				}
				if r.URL.Query().Get("page") == "" {
					w.Header().Set("Link", fmt.Sprintf(`<http://%s/groups?per_page=50&page=2>; rel="next"`, r.Host))
					_, _ = w.Write([]byte(`[{"id":101},{"id":202}]`))
					return
				}
				if r.URL.Query().Get("page") != "2" {
					t.Errorf("unexpected page %q", r.URL.Query().Get("page"))
				}
				switch outcome {
				case "later404":
					w.WriteHeader(404)
				case "laterMalformed":
					_, _ = w.Write([]byte(`[{`))
				case "cancel":
					cancel()
					<-r.Context().Done()
				default:
					_, _ = w.Write([]byte(`[{"id":303}]`))
				}
			}))
			defer server.Close()
			root := t.TempDir()
			if err := InitProject(root); err != nil {
				t.Fatal(err)
			}
			a, err := Open(root)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = a.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: "gitlab", Config: map[string]string{"base_url": server.URL}, Secrets: map[string]string{"access_token": "local-overwrite-fixture"}}); err != nil {
				t.Fatal(err)
			}
			dir := filepath.Join(root, ".polymetrics", "warehouse")
			if _, err = a.AddCredential(ctx, AddCredentialRequest{Name: "warehouse", Connector: "warehouse", Config: map[string]string{"path": dir}}); err != nil {
				t.Fatal(err)
			}
			_, err = a.CreateConnection(ctx, CreateConnectionRequest{Name: "groups_to_warehouse", Source: EndpointConfig{Connector: "gitlab", Credential: "source"}, Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"}, Streams: map[string]StreamConfig{"groups": {SyncMode: "full_append", PrimaryKey: []string{"id"}, DestinationTable: "groups"}}})
			if err != nil {
				t.Fatal(err)
			}
			request := RunETLRequest{Connection: "groups_to_warehouse", Stream: "groups", BatchSize: 2}
			seed, err := a.RunETL(ctx, request)
			if err != nil {
				t.Fatalf("seed: %v", err)
			}
			if seed.RecordsLoaded != 1 {
				t.Fatalf("seed rows=%d", seed.RecordsLoaded)
			}
			conn, _ := a.findConnection(request.Connection)
			loc, err := a.warehouseLocation(dir, conn)
			if err != nil {
				t.Fatal(err)
			}
			table, err := loc.TablePath("groups")
			if err != nil {
				t.Fatal(err)
			}
			wal, err := loc.WALPath("groups")
			if err != nil {
				t.Fatal(err)
			}
			priorTable, err := os.ReadFile(table)
			if err != nil {
				t.Fatal(err)
			}
			priorWAL, err := os.ReadFile(wal)
			if err != nil {
				t.Fatal(err)
			}
			priorState := cloneStreamState(a.state.StreamStates[streamStateKey(request.Connection, "groups")])
			for i := range a.state.Connections {
				if a.state.Connections[i].ID == conn.ID {
					s := a.state.Connections[i].Streams["groups"]
					s.SyncMode = "full_overwrite"
					a.state.Connections[i].Streams["groups"] = s
				}
			}
			if err := a.save(); err != nil {
				t.Fatal(err)
			}
			phase.Store(1)
			reads.Store(0)
			run, runErr := a.RunETL(ctx, request)
			wantReads := int32(2)
			if outcome == "empty" {
				wantReads = 1
			}
			if reads.Load() != wantReads {
				t.Fatalf("canonical source reads=%d want=%d; error=%v (missing port is not source-fault coverage)", reads.Load(), wantReads, runErr)
			}
			reopened, err := Open(root)
			if err != nil {
				t.Fatal(err)
			}
			if outcome == "replace" || outcome == "empty" {
				if runErr != nil {
					t.Fatalf("canonical overwrite: %v", runErr)
				}
				want := []string{"101", "202", "303"}
				if outcome == "empty" {
					want = []string{}
				}
				if run.Status != "completed" || run.RecordsRead != len(want) || run.RecordsLoaded != len(want) {
					t.Fatalf("run=%+v want rows=%d", run, len(want))
				}
				rows, err := reopened.QueryTable(context.Background(), QueryTableRequest{Table: "groups", Limit: 10})
				if err != nil {
					t.Fatal(err)
				}
				got := make([]string, 0, len(rows))
				for _, row := range rows {
					got = append(got, toComparableString(row["id"]))
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("fresh Parquet IDs=%v want=%v", got, want)
				}
				rawIDs := []string{}
				if err := forEachLocalRawRecord(context.Background(), wal, func(raw localRawRecord) error {
					rawIDs = append(rawIDs, toComparableString(raw.Record["id"]))
					return nil
				}); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(rawIDs, want) {
					t.Fatalf("JSONL IDs=%v want=%v", rawIDs, want)
				}
				state := reopened.state.StreamStates[streamStateKey(request.Connection, "groups")]
				if state.LastSuccessfulRunID != run.ID {
					t.Fatalf("terminal run identity=%s want=%s", state.LastSuccessfulRunID, run.ID)
				}
			} else {
				if runErr == nil {
					t.Fatal("source failure succeeded")
				}
				afterTable, err := os.ReadFile(table)
				if err != nil {
					t.Fatal(err)
				}
				afterWAL, err := os.ReadFile(wal)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(priorTable, afterTable) || !bytes.Equal(priorWAL, afterWAL) {
					t.Fatal("source failure changed published WAL/table bytes")
				}
				after := reopened.state.StreamStates[streamStateKey(request.Connection, "groups")]
				beforeJSON, _ := json.Marshal(priorState.Checkpoint)
				afterJSON, _ := json.Marshal(after.Checkpoint)
				if !bytes.Equal(beforeJSON, afterJSON) || after.LastSuccessfulRunID != priorState.LastSuccessfulRunID {
					t.Fatal("source failure advanced committed checkpoint/run")
				}
			}
		})
	}
}
