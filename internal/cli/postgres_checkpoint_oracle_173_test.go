//go:build databaseintegration

package cli

import (
	"strings"
	"testing"
)

func TestPostgresCheckpointAdvanceOracle173(t *testing.T) {
	const before = `{"s":{"active_work_fence":1,"checkpoint":{"mechanism":"logical_replication","source":{"engine":"postgres","account_or_cluster":"fixture","object_scope":"public.events"},"position":{"primary":"AQ=="},"committed_at":"2026-09-08T00:00:00Z"}}}`
	advanced := strings.Replace(strings.Replace(before, `"AQ=="`, `"Ag=="`, 1), "00:00:00Z", "00:00:01Z", 1)
	for _, tc := range []struct {
		name, after string
		want        bool
	}{
		{"advanced_position", advanced, true},
		{"identical", before, false},
		{"lease_only", strings.Replace(before, `"active_work_fence":1`, `"active_work_fence":2`, 1), false},
		{"timestamp_only", strings.Replace(before, "00:00:00Z", "00:00:01Z", 1), false},
		{"position_without_commit", strings.Replace(before, `"AQ=="`, `"Ag=="`, 1), false},
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
