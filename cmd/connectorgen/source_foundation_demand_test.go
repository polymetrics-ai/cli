package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func sourceFoundationUniverseFixture(t *testing.T) (string, sourceLaneCohort) {
	t.Helper()
	repo, cohort := sourceInventoryFixture(t, []string{"source.b", "source.a"}, 2)
	// Use the actual Atlas rather than a synthetic owner string. This initial
	// projection loads its IDs/owners only; declaration/proof admission is later.
	project, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(project, sourceDemandAtlasPath))
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(repo, sourceDemandAtlasPath)
	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return repo, cohort
}

func TestSourceFoundationUniverseProjection(t *testing.T) {
	repo, cohort := sourceFoundationUniverseFixture(t)
	before, err := os.ReadFile(filepath.Join(repo, "source.json"))
	if err != nil {
		t.Fatal(err)
	}
	atlasBefore, err := os.ReadFile(filepath.Join(repo, sourceDemandAtlasPath))
	if err != nil {
		t.Fatal(err)
	}
	expectedManifest, err := buildSourceLaneManifest(t.Context(), repo, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := buildSourceFoundationUniverse(t.Context(), repo, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	// The expected identities and lane vocabulary are independent of the new
	// projection and sourceLaneNames. Both source rows must retain all lanes.
	want := []sourceFoundationCell{}
	for _, id := range []string{"source.a", "source.b"} {
		for _, lane := range []string{"direct_read", "direct_write", "binary_download", "binary_upload", "etl", "reverse_etl", "sync_transport"} {
			want = append(want, sourceFoundationCell{Key: sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: id}, Lane: lane})
		}
	}
	if !reflect.DeepEqual(got.cells, want) {
		t.Fatalf("source universe = %#v; want fourteen exact independently named cells", got.cells)
	}
	if !sourceLaneJSONEqual(got.manifest, expectedManifest) {
		t.Fatal("foundation projection changed original source/lane facts")
	}
	if got.atlasOwners["transport.sync-contract.v1"] != "polymetrics.ai/internal/synctransport" || got.atlasPin != (sourceArtifactPin{Path: sourceDemandAtlasPath, SHA256: sourceBytesHash(atlasBefore), Bytes: int64(len(atlasBefore))}) {
		t.Fatal("actual Atlas owner or complete input pin lost")
	}
	for _, row := range got.manifest.SourceOperations {
		for _, cell := range row.Lanes {
			if cell.State == "implemented" || len(cell.ProofRefs) != 0 {
				t.Fatal("non-authorizing projection promoted fixture capability")
			}
		}
	}
	for name, expected := range map[string][]byte{"source.json": before, sourceDemandAtlasPath: atlasBefore} {
		after, err := os.ReadFile(filepath.Join(repo, name))
		if err != nil || !bytes.Equal(after, expected) {
			t.Fatalf("read-only projection altered %s: %v", name, err)
		}
	}
}

func TestSourceFoundationUniverseInvalidInputs(t *testing.T) {
	for _, scenario := range []string{"retained replacement", "missing atlas", "cancelled"} {
		t.Run(scenario, func(t *testing.T) {
			repo, cohort := sourceFoundationUniverseFixture(t)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			switch scenario {
			case "retained replacement":
				if err := os.WriteFile(filepath.Join(repo, "source.json"), []byte(`{"readable":"unrelated"}`), 0600); err != nil {
					t.Fatal(err)
				}
			case "missing atlas":
				if err := os.Remove(filepath.Join(repo, sourceDemandAtlasPath)); err != nil {
					t.Fatal(err)
				}
			case "cancelled":
				cancel()
			}
			got, err := buildSourceFoundationUniverse(ctx, repo, cohort, nil)
			if err == nil || len(got.cells) != 0 || len(got.atlasOwners) != 0 {
				t.Fatalf("invalid input exposed admitted universe: %v", err)
			}
			if scenario == "cancelled" && !errors.Is(err, context.Canceled) {
				t.Fatalf("lost cancellation cause: %v", err)
			}
		})
	}
}
