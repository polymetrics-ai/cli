package warehouse_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/warehouse"
)

func TestSegmentManifestV1PublishesAndReopensImmutableSegment(t *testing.T) {
	ctx := context.Background()
	location := segmentManifestLocation(t)
	manifest, err := warehouse.NewSegmentManifestV1(warehouse.SegmentManifestV1{
		ID: "segment-001", SchemaHash: segmentHash("schema"), TransformPlanHash: segmentHash("transform"),
		SourceLogicalBytes: 2048, SourceRows: 8, TransformedRows: 7, TransformedBytes: 1792,
		ParquetFile: "segment-001.parquet", ParquetBytes: 512, ParquetSHA256: segmentHash("parquet"),
	})
	if err != nil {
		t.Fatalf("NewSegmentManifestV1() error = %v", err)
	}
	path, err := warehouse.WriteSegmentManifest(ctx, location, "events", manifest)
	if err != nil {
		t.Fatalf("WriteSegmentManifest() error = %v", err)
	}
	if filepath.Base(path) != "manifest-segment-001.json" {
		t.Fatalf("manifest path = %q", path)
	}
	reopened, err := warehouse.OpenSegmentManifest(ctx, location, "events", "segment-001")
	if err != nil {
		t.Fatalf("OpenSegmentManifest() error = %v", err)
	}
	if reopened != manifest {
		t.Fatalf("reopened manifest = %#v, want %#v", reopened, manifest)
	}
}

func TestSegmentManifestV1RefusesInvalidIdentityBeforePublishing(t *testing.T) {
	_, err := warehouse.NewSegmentManifestV1(warehouse.SegmentManifestV1{
		ID: "../escape", SchemaHash: segmentHash("schema"), ParquetFile: "segment.parquet", ParquetSHA256: segmentHash("parquet"),
	})
	if !errors.Is(err, warehouse.ErrTransportSegmentManifestInvalid) {
		t.Fatalf("NewSegmentManifestV1(unsafe ID) error = %T %v, want ErrTransportSegmentManifestInvalid", err, err)
	}
}

func TestSegmentManifestV1RejectsReplayAndCorruption(t *testing.T) {
	ctx := context.Background()
	location := segmentManifestLocation(t)
	manifest, err := warehouse.NewSegmentManifestV1(warehouse.SegmentManifestV1{
		ID: "segment-replay", SchemaHash: segmentHash("schema"), ParquetFile: "segment-replay.parquet", ParquetSHA256: segmentHash("parquet"),
	})
	if err != nil {
		t.Fatal(err)
	}
	path, err := warehouse.WriteSegmentManifest(ctx, location, "events", manifest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := warehouse.WriteSegmentManifest(ctx, location, "events", manifest); !errors.Is(err, warehouse.ErrTransportSegmentManifestInvalid) {
		t.Fatalf("WriteSegmentManifest(replay) error = %T %v, want refusal", err, err)
	}
	if err := os.WriteFile(path, []byte(`{"version":1,"id":"segment-replay"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := warehouse.OpenSegmentManifest(ctx, location, "events", "segment-replay"); !errors.Is(err, warehouse.ErrTransportSegmentManifestInvalid) {
		t.Fatalf("OpenSegmentManifest(corrupt) error = %T %v, want refusal", err, err)
	}
}

func segmentManifestLocation(t *testing.T) warehouse.Location {
	t.Helper()
	location, err := warehouse.LocationFor(t.TempDir(), "workspace", "postgres", "conn-1", "Segment test")
	if err != nil {
		t.Fatal(err)
	}
	return location
}

func segmentHash(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
