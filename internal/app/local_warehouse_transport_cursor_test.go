package app

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

// Full snapshots are bounded by their completed request sequence, not by a
// row cursor. A configured legacy cursor must therefore never turn a source
// operation with no cited cursor field into a failed local warehouse write.
func TestLocalWarehouseTransportRawRecordsFullSnapshotsDoNotRequireRowCursor(t *testing.T) {
	tests := []struct {
		name     string
		strategy connectors.DestinationApplyStrategy
	}{
		{
			name:     "full append",
			strategy: connectors.DestinationApplyStrategy{Mode: synccontract.ModeFullAppend, Strategy: connectors.ApplyStrategyAppend},
		},
		{
			name:     "full overwrite",
			strategy: connectors.DestinationApplyStrategy{Mode: synccontract.ModeFullOverwrite, Strategy: connectors.ApplyStrategyReplace},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records, err := localWarehouseTransportRawRecords(
				synctransport.WarehouseReceipt{ID: "full-snapshot", Generation: 1},
				synctransport.WarehouseWorkset{Records: []connectors.Record{{"id": "without-cursor"}}},
				StreamConfig{CursorField: "created_at"},
				tt.strategy,
			)
			if err != nil {
				t.Fatalf("full snapshot raw records: %v", err)
			}
			if len(records) != 1 || records[0].Cursor != "" {
				t.Fatalf("full snapshot raw records = %+v, want one record with no invented cursor", records)
			}
		})
	}
}

func TestLocalWarehouseTransportRawRecordsRetainsIncrementalAndHistoryCursorRequirements(t *testing.T) {
	tests := []struct {
		name       string
		workset    synctransport.WarehouseWorkset
		stream     StreamConfig
		strategy   connectors.DestinationApplyStrategy
		wantErr    string
		wantCursor string
	}{
		{
			name:     "incremental append rejects an empty durable checkpoint",
			workset:  synctransport.WarehouseWorkset{Records: []connectors.Record{{"id": "incremental"}}},
			strategy: connectors.DestinationApplyStrategy{Mode: synccontract.ModeIncrementalAppend, Strategy: connectors.ApplyStrategyAppend},
			wantErr:  "requires a durable source checkpoint",
		},
		{
			name: "incremental append retains the opaque durable checkpoint",
			workset: synctransport.WarehouseWorkset{
				Records:             []connectors.Record{{"id": "incremental"}},
				CandidateCheckpoint: synccontract.CheckpointEnvelope{Position: synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("provider-position")}},
			},
			strategy:   connectors.DestinationApplyStrategy{Mode: synccontract.ModeIncrementalAppend, Strategy: connectors.ApplyStrategyAppend},
			wantCursor: "provider-position",
		},
		{
			name:     "dedupe history rejects a missing declared row cursor",
			workset:  synctransport.WarehouseWorkset{Records: []connectors.Record{{"id": "history"}}},
			stream:   StreamConfig{PrimaryKey: []string{"id"}, CursorField: "updated_at"},
			strategy: connectors.DestinationApplyStrategy{Mode: synccontract.ModeIncrementalDedupeHistory, Strategy: connectors.ApplyStrategyDedupeHistory},
			wantErr:  "missing cursor field \"updated_at\"",
		},
		{
			name:       "dedupe history retains the declared row cursor",
			workset:    synctransport.WarehouseWorkset{Records: []connectors.Record{{"id": "history", "updated_at": "2026-08-29T00:00:00Z"}}},
			stream:     StreamConfig{PrimaryKey: []string{"id"}, CursorField: "updated_at"},
			strategy:   connectors.DestinationApplyStrategy{Mode: synccontract.ModeIncrementalDedupeHistory, Strategy: connectors.ApplyStrategyDedupeHistory},
			wantCursor: "2026-08-29T00:00:00Z",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records, err := localWarehouseTransportRawRecords(
				synctransport.WarehouseReceipt{ID: "cursor-requirements", Generation: 1},
				tt.workset,
				tt.stream,
				tt.strategy,
			)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("raw records error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("raw records: %v", err)
			}
			if len(records) != 1 || records[0].Cursor != tt.wantCursor {
				t.Fatalf("raw records = %+v, want one record with cursor %q", records, tt.wantCursor)
			}
		})
	}
}
