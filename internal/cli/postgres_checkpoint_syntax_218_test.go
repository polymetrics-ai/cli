//go:build databaseintegration

package cli

import (
	"encoding/json"
	"polymetrics.ai/internal/synccontract"
	"testing"
	"time"
)

// Retained205 counterexamples. Expected LSN ordering and validity are literal fixture facts.
func TestPostgresCheckpointSyntax218(t *testing.T) {
	at := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	envelope := func(position string, when time.Time) string {
		c := synccontract.CheckpointEnvelope{StateVersion: 1, Source: synccontract.SourceIdentity{Engine: "postgres", AccountOrCluster: "fixture", ObjectScope: "public.events"}, Mechanism: "logical_replication", SchemaVersion: "postgres-cdc-v2", ProtocolVersion: "pgoutput-v2", SourceGeneration: synccontract.OpaqueToken("source"), SnapshotBarrier: &synccontract.SnapshotBarrier{Kind: "postgres_logical_slot", Token: synccontract.OpaqueToken("0/1")}, Position: synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken(position)}, CommittedAt: &when}
		raw, err := json.Marshal(map[string]any{"s": map[string]any{"checkpoint": c}})
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	for _, tc := range []struct {
		name, before, after string
		seconds             int
		want                bool
	}{
		{"digits_forward", "0/20", "0/30", 1, true},
		{"hex_forward", "0/F", "0/10", 1, true},
		{"half_carry", "1/FFFFFFFF", "2/0", 1, true},
		{"maximum_valid_forward", "FFFFFFFF/FFFFFFFE", "FFFFFFFF/FFFFFFFF", 1, true},
		{"backward_later_time", "0/20", "0/10", 1, false},
		{"equal_later_time", "0/20", "0/20", 1, false},
		{"equal_case_variant", "0/f", "0/F", 1, false},
		{"forward_equal_time", "0/20", "0/30", 0, false},
		{"forward_earlier_time", "0/20", "0/30", -1, false},
		{"after_trailing_junk", "0/20", "0/30junk", 1, false},
		{"before_trailing_junk", "0/20junk", "0/30", 1, false},
		{"after_trailing_component", "0/20", "0/30/40", 1, false},
		{"after_overwide_lower", "0/20", "0/100000030", 1, false},
		{"after_overwide_upper", "0/20", "100000000/30", 1, false},
		{"after_scalar_decimal", "0/20", "48", 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := postgresCheckpointAdvanced173(t, envelope(tc.before, at), envelope(tc.after, at.Add(time.Duration(tc.seconds)*time.Second)), "s")
			if got != tc.want {
				t.Errorf("advance(%q -> %q, commit delta=%ds)=%t want=%t", tc.before, tc.after, tc.seconds, got, tc.want)
			}
		})
	}
}

func TestPostgresCheckpointSyntaxSiblings218(t *testing.T) {
	at := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	encode := func(position string, when time.Time) string {
		c := synccontract.CheckpointEnvelope{StateVersion: 1, Source: synccontract.SourceIdentity{Engine: "postgres", AccountOrCluster: "fixture", ObjectScope: "public.events"}, Mechanism: "logical_replication", SchemaVersion: "postgres-cdc-v2", ProtocolVersion: "pgoutput-v2", SourceGeneration: synccontract.OpaqueToken("source"), SnapshotBarrier: &synccontract.SnapshotBarrier{Kind: "postgres_logical_slot", Token: synccontract.OpaqueToken("0/1")}, Position: synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken(position)}, CommittedAt: &when}
		raw, err := json.Marshal(map[string]any{"s": map[string]any{"checkpoint": c}})
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	for _, tc := range []struct{ name, token string }{
		{"empty", ""}, {"missing_upper", "/30"}, {"missing_lower", "0/"}, {"no_separator", "30"}, {"spaces", " 0/30 "}, {"leading_space", " 0/30"}, {"trailing_space", "0/30 "}, {"inner_space", "0/ 30"}, {"tab", "0/30\t"}, {"newline", "0/30\n"}, {"plus_upper", "+0/30"}, {"minus_upper", "-0/30"}, {"plus_lower", "0/+30"}, {"minus_lower", "0/-30"}, {"extra_separator", "0//30"}, {"suffix_separator", "0/30/"}, {"prefix_separator", "/0/30"}, {"hex_prefix", "0/0x30"}, {"unicode_digit", "0/３0"},
	} {
		for _, side := range []string{"before", "after"} {
			t.Run(side+"/"+tc.name, func(t *testing.T) {
				before, after := "0/20", "0/30"
				if side == "before" {
					before = tc.token
				} else {
					after = tc.token
				}
				if postgresCheckpointAdvanced173(t, encode(before, at), encode(after, at.Add(time.Second)), "s") {
					t.Fatal("malformed LSN certified as advancement")
				}
			})
		}
	}
}
