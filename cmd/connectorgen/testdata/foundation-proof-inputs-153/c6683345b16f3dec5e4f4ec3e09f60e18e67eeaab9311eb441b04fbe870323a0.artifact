package engine

import (
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
)

// writesWithBatchable renders a writes.json whose single action carries the
// supplied raw JSON for the "batchable" key. An empty declaration omits the key
// entirely, which is how every bundle in defs/ is written today.
func writesWithBatchable(declaration string) []byte {
	batchable := ""
	if declaration != "" {
		batchable = `"batchable": ` + declaration + `,`
	}
	return []byte(`{
		"actions": [
			{
				"name": "cast_vote",
				"kind": "custom",
				"method": "POST",
				"path": "/api/vote",
				` + batchable + `
				"body_type": "form",
				"body_fields": ["id", "dir"],
				"record_schema": {
					"type": "object",
					"required": ["id", "dir"],
					"properties": {
						"id": { "type": "string" },
						"dir": { "type": "integer" }
					}
				},
				"risk": "casts a vote as the authenticated user"
			}
		]
	}`)
}

func loadBundleWithBatchable(t *testing.T, declaration string) Bundle {
	t.Helper()
	fsys := fullValidBundleFS("acme")
	fsys["acme/writes.json"] = &fstest.MapFile{Data: writesWithBatchable(declaration)}
	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load(batchable=%s): %v", declaration, err)
	}
	if len(b.Writes) != 1 {
		t.Fatalf("Writes = %d, want 1", len(b.Writes))
	}
	return b
}

// R1/R2: the declaration parses, and every shape that does not explicitly say
// "false" must be batchable. The Go zero value is included deliberately: bool's
// zero value is false, and false here means non-batchable, so a plain bool field
// would silently mark every hand-constructed action as non-batchable.
func TestWriteActionBatchableDefaultsToTrue(t *testing.T) {
	for _, tc := range []struct {
		name        string
		declaration string
		want        bool
	}{
		{name: "absent", declaration: "", want: true},
		{name: "explicit true", declaration: "true", want: true},
		{name: "explicit false", declaration: "false", want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := loadBundleWithBatchable(t, tc.declaration)
			if got := b.Writes[0].IsBatchable(); got != tc.want {
				t.Fatalf("IsBatchable() = %v, want %v", got, tc.want)
			}
		})
	}

	t.Run("go zero value", func(t *testing.T) {
		var zero WriteAction
		if !zero.IsBatchable() {
			t.Fatal("zero-value WriteAction.IsBatchable() = false; the safe default must survive a bare struct literal")
		}
	})
}

func TestWriteActionBatchableRejectsNonBoolean(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/writes.json"] = &fstest.MapFile{Data: writesWithBatchable(`"false"`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatal("Load() error = nil, want a rejection for a non-boolean batchable")
	}
	if !strings.Contains(err.Error(), "writes.json") || !strings.Contains(err.Error(), "batchable") {
		t.Fatalf("Load() error = %q, want it to name writes.json and the batchable field", err.Error())
	}
}

// R4/R5: the app reads write actions through connectors.ManifestOf, and
// pm connectors describe reads them through Definition, so the declaration has
// to survive both projections.
func TestBatchablePropagatesToManifestAndDefinition(t *testing.T) {
	for _, tc := range []struct {
		name        string
		declaration string
		want        bool
	}{
		{name: "absent", declaration: "", want: true},
		{name: "explicit false", declaration: "false", want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := New(loadBundleWithBatchable(t, tc.declaration), nil)

			manifest := connectors.ManifestOf(c)
			if len(manifest.WriteActions) != 1 {
				t.Fatalf("manifest WriteActions = %d, want 1", len(manifest.WriteActions))
			}
			if got := manifest.WriteActions[0].IsBatchable(); got != tc.want {
				t.Fatalf("manifest IsBatchable() = %v, want %v", got, tc.want)
			}

			def, ok := connectors.DefinitionOf(c)
			if !ok {
				t.Fatal("DefinitionOf() returned false; engine connectors must provide a definition")
			}
			if len(def.WriteActions) != 1 {
				t.Fatalf("definition WriteActions = %d, want 1", len(def.WriteActions))
			}
			if got := def.WriteActions[0].IsBatchable(); got != tc.want {
				t.Fatalf("definition IsBatchable() = %v, want %v", got, tc.want)
			}
		})
	}
}

// R6: the manifest hands out a *bool. Sharing the loaded bundle's pointer would
// let any manifest consumer flip a live connector's policy.
func TestBatchableManifestDoesNotAliasBundlePointer(t *testing.T) {
	b := loadBundleWithBatchable(t, "false")
	c := New(b, nil)

	manifest := connectors.ManifestOf(c)
	spec := manifest.WriteActions[0]
	if spec.Batchable == nil {
		t.Fatal("manifest Batchable = nil, want an explicit false")
	}
	*spec.Batchable = true

	if b.Writes[0].IsBatchable() {
		t.Fatal("mutating the manifest spec flipped the loaded bundle; the pointer must be copied, not shared")
	}
	if connectors.ManifestOf(c).WriteActions[0].IsBatchable() {
		t.Fatal("mutating one manifest spec flipped the connector's manifest")
	}
}

// R3 originally guarded the transition where no shipped bundle declared the
// field. Google Calendar is the first intentional adopter, so lock the exact
// shipped policy: every other action remains batchable by default, and these
// safety-sensitive actions cannot silently become batchable.
func TestEveryShippedWriteActionHasExpectedBatchability(t *testing.T) {
	expectedNonBatchable := map[string]struct{}{
		"google-calendar/clear_calendar":              {},
		"google-calendar/delete_calendar":             {},
		"google-calendar/move_event":                  {},
		"google-calendar/quick_add_event":             {},
		"google-calendar/stop_channel":                {},
		"google-calendar/transfer_calendar_ownership": {},
		"google-calendar/watch_acl":                   {},
		"google-calendar/watch_calendar_list":         {},
		"google-calendar/watch_events":                {},
		"google-calendar/watch_settings":              {},
	}
	seenNonBatchable := make(map[string]struct{}, len(expectedNonBatchable))

	bundles, err := LoadAll(defs.FS)
	if err != nil {
		t.Fatalf("LoadAll(defs): %v", err)
	}
	for _, b := range bundles {
		for _, action := range b.Writes {
			key := b.Name + "/" + action.Name
			_, wantNonBatchable := expectedNonBatchable[key]
			wantBatchable := !wantNonBatchable
			if got := action.IsBatchable(); got != wantBatchable {
				t.Errorf("bundle %q action %q IsBatchable() = %v, want %v", b.Name, action.Name, got, wantBatchable)
			}
			if wantNonBatchable {
				seenNonBatchable[key] = struct{}{}
			}
		}
	}

	for key := range expectedNonBatchable {
		if _, ok := seenNonBatchable[key]; !ok {
			t.Errorf("expected non-batchable shipped action %q was not loaded", key)
		}
	}
}
