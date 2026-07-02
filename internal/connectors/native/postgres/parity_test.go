// Package postgres_test parity-tests the Tier-3 native postgres connector
// (internal/connectors/native/postgres) against the legacy
// internal/connectors/postgres package, per PLAN.md T-17/B-17: this is the
// golden migration reference for every future database/file/native
// connector migration. Parity is checked in fixture mode only (cfg
// mode=fixture short-circuits all network access on both sides), across:
//
//   - Check: same accept/reject outcome for every resolveConfig validation
//     rule (host/port/sslmode/required fields), exact-match on rejection
//     (both sides return non-nil, no live network dial happens in fixture
//     mode) but SEMANTIC-match only on error wording — this package does not
//     assert the legacy Go error string verbatim, since the design doc
//     explicitly ports "logic" not "byte-identical error text" (documented
//     choice, see traces/waveF-b17-ledger.md "parity choices").
//   - Catalog: identical stream NAME SET (fixture streams are a fixed canned
//     catalog on both sides).
//   - Read: identical RECORD SET (by primary key) for the first fixture
//     stream, including incremental cursor-lower-bound filtering.
//   - Definition(): smoke — name, capabilities, spec fields (host/port/
//     database/username/password/sslmode, x-secret on password).
package postgres_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"polymetrics.ai/internal/connectors"
	legacy "polymetrics.ai/internal/connectors/postgres"

	native "polymetrics.ai/internal/connectors/native/postgres"
)

// --- shared fixture-mode config (mirrors legacy postgres_test.go) ---------

func parityFixtureConfig() connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"mode":     "fixture",
			"host":     "db.internal",
			"database": "analytics",
			"username": "reader",
			"sslmode":  "require",
		},
		Secrets: map[string]string{"password": "s3cret"},
	}
}

// invalidConfigTable enumerates every resolveConfig validation rule
// (legacy postgres.go:119 resolveConfig): each case must be REJECTED
// (non-nil error) by BOTH the legacy and native connectors' Check, and the
// rejection CLASS (which rule fired) must match even though wording may
// differ — see classifyConfigError below and the ledger note on exact-match
// vs semantic-match.
func invalidConfigTable() []struct {
	name  string
	class string
	cfg   connectors.RuntimeConfig
} {
	return []struct {
		name  string
		class string
		cfg   connectors.RuntimeConfig
	}{
		{
			name:  "missing host",
			class: "missing_host",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"database": "d", "username": "u"},
				Secrets: map[string]string{"password": "p"},
			},
		},
		{
			name:  "missing database",
			class: "missing_database",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"host": "h", "username": "u"},
				Secrets: map[string]string{"password": "p"},
			},
		},
		{
			name:  "missing username",
			class: "missing_username",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"host": "h", "database": "d"},
				Secrets: map[string]string{"password": "p"},
			},
		},
		{
			name:  "missing password secret",
			class: "missing_password",
			cfg: connectors.RuntimeConfig{
				Config: map[string]string{"host": "h", "database": "d", "username": "u", "sslmode": "require"},
			},
		},
		{
			name:  "invalid sslmode",
			class: "invalid_sslmode",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"host": "h", "database": "d", "username": "u", "sslmode": "bananas"},
				Secrets: map[string]string{"password": "p"},
			},
		},
		{
			name:  "invalid port (non-numeric)",
			class: "invalid_port",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"host": "h", "database": "d", "username": "u", "port": "not-a-number"},
				Secrets: map[string]string{"password": "p"},
			},
		},
		{
			name:  "invalid port (out of range)",
			class: "invalid_port",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"host": "h", "database": "d", "username": "u", "port": "70000"},
				Secrets: map[string]string{"password": "p"},
			},
		},
		{
			name:  "host with scheme (SSRF guard)",
			class: "invalid_host",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"host": "http://evil.example.com", "database": "d", "username": "u"},
				Secrets: map[string]string{"password": "p"},
			},
		},
		{
			name:  "host with path/query characters",
			class: "invalid_host",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"host": "evil.example.com/../x", "database": "d", "username": "u"},
				Secrets: map[string]string{"password": "p"},
			},
		},
	}
}

// TestParityCheckFixtureModeAccepts asserts both connectors accept the same
// valid fixture-mode config.
func TestParityCheckFixtureModeAccepts(t *testing.T) {
	cfg := parityFixtureConfig()

	lc := legacy.New()
	if err := lc.Check(context.Background(), cfg); err != nil {
		t.Fatalf("legacy Check(valid fixture cfg) = %v, want nil", err)
	}

	nc := native.New()
	if err := nc.Check(context.Background(), cfg); err != nil {
		t.Fatalf("native Check(valid fixture cfg) = %v, want nil", err)
	}
}

// TestParityConfigValidationErrorTable drives invalidConfigTable() against
// BOTH connectors' Check and asserts they reject with the SAME classified
// rule (semantic-match on classification, not exact string match on
// wording — documented choice in the ledger).
func TestParityConfigValidationErrorTable(t *testing.T) {
	lc := legacy.New()
	nc := native.New()

	for _, tc := range invalidConfigTable() {
		t.Run(tc.name, func(t *testing.T) {
			legacyErr := lc.Check(context.Background(), tc.cfg)
			if legacyErr == nil {
				t.Fatalf("legacy Check(%s) = nil, want a validation error (test table itself only encodes real legacy rules)", tc.name)
			}
			nativeErr := nc.Check(context.Background(), tc.cfg)
			if nativeErr == nil {
				t.Fatalf("native Check(%s) = nil, want a validation error", tc.name)
			}

			legacyClass := classifyConfigError(legacyErr.Error())
			nativeClass := classifyConfigError(nativeErr.Error())
			if legacyClass != tc.class {
				t.Fatalf("legacy Check(%s) error %q classified as %q, want %q (test table needs updating, or legacy rule drifted)", tc.name, legacyErr, legacyClass, tc.class)
			}
			if nativeClass != tc.class {
				t.Fatalf("native Check(%s) error %q classified as %q, want %q (parity break vs legacy classification %q)", tc.name, nativeErr, nativeClass, tc.class, legacyClass)
			}
		})
	}
}

// classifyConfigError maps a resolveConfig-family error message to the rule
// that produced it. This mirrors the classification approach the
// coordinator's own conformance/connectorgen packages use for loader errors
// (classifyLoadError) — kept local since PLAN.md forbids cross-package
// sharing between independently-owned parity corpora.
func classifyConfigError(msg string) string {
	contains := func(sub string) bool {
		return len(msg) >= len(sub) && indexOf(msg, sub) >= 0
	}
	switch {
	case contains("host") && (contains("URL") || contains("scheme") || contains("bare hostname")):
		return "invalid_host"
	case contains("requires config host") || (contains("host") && contains("required")):
		return "missing_host"
	case contains("requires config database") || (contains("database") && contains("required")):
		return "missing_database"
	case contains("requires config username") || (contains("username") && contains("required")):
		return "missing_username"
	case contains("requires secret password") || (contains("password") && contains("required")):
		return "missing_password"
	case contains("sslmode"):
		return "invalid_sslmode"
	case contains("port"):
		return "invalid_port"
	default:
		return "unknown"
	}
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}

// TestParityCatalogStreamSet asserts both connectors' fixture-mode Catalog
// return the same set of stream names.
func TestParityCatalogStreamSet(t *testing.T) {
	cfg := parityFixtureConfig()

	lc := legacy.New()
	legacyCat, err := lc.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("legacy Catalog: %v", err)
	}

	nc := native.New()
	nativeCat, err := nc.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("native Catalog: %v", err)
	}

	legacyNames := streamNames(legacyCat.Streams)
	nativeNames := streamNames(nativeCat.Streams)
	if !reflect.DeepEqual(legacyNames, nativeNames) {
		t.Fatalf("stream name set mismatch:\n legacy=%v\n native=%v", legacyNames, nativeNames)
	}
	if len(legacyNames) == 0 {
		t.Fatal("expected at least one fixture stream")
	}
}

func streamNames(streams []connectors.Stream) []string {
	out := make([]string, 0, len(streams))
	for _, s := range streams {
		out = append(out, s.Name)
	}
	sort.Strings(out)
	return out
}

// TestParityReadRecordEquality asserts that, for the SAME fixture stream and
// identical config, the legacy and native connectors emit the same set of
// records (keyed by primary key) — full snapshot AND incremental-cursor
// filtered.
func TestParityReadRecordEquality(t *testing.T) {
	cfg := parityFixtureConfig()

	lc := legacy.New()
	legacyCat, err := lc.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("legacy Catalog: %v", err)
	}
	nc := native.New()
	nativeCat, err := nc.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("native Catalog: %v", err)
	}
	if len(legacyCat.Streams) == 0 || len(nativeCat.Streams) == 0 {
		t.Fatal("expected non-empty fixture catalogs on both sides")
	}

	for _, stream := range streamNames(legacyCat.Streams) {
		t.Run(stream, func(t *testing.T) {
			legacyRecords := readAll(t, lc, stream, cfg, nil)
			nativeRecords := readAll(t, nc, stream, cfg, nil)
			assertRecordSetsEqual(t, "full snapshot", legacyRecords, nativeRecords)

			// Incremental: seed a cursor state past the first row on both
			// sides and assert the resulting filtered sets still match.
			state := map[string]string{"cursor": "1000"}
			legacyIncremental := readAll(t, lc, stream, cfg, state)
			nativeIncremental := readAll(t, nc, stream, cfg, state)
			assertRecordSetsEqual(t, "incremental cursor=1000", legacyIncremental, nativeIncremental)
		})
	}
}

func readAll(t *testing.T, c connectors.Connector, stream string, cfg connectors.RuntimeConfig, state map[string]string) []connectors.Record {
	t.Helper()
	var out []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg, State: state}, func(r connectors.Record) error {
		out = append(out, r)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(%s): %v", stream, err)
	}
	return out
}

// assertRecordSetsEqual compares two record slices by content, independent
// of emission order (fixture read order is deterministic today on both
// sides, but the parity contract is about record equality, not iteration
// order).
func assertRecordSetsEqual(t *testing.T, label string, got, want []connectors.Record) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: record count = %d, want %d\n got=%+v\n want=%+v", label, len(got), len(want), got, want)
	}
	gotSet := recordSet(got)
	wantSet := recordSet(want)
	if !reflect.DeepEqual(gotSet, wantSet) {
		t.Fatalf("%s: record sets differ:\n got=%+v\n want=%+v", label, gotSet, wantSet)
	}
}

// recordSet renders records to a comparable, order-independent form: sorted
// stringified rows (values compared with fmt-equivalent stringification via
// reflect.DeepEqual on a normalized map is order-sensitive on nothing but
// map key order, which Go's map type already ignores for DeepEqual).
func recordSet(records []connectors.Record) []connectors.Record {
	out := make([]connectors.Record, len(records))
	copy(out, records)
	sort.Slice(out, func(i, j int) bool {
		return recordSortKey(out[i]) < recordSortKey(out[j])
	})
	return out
}

func recordSortKey(r connectors.Record) string {
	if id, ok := r["id"]; ok {
		return fmt.Sprintf("%v", id)
	}
	return fmt.Sprintf("%v", r)
}

// --- Definition() smoke ---------------------------------------------------

func TestDefinitionServedFromBundle(t *testing.T) {
	nc := native.New()
	provider, ok := any(nc).(connectors.DefinitionProvider)
	if !ok {
		t.Fatal("native postgres connector must implement connectors.DefinitionProvider (engine.Base)")
	}
	def := provider.Definition()

	if def.Name != "postgres" {
		t.Fatalf("Definition().Name = %q, want postgres", def.Name)
	}
	if def.IntegrationType != "database" {
		t.Fatalf("Definition().IntegrationType = %q, want database", def.IntegrationType)
	}
	if !def.Capabilities.Check || !def.Capabilities.Read {
		t.Fatalf("Definition().Capabilities = %+v, want Check && Read", def.Capabilities)
	}
	if def.Capabilities.Write {
		t.Fatalf("Definition().Capabilities.Write = true, want false (read-only source, wave0 parity)")
	}

	var spec map[string]any
	if err := json.Unmarshal(def.Spec, &spec); err != nil {
		t.Fatalf("Definition().Spec did not decode as JSON: %v", err)
	}
	props, ok := spec["properties"].(map[string]any)
	if !ok {
		t.Fatalf("Definition().Spec has no properties object: %v", spec)
	}
	for _, want := range []string{"host", "port", "database", "username", "password", "sslmode"} {
		if _, ok := props[want]; !ok {
			t.Fatalf("Definition().Spec.properties missing %q: %v", want, props)
		}
	}
	passwordProp, _ := props["password"].(map[string]any)
	if passwordProp == nil || passwordProp["x-secret"] != true {
		t.Fatalf("Definition().Spec.properties.password missing x-secret:true: %v", passwordProp)
	}
}
