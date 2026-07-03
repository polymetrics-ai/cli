// Package faker_test parity-tests the Tier-3 native faker connector
// (internal/connectors/native/faker) against the legacy
// internal/connectors/faker package. Unlike postgres's fixture-mode-only
// parity (a live vs. fixture split), faker has no such split at all: legacy
// is already a pure, deterministic, network-free generator on every call, so
// parity here is checked directly against the SAME live behavior — there is
// no separate "fixture path" to reconcile.
//
//   - Check: both always succeed (subject to context cancellation).
//   - Catalog: identical stream NAME SET and PrimaryKey/CursorFields shape.
//   - Read: identical RECORD SET, field-for-field, for every stream across a
//     representative set of count/seed configs, including the edge cases
//     (unset stream defaults to "users", negative seed clamps to 0, products
//     always emits exactly 10 regardless of count).
//   - Definition(): smoke — name, capabilities.
package faker_test

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"polymetrics.ai/internal/connectors"
	legacy "polymetrics.ai/internal/connectors/faker"

	native "polymetrics.ai/internal/connectors/native/faker"
)

func readAll(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
	t.Helper()
	var got []connectors.Record
	if err := c.Read(context.Background(), req, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	}); err != nil {
		t.Fatalf("Read(%s): %v", req.Stream, err)
	}
	return got
}

func TestParityFaker_CatalogMatchesLegacy(t *testing.T) {
	nativeConn := native.New()
	legacyConn := legacy.New()

	nativeCat, err := nativeConn.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("native Catalog: %v", err)
	}
	legacyCat, err := legacyConn.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("legacy Catalog: %v", err)
	}

	nativeNames := streamNames(nativeCat.Streams)
	legacyNames := streamNames(legacyCat.Streams)
	if !reflect.DeepEqual(nativeNames, legacyNames) {
		t.Fatalf("stream name set mismatch: native=%v legacy=%v", nativeNames, legacyNames)
	}

	for _, name := range legacyNames {
		nStream := findStream(nativeCat.Streams, name)
		lStream := findStream(legacyCat.Streams, name)
		if !reflect.DeepEqual(nStream.PrimaryKey, lStream.PrimaryKey) {
			t.Errorf("stream %s PrimaryKey mismatch: native=%v legacy=%v", name, nStream.PrimaryKey, lStream.PrimaryKey)
		}
		if !reflect.DeepEqual(nStream.CursorFields, lStream.CursorFields) {
			t.Errorf("stream %s CursorFields mismatch: native=%v legacy=%v", name, nStream.CursorFields, lStream.CursorFields)
		}
	}
}

func TestParityFaker_ReadMatchesLegacyAcrossConfigs(t *testing.T) {
	nativeConn := native.New()
	legacyConn := legacy.New()

	cases := []struct {
		name   string
		stream string
		config map[string]string
	}{
		{"users default", "users", nil},
		{"users small count", "users", map[string]string{"count": "5"}},
		{"users with seed", "users", map[string]string{"count": "5", "seed": "41"}},
		{"users negative seed clamps", "users", map[string]string{"count": "3", "seed": "-100"}},
		{"purchases default", "purchases", nil},
		{"purchases small count", "purchases", map[string]string{"count": "12"}},
		{"purchases with seed", "purchases", map[string]string{"count": "4", "seed": "7"}},
		{"products ignores large count", "products", map[string]string{"count": "9999"}},
		{"empty stream defaults to users", "", map[string]string{"count": "2"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := connectors.RuntimeConfig{Config: tc.config}
			req := connectors.ReadRequest{Stream: tc.stream, Config: cfg}

			nativeGot := readAll(t, nativeConn, req)
			legacyGot := readAll(t, legacyConn, req)

			if len(nativeGot) != len(legacyGot) {
				t.Fatalf("record count mismatch: native=%d legacy=%d", len(nativeGot), len(legacyGot))
			}
			for i := range legacyGot {
				if !reflect.DeepEqual(nativeGot[i], legacyGot[i]) {
					t.Fatalf("record[%d] mismatch:\n native=%+v\n legacy=%+v", i, nativeGot[i], legacyGot[i])
				}
			}
		})
	}
}

func TestParityFaker_InvalidConfigRejectedBySides(t *testing.T) {
	nativeConn := native.New()
	legacyConn := legacy.New()

	cases := []struct {
		name   string
		stream string
		config map[string]string
	}{
		{"count zero", "users", map[string]string{"count": "0"}},
		{"count negative", "users", map[string]string{"count": "-3"}},
		{"count not a number", "users", map[string]string{"count": "abc"}},
		{"seed not a number", "users", map[string]string{"seed": "abc"}},
		{"unknown stream", "bogus", nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := connectors.RuntimeConfig{Config: tc.config}
			req := connectors.ReadRequest{Stream: tc.stream, Config: cfg}

			nativeErr := nativeConn.Read(context.Background(), req, func(connectors.Record) error { return nil })
			legacyErr := legacyConn.Read(context.Background(), req, func(connectors.Record) error { return nil })

			if nativeErr == nil {
				t.Fatalf("native Read: want error, got nil")
			}
			if legacyErr == nil {
				t.Fatalf("legacy Read: want error, got nil")
			}
		})
	}
}

func TestParityFaker_DefinitionSmoke(t *testing.T) {
	nativeConn := native.New()
	legacyConn := legacy.New()

	if nativeConn.Name() != legacyConn.Name() {
		t.Fatalf("Name mismatch: native=%q legacy=%q", nativeConn.Name(), legacyConn.Name())
	}

	nativeMeta := nativeConn.Metadata()
	legacyMeta := legacyConn.Metadata()
	if nativeMeta.Capabilities.Check != legacyMeta.Capabilities.Check ||
		nativeMeta.Capabilities.Catalog != legacyMeta.Capabilities.Catalog ||
		nativeMeta.Capabilities.Read != legacyMeta.Capabilities.Read ||
		nativeMeta.Capabilities.Write != legacyMeta.Capabilities.Write {
		t.Fatalf("Capabilities mismatch: native=%+v legacy=%+v", nativeMeta.Capabilities, legacyMeta.Capabilities)
	}
	if legacyMeta.Capabilities.Write {
		t.Fatal("legacy faker unexpectedly declares Write=true; parity assumption invalid")
	}
}

func TestParityFaker_CheckAlwaysSucceeds(t *testing.T) {
	nativeConn := native.New()
	legacyConn := legacy.New()

	if err := nativeConn.Check(context.Background(), connectors.RuntimeConfig{}); err != nil {
		t.Fatalf("native Check: %v", err)
	}
	if err := legacyConn.Check(context.Background(), connectors.RuntimeConfig{}); err != nil {
		t.Fatalf("legacy Check: %v", err)
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

func findStream(streams []connectors.Stream, name string) connectors.Stream {
	for _, s := range streams {
		if s.Name == name {
			return s
		}
	}
	return connectors.Stream{}
}
