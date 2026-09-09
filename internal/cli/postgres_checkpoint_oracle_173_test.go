//go:build databaseintegration

package cli

import (
	"encoding/json"
	"polymetrics.ai/internal/synccontract"
	"strings"
	"testing"
	"time"
)

func TestPostgresCheckpointAdvanceOracle173(t *testing.T) {
	const before = `{"s":{"active_work_fence":1,"checkpoint":{"state_version":1,"schema_version":"postgres-cdc-v2","protocol_version":"pgoutput-v2","source_generation":"c291cmNl","snapshot_barrier":{"kind":"postgres_logical_slot","token":"MC8xMA=="},"mechanism":"logical_replication","source":{"engine":"postgres","account_or_cluster":"fixture","object_scope":"public.events"},"position":{"primary":"MC8yMA=="},"committed_at":"2026-09-08T00:00:00Z"}}}`
	advanced := strings.Replace(strings.Replace(before, `"MC8yMA=="`, `"MC8zMA=="`, 1), "00:00:00Z", "00:00:01Z", 1)
	for _, tc := range []struct {
		name, after string
		want        bool
	}{
		{"advanced_position", advanced, true},
		{"identical", before, false},
		{"lease_only", strings.Replace(before, `"active_work_fence":1`, `"active_work_fence":2`, 1), false},
		{"timestamp_only", strings.Replace(before, "00:00:00Z", "00:00:01Z", 1), false},
		{"position_without_commit", strings.Replace(before, `"MC8yMA=="`, `"MC8zMA=="`, 1), false},
		{"wrong_source", strings.Replace(advanced, "fixture", "sibling", 1), false},
		{"wrong_mechanism", strings.Replace(advanced, "logical_replication", "polling", 1), false},
		{"missing_checkpoint", `{"s":{"active_work_fence":2}}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := postgresCheckpointAdvanced173(t, before, tc.after, "s"); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestPostgresCheckpointMonotonicity194(t *testing.T) {
	at := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	later := at.Add(time.Second)
	before := synccontract.CheckpointEnvelope{StateVersion: synccontract.StateVersion, Source: synccontract.SourceIdentity{Engine: "postgres", AccountOrCluster: "fixture", ObjectScope: "public.events"}, Mechanism: "logical_replication", SchemaVersion: "postgres-cdc-v2", ProtocolVersion: "pgoutput-v2", SourceGeneration: synccontract.OpaqueToken("source"), SnapshotBarrier: &synccontract.SnapshotBarrier{Kind: "postgres_logical_slot", Token: synccontract.OpaqueToken("0/10")}, Position: synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("0/20"), TieBreaker: synccontract.OpaqueToken("0/1F")}, CommittedAt: &at}
	encode := func(c synccontract.CheckpointEnvelope) string {
		raw, err := json.Marshal(map[string]any{"s": map[string]any{"checkpoint": c}})
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	forward := before.Clone()
	forward.Position.Primary = synccontract.OpaqueToken("0/30")
	forward.Position.TieBreaker = synccontract.OpaqueToken("0/2F")
	forward.CommittedAt = &later
	cases := []struct {
		name   string
		mutate func(*synccontract.CheckpointEnvelope)
		want   bool
	}{
		{"forward", func(c *synccontract.CheckpointEnvelope) {}, true},
		{"backward", func(c *synccontract.CheckpointEnvelope) { c.Position.Primary = synccontract.OpaqueToken("0/10") }, false},
		{"same_lsn_new_tie_time", func(c *synccontract.CheckpointEnvelope) { c.Position.Primary = before.Position.Primary }, false},
		{"malformed_after", func(c *synccontract.CheckpointEnvelope) { c.Position.Primary = synccontract.OpaqueToken("not-lsn") }, false},
		{"empty_position", func(c *synccontract.CheckpointEnvelope) { c.Position.Primary = nil }, false},
		{"changed_engine", func(c *synccontract.CheckpointEnvelope) { c.Source.Engine = "other" }, false},
		{"changed_account", func(c *synccontract.CheckpointEnvelope) { c.Source.AccountOrCluster = "other" }, false},
		{"changed_scope", func(c *synccontract.CheckpointEnvelope) { c.Source.ObjectScope = "other" }, false},
		{"changed_mechanism", func(c *synccontract.CheckpointEnvelope) { c.Mechanism = "polling" }, false},
		{"changed_state_version", func(c *synccontract.CheckpointEnvelope) { c.StateVersion++ }, false},
		{"changed_schema", func(c *synccontract.CheckpointEnvelope) { c.SchemaVersion = "other" }, false},
		{"changed_protocol", func(c *synccontract.CheckpointEnvelope) { c.ProtocolVersion = "other" }, false},
		{"changed_generation", func(c *synccontract.CheckpointEnvelope) { c.SourceGeneration = synccontract.OpaqueToken("other") }, false},
		{"changed_barrier", func(c *synccontract.CheckpointEnvelope) { c.SnapshotBarrier.Token = synccontract.OpaqueToken("0/F") }, false},
		{"missing_barrier", func(c *synccontract.CheckpointEnvelope) { c.SnapshotBarrier = nil }, false},
		{"same_commit_time", func(c *synccontract.CheckpointEnvelope) { c.CommittedAt = &at }, false},
		{"missing_commit", func(c *synccontract.CheckpointEnvelope) { c.CommittedAt = nil }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			after := forward.Clone()
			tc.mutate(&after)
			got := postgresCheckpointAdvanced173(t, encode(before), encode(after), "s")
			if got != tc.want {
				t.Fatalf("got=%v want=%v before=%q after=%q", got, tc.want, before.Position.Primary, after.Position.Primary)
			}
		})
	}
	t.Run("unchanged", func(t *testing.T) {
		if postgresCheckpointAdvanced173(t, encode(before), encode(before), "s") {
			t.Fatal("unchanged checkpoint advanced")
		}
	})
	t.Run("malformed_before", func(t *testing.T) {
		bad := before.Clone()
		bad.Position.Primary = synccontract.OpaqueToken("bad")
		if postgresCheckpointAdvanced173(t, encode(bad), encode(forward), "s") {
			t.Fatal("invalid predecessor advanced")
		}
	})
}
