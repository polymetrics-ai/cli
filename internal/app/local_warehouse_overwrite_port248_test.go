package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
	"polymetrics.ai/internal/warehouse"
)

func overwritePortFixture248(t *testing.T) (*localWarehouseDestinationExecutor, synctransport.FullOverwriteRunRequest, string) {
	t.Helper()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	conn := Connection{ID: "overwrite_owner", Name: "overwrite_owner", Source: EndpointConfig{Connector: "gitlab"}, Destination: EndpointConfig{Connector: "warehouse"}, Streams: map[string]StreamConfig{"groups": {StreamID: "groups_identity", SyncMode: "full_overwrite", DestinationTable: "groups", PrimaryKey: []string{"id"}}}}
	a.state.Connections = append(a.state.Connections, conn)
	dst, ok := a.registry.Get("warehouse")
	if !ok {
		t.Fatal("warehouse absent")
	}
	source, ok := a.registry.Get("gitlab")
	if !ok {
		t.Fatal("gitlab absent")
	}
	executor, err := newLocalWarehouseDestinationExecutor(a, dst)
	if err != nil {
		t.Fatal(err)
	}
	strategy, err := localWarehouseApplyStrategy(synccontract.ModeFullOverwrite)
	if err != nil {
		t.Fatal(err)
	}
	req := synctransport.FullOverwriteRunRequest{ConnectionID: conn.ID, Generation: 1, Plan: synctransport.DestinationPlan{ApplyStrategy: strategy}, Runtime: connectors.RuntimeConfig{ProjectDir: root, Config: map[string]string{"path": filepath.Join(root, "warehouse")}}, Source: source, Binding: synctransport.DestinationBinding{WorkspaceID: a.state.WorkspaceID, SourceConnectorID: "gitlab", ConnectionID: conn.ID, StreamID: "groups_identity", PrimaryKey: []string{"id"}}, Stream: "groups", Mode: synccontract.ModeFullOverwrite, BatchSize: 2}
	loc, err := a.warehouseLocation(req.Runtime.Config["path"], conn)
	if err != nil {
		t.Fatal(err)
	}
	if err := loc.EnsureOwnership(); err != nil {
		t.Fatal(err)
	}
	table, err := loc.TablePath("groups")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(table), 0700); err != nil {
		t.Fatal(err)
	}
	if err := warehouse.WriteTable(context.Background(), table, []warehouse.Row{{"id": "old"}}); err != nil {
		t.Fatal(err)
	}
	return executor, req, table
}

func overwritePage248(req synctransport.FullOverwriteRunRequest) synctransport.DestinationApplyRequest {
	receipt := synctransport.WarehouseReceipt{ID: "page_one", Owner: req.ConnectionID, Generation: req.Generation, Stream: req.Stream, Mode: req.Mode, CheckpointSHA256: "checkpoint", TombstonesSHA256: "tombstones", ManifestSHA256: "manifest", ContentSHA256: "content", ParquetSHA256: "parquet", Records: 2}
	return synctransport.DestinationApplyRequest{ConnectionID: req.ConnectionID, Plan: req.Plan, Receipt: receipt, Workset: synctransport.WarehouseWorkset{ID: receipt.ID, Records: []connectors.Record{{"id": "new1"}, {"id": "new2"}}}, Runtime: req.Runtime, Source: req.Source, Binding: req.Binding, Stream: req.Stream, Mode: req.Mode, BatchSize: req.BatchSize}
}

func requireOverwritePort248(t *testing.T, e *localWarehouseDestinationExecutor) synctransport.FullOverwriteDestination {
	t.Helper()
	port, ok := any(e).(synctransport.FullOverwriteDestination)
	if !ok {
		t.Fatal("actual local warehouse executor lacks canonical run-scoped overwrite port")
	}
	return port
}

func TestLocalWarehouseOverwritePort248(t *testing.T) {
	for _, name := range []string{"mode", "connection", "workspace", "source", "stream_identity", "stream", "generation", "batch", "transform"} {
		t.Run("begin_"+name, func(t *testing.T) {
			e, r, table := overwritePortFixture248(t)
			port := requireOverwritePort248(t, e)
			switch name {
			case "mode":
				r.Mode = synccontract.ModeFullAppend
			case "connection":
				r.ConnectionID = "different"
			case "workspace":
				r.Binding.WorkspaceID = "different"
			case "source":
				r.Binding.SourceConnectorID = "different"
			case "stream_identity":
				r.Binding.StreamID = "different"
			case "stream":
				r.Stream = "other"
			case "generation":
				r.Generation = 0
			case "batch":
				r.BatchSize = 0
			case "transform":
				r.TransformPlanJSON = "{}"
			}
			before, err := os.ReadFile(table)
			if err != nil {
				t.Fatal(err)
			}
			run, err := port.BeginFullOverwrite(context.Background(), r)
			if err == nil {
				if run != nil {
					_ = run.AbortFullOverwrite(context.Background())
				}
				t.Fatal("invalid binding admitted")
			}
			after, err := os.ReadFile(table)
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) {
				t.Fatal("invalid begin changed public table")
			}
		})
	}
	for _, name := range []string{"owner", "workset", "generation", "stream", "mode", "count", "tombstones", "binding", "canceled", "aggregate"} {
		t.Run("page_"+name, func(t *testing.T) {
			e, r, table := overwritePortFixture248(t)
			port := requireOverwritePort248(t, e)
			run, err := port.BeginFullOverwrite(context.Background(), r)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := run.AbortFullOverwrite(context.Background()); err != nil {
					t.Error(err)
				}
			}()
			p := overwritePage248(r)
			ctx := context.Background()
			switch name {
			case "owner":
				p.Receipt.Owner = "other"
			case "workset":
				p.Workset.ID = "other"
			case "generation":
				p.Receipt.Generation++
			case "stream":
				p.Receipt.Stream = "other"
			case "mode":
				p.Receipt.Mode = synccontract.ModeFullAppend
			case "count":
				p.Receipt.Records++
			case "tombstones":
				p.Workset.Tombstones = append(p.Workset.Tombstones, synccontract.Tombstone{})
			case "binding":
				p.Binding.ConnectionID = "other"
			case "canceled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			before, err := os.ReadFile(table)
			if err != nil {
				t.Fatal(err)
			}
			applyErr := run.ApplyFullOverwrite(ctx, p)
			if name == "aggregate" {
				if applyErr != nil {
					t.Fatal(applyErr)
				}
				_, err = run.PublishFullOverwrite(ctx, synctransport.FullOverwritePublicationRequest{Pages: 1, Records: 3})
				if err == nil {
					t.Fatal("incorrect aggregate published")
				}
			} else if applyErr == nil {
				t.Fatal("invalid page admitted")
			}
			after, err := os.ReadFile(table)
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) {
				t.Fatal("unpublished page changed public table")
			}
		})
	}
	t.Run("publish_readback_terminal", func(t *testing.T) {
		e, r, table := overwritePortFixture248(t)
		run, err := requireOverwritePort248(t, e).BeginFullOverwrite(context.Background(), r)
		if err != nil {
			t.Fatal(err)
		}
		p := overwritePage248(r)
		if err := run.ApplyFullOverwrite(context.Background(), p); err != nil {
			t.Fatal(err)
		}
		ack, err := run.PublishFullOverwrite(context.Background(), synctransport.FullOverwritePublicationRequest{Pages: 1, Records: 2})
		if err != nil {
			t.Fatal(err)
		}
		if err := run.ReadBackFullOverwrite(context.Background(), ack); err != nil {
			t.Fatal(err)
		}
		originalOutput := append([]byte(nil), ack.Output...)
		ack.Output[0] = '['
		if err := run.ReadBackFullOverwrite(context.Background(), ack); err == nil {
			t.Fatal("in-place acknowledgement mutation changed the saved publication witness")
		}
		copy(ack.Output, originalOutput)
		changed := ack
		changed.Output = json.RawMessage(`{}`)
		if err := run.ReadBackFullOverwrite(context.Background(), changed); err == nil {
			t.Fatal("different receipt admitted")
		}
		if err := run.ApplyFullOverwrite(context.Background(), p); err == nil {
			t.Fatal("page after publication admitted")
		}
		if _, err := run.PublishFullOverwrite(context.Background(), synctransport.FullOverwritePublicationRequest{Pages: 1, Records: 2}); err == nil {
			t.Fatal("second publication admitted")
		}
		if err := run.AbortFullOverwrite(context.Background()); err != nil {
			t.Fatal(err)
		}
		if err := run.ReadBackFullOverwrite(context.Background(), ack); err != nil {
			t.Fatalf("abort rolled back published output: %v", err)
		}
		if err := warehouse.WriteTable(context.Background(), table, []warehouse.Row{{"id": "tampered"}}); err != nil {
			t.Fatal(err)
		}
		if err := run.ReadBackFullOverwrite(context.Background(), ack); err == nil {
			t.Fatal("read-back accepted changed table")
		}
	})
	t.Run("empty_recovery", func(t *testing.T) {
		e, r, _ := overwritePortFixture248(t)
		run, err := requireOverwritePort248(t, e).BeginFullOverwrite(context.Background(), r)
		if err != nil {
			t.Fatal(err)
		}
		ack, err := run.PublishFullOverwrite(context.Background(), synctransport.FullOverwritePublicationRequest{})
		if err != nil {
			t.Fatal(err)
		}
		witness, err := ack.PublicationWitness()
		if err != nil {
			t.Fatal(err)
		}
		dst, _ := e.app.registry.Get("warehouse")
		fresh, err := newLocalWarehouseDestinationExecutor(e.app, dst)
		if err != nil {
			t.Fatal(err)
		}
		recovery, ok := any(fresh).(synctransport.EmptyPublicationReadBackDestination)
		if !ok {
			t.Fatal("missing durable empty recovery")
		}
		req := synctransport.EmptyPublicationReadBackRequest{Runtime: r.Runtime, Source: r.Source, Binding: r.Binding, Stream: r.Stream, Receipt: synctransport.EmptyPublicationReadBackReceipt{Witness: witness, Output: ack.Output}}
		if err := recovery.ReadBackEmptyFullOverwrite(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		req.Binding.StreamID = "other"
		if err := recovery.ReadBackEmptyFullOverwrite(context.Background(), req); err == nil {
			t.Fatal("empty recovery accepted wrong stream identity")
		}
	})
}

func TestLocalWarehouseOverwritePublicationCuts248(t *testing.T) {
	for _, cut := range []string{"private_sync", "wal_rename", "wal_sync", "table_rename", "table_sync"} {
		t.Run(cut, func(t *testing.T) {
			e, r, table := overwritePortFixture248(t)
			conn, _ := e.app.findConnectionByID(r.ConnectionID)
			loc, err := e.app.warehouseLocation(r.Runtime.Config["path"], conn)
			if err != nil {
				t.Fatal(err)
			}
			wal, err := loc.WALPath(r.Stream)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Dir(wal), 0700); err != nil {
				t.Fatal(err)
			}
			oldWAL := []byte("{\"record\":{\"id\":\"old\"}}\n")
			if err := os.WriteFile(wal, oldWAL, 0600); err != nil {
				t.Fatal(err)
			}
			oldTable, err := os.ReadFile(table)
			if err != nil {
				t.Fatal(err)
			}
			run, err := requireOverwritePort248(t, e).BeginFullOverwrite(context.Background(), r)
			if err != nil {
				t.Fatal(err)
			}
			if err := run.ApplyFullOverwrite(context.Background(), overwritePage248(r)); err != nil {
				t.Fatal(err)
			}
			originalRename := localWarehouseOverwriteRename
			originalSync := syncLocalWarehouseDirectoryCommit
			reached := false
			walPublished := false
			tablePublished := false
			localWarehouseOverwriteRename = func(from, to string) error {
				if (cut == "wal_rename" && to == wal) || (cut == "table_rename" && to == table) {
					reached = true
					return fmt.Errorf("injected publication cut %s", cut)
				}
				err := originalRename(from, to)
				if err == nil {
					if to == wal {
						walPublished = true
					}
					if to == table {
						tablePublished = true
					}
				}
				return err
			}
			syncLocalWarehouseDirectoryCommit = func(path string) error {
				if (cut == "private_sync" && strings.Contains(filepath.Base(path), ".overwrite-")) || (cut == "wal_sync" && walPublished && path == filepath.Dir(wal)) || (cut == "table_sync" && tablePublished && path == filepath.Dir(table)) {
					reached = true
					return fmt.Errorf("injected publication cut %s", cut)
				}
				return originalSync(path)
			}
			t.Cleanup(func() {
				localWarehouseOverwriteRename = originalRename
				syncLocalWarehouseDirectoryCommit = originalSync
			})
			ack, publishErr := run.PublishFullOverwrite(context.Background(), synctransport.FullOverwritePublicationRequest{Pages: 1, Records: 2})
			localWarehouseOverwriteRename = originalRename
			syncLocalWarehouseDirectoryCommit = originalSync
			if publishErr == nil || !reached {
				t.Fatalf("cut not reached/error missing: reached=%t err=%v", reached, publishErr)
			}
			if _, err := ack.PublicationWitness(); err == nil {
				t.Fatal("failed durability cut produced checkpoint authority")
			}
			if err := run.AbortFullOverwrite(context.Background()); err != nil {
				t.Fatal(err)
			}
			if err := run.AbortFullOverwrite(context.Background()); err != nil {
				t.Fatal(err)
			}
			afterTable, err := os.ReadFile(table)
			if err != nil {
				t.Fatal(err)
			}
			afterWAL, err := os.ReadFile(wal)
			if err != nil {
				t.Fatal(err)
			}
			if tablePublished == bytes.Equal(oldTable, afterTable) {
				t.Fatal("table state does not match actual rename frontier")
			}
			if walPublished == bytes.Equal(oldWAL, afterWAL) {
				t.Fatal("WAL state does not match actual rename frontier")
			}
			if walPublished {
				ids := []string{}
				if err := forEachLocalRawRecord(context.Background(), wal, func(row localRawRecord) error { ids = append(ids, toComparableString(row.Record["id"])); return nil }); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(ids, []string{"new1", "new2"}) {
					t.Fatalf("complete WAL lost after cut: %v", ids)
				}
			}
		})
	}
}

func TestLocalWarehouseOverwriteRejectsPerPageReplacement248(t *testing.T) {
	e, r, table := overwritePortFixture248(t)
	before, err := os.ReadFile(table)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.ApplyDestination(context.Background(), overwritePage248(r)); err == nil {
		t.Fatal("per-page full_overwrite bypassed the run-scoped publication port")
	}
	after, err := os.ReadFile(table)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("refused per-page replacement changed public table")
	}
}
