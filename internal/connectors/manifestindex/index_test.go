package manifestindex

import (
	"errors"
	"testing"
)

func TestIndexIsSortedUniqueAndBounded(t *testing.T) {
	index, err := New([]Entry{
		{Connector: "z", Generation: "g", Digest: "d", Executor: "api_engine.v1", Bytes: 1},
		{Connector: "a", Generation: "g", Digest: "d", Executor: "api_engine.v1", Bytes: 1},
	}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got := index.List(); got[0].Connector != "a" {
		t.Fatalf("list = %#v", got)
	}
	if got, ok := index.Lookup("z"); !ok || got.Executor != "api_engine.v1" {
		t.Fatalf("lookup = %#v/%t", got, ok)
	}
	if _, err := New([]Entry{
		{Connector: "a", Generation: "g", Digest: "d", Executor: "api_engine.v1", Bytes: 1},
		{Connector: "a", Generation: "g", Digest: "d", Executor: "api_engine.v1", Bytes: 1},
	}, 2); err == nil {
		t.Fatal("duplicate accepted")
	}
	if _, err := New([]Entry{{Connector: "a", Generation: "g", Digest: "d", Executor: "api_engine.v1", Bytes: 1}}, 0); err == nil {
		t.Fatal("limit accepted")
	}
}

func TestIndexRejectsExecutorOutsideClosedVocabulary(t *testing.T) {
	_, err := New([]Entry{{
		Connector:  "alpha",
		Generation: "generation-1",
		Digest:     "digest-1",
		Executor:   "arbitrary_executor.v1",
		Bytes:      1,
	}}, 1)
	if err == nil {
		t.Fatal("New accepted an executor outside the closed vocabulary")
	}
	var typed interface{ UnknownExecutor() string }
	if !errors.As(err, &typed) || typed.UnknownExecutor() != "arbitrary_executor.v1" {
		t.Fatalf("New error = %v, want typed unknown-executor error", err)
	}
}

func TestIndexRejectsMissingByteCharge(t *testing.T) {
	_, err := New([]Entry{{
		Connector:  "alpha",
		Generation: "generation-1",
		Digest:     "digest-1",
		Executor:   "api_engine.v1",
	}}, 1)
	if err == nil {
		t.Fatal("New accepted a manifest entry without a bounded byte charge")
	}
}

func TestGeneratedEntriesAreClosedRuntimeProjection(t *testing.T) {
	entries := GeneratedEntries()
	if len(entries) != 553 {
		t.Fatalf("generated entry count = %d, want 553", len(entries))
	}
	index, err := New(entries, len(entries))
	if err != nil {
		t.Fatalf("New(GeneratedEntries()) error = %v", err)
	}
	for connector, wantExecutor := range map[string]string{
		"dynamodb":    "native_database/dynamodb.v1",
		"mysql":       "native_database/mysql.v1",
		"github":      "api_engine.v1",
		"postgres":    "native_database/postgres.v1",
		"bing-ads":    "closed_typed/bing-ads.v1",
		"tally-prime": "closed_typed/tally-prime.v1",
	} {
		entry, ok := index.Lookup(connector)
		if !ok {
			t.Fatalf("generated index is missing %q", connector)
		}
		if entry.Executor != wantExecutor {
			t.Fatalf("generated executor for %q = %q, want %q", connector, entry.Executor, wantExecutor)
		}
		if entry.Generation == "" || entry.Digest == "" || entry.Bytes <= 0 {
			t.Fatalf("generated entry for %q has incomplete identity: %#v", connector, entry)
		}
	}
}

func TestGeneratedMetadataPreservesCatalogProjection(t *testing.T) {
	index, err := New(GeneratedEntries(), len(GeneratedEntries()))
	if err != nil {
		t.Fatal(err)
	}
	github, ok := index.Lookup("github")
	if !ok || !github.Metadata.Capabilities.Catalog {
		t.Fatalf("github generated metadata = %#v, want catalog capability", github.Metadata)
	}
	postgres, ok := index.Lookup("postgres")
	if !ok || !postgres.Metadata.Capabilities.Catalog || !postgres.Metadata.Capabilities.CDC {
		t.Fatalf("postgres generated metadata = %#v, want catalog and native CDC capabilities", postgres.Metadata)
	}
}
