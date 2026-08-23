package connectors

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestCDCArtifactManifestRestorationIsStrictAndBounded(t *testing.T) {
	manifest, err := NewCDCArtifactManifest(
		"connection",
		"public.users",
		1,
		"transaction-key",
		2,
		strings.Repeat("a", 64),
		strings.Repeat("b", 64),
		strings.Repeat("c", 64),
	)
	if err != nil {
		t.Fatalf("NewCDCArtifactManifest() error = %v", err)
	}
	receipt, err := NewCDCTransactionReceiptWithArtifactManifest("receipt", "warehouse:connection", time.Now().UTC(), manifest)
	if err != nil {
		t.Fatalf("NewCDCTransactionReceiptWithArtifactManifest() error = %v", err)
	}
	payload, err := receipt.ArtifactManifestJSON()
	if err != nil {
		t.Fatalf("ArtifactManifestJSON() error = %v", err)
	}

	restored, err := NewCDCTransactionReceiptWithArtifactManifestJSON("receipt", "warehouse:connection", time.Now().UTC(), payload)
	if err != nil {
		t.Fatalf("NewCDCTransactionReceiptWithArtifactManifestJSON(valid) error = %v", err)
	}
	got, err := restored.ArtifactManifest()
	if err != nil {
		t.Fatalf("ArtifactManifest() error = %v", err)
	}
	if got != manifest {
		t.Fatalf("ArtifactManifest() = %#v, want %#v", got, manifest)
	}

	for name, invalidPayload := range map[string]string{
		"missing":   "",
		"unknown":   strings.TrimSuffix(payload, "}") + `,"unrecognized":true}`,
		"trailing":  payload + "{}",
		"oversized": strings.Repeat(" ", (8<<10)+1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewCDCTransactionReceiptWithArtifactManifestJSON("receipt", "warehouse:connection", time.Now().UTC(), invalidPayload); err == nil {
				t.Fatalf("NewCDCTransactionReceiptWithArtifactManifestJSON(%s) error = nil, want strict private manifest refusal", name)
			}
		})
	}
}

type changefeedTestReader struct{}

func (changefeedTestReader) Name() string { return "changefeed-test" }

func (changefeedTestReader) Metadata() Metadata {
	return Metadata{
		Name:         "changefeed-test",
		Capabilities: Capabilities{CDC: true},
	}
}

func (changefeedTestReader) Check(context.Context, RuntimeConfig) error { return nil }

func (changefeedTestReader) Catalog(context.Context, RuntimeConfig) (Catalog, error) {
	return Catalog{}, nil
}

func (changefeedTestReader) Read(context.Context, ReadRequest, func(Record) error) error { return nil }

func (changefeedTestReader) Write(context.Context, WriteRequest, []Record) (WriteResult, error) {
	return WriteResult{}, nil
}

func (changefeedTestReader) ReadCDC(context.Context, CDCReadRequest, func(CDCEvent) error) error {
	return nil
}

type changefeedTestExecutor struct {
	changefeedTestReader
	descriptor ChangefeedExecutorDescriptor
}

func (c changefeedTestExecutor) ChangefeedExecutorDescriptor() ChangefeedExecutorDescriptor {
	return c.descriptor
}

type changefeedDefinitionExecutor struct {
	changefeedTestExecutor
	definition Definition
}

func (c changefeedDefinitionExecutor) Definition() Definition { return c.definition }

func TestHasImplementedChangefeedRequiresMatchingExecutor(t *testing.T) {
	declaration := &ChangefeedDescriptor{
		Status:    ChangefeedStatusImplemented,
		Mechanism: ChangefeedMechanismLogicalReplication,
		Source: ChangefeedSource{
			ArtifactURL:     "https://example.test/changefeed",
			ArtifactVersion: "v1",
			RetrievedAt:     "2026-08-05",
		},
		Executor: &ChangefeedExecutorRef{Kind: "native", ID: "acme-logical"},
		Checkpoint: &ChangefeedCheckpoint{
			Kind:        "lsn",
			Keys:        []string{"lsn"},
			CommitAfter: "downstream_ack",
			OnInvalid:   "resnapshot_required",
		},
		Delivery: &ChangefeedDelivery{
			Ordering:   "source_ordered",
			Duplicates: "at_least_once",
			Deletes:    "tombstone",
			DedupeKey:  []string{"source_id", "lsn"},
		},
		Streams: []string{"widgets"},
	}

	matching := ChangefeedExecutorDescriptor{
		Status:    ChangefeedStatusImplemented,
		Mechanism: ChangefeedMechanismLogicalReplication,
		Executor:  ChangefeedExecutorRef{Kind: "native", ID: "acme-logical"},
		Checkpoint: ChangefeedCheckpoint{
			Kind:        "lsn",
			Keys:        []string{"lsn"},
			CommitAfter: "downstream_ack",
			OnInvalid:   "resnapshot_required",
		},
	}

	cases := []struct {
		name      string
		connector Connector
		want      bool
	}{
		{
			name:      "legacy reader alone is insufficient",
			connector: changefeedTestReader{},
			want:      false,
		},
		{
			name: "wrong checkpoint does not match",
			connector: changefeedTestExecutor{descriptor: ChangefeedExecutorDescriptor{
				Status:    matching.Status,
				Mechanism: matching.Mechanism,
				Executor:  matching.Executor,
				Checkpoint: ChangefeedCheckpoint{
					Kind:        "cursor",
					Keys:        []string{"cursor"},
					CommitAfter: matching.Checkpoint.CommitAfter,
					OnInvalid:   matching.Checkpoint.OnInvalid,
				},
			}},
			want: false,
		},
		{
			name: "empty checkpoint commit_after does not match",
			connector: changefeedTestExecutor{descriptor: ChangefeedExecutorDescriptor{
				Status:    matching.Status,
				Mechanism: matching.Mechanism,
				Executor:  matching.Executor,
				Checkpoint: ChangefeedCheckpoint{
					Kind:      matching.Checkpoint.Kind,
					Keys:      matching.Checkpoint.Keys,
					OnInvalid: matching.Checkpoint.OnInvalid,
				},
			}},
			want: false,
		},
		{
			name: "empty checkpoint on_invalid does not match",
			connector: changefeedTestExecutor{descriptor: ChangefeedExecutorDescriptor{
				Status:    matching.Status,
				Mechanism: matching.Mechanism,
				Executor:  matching.Executor,
				Checkpoint: ChangefeedCheckpoint{
					Kind:        matching.Checkpoint.Kind,
					Keys:        matching.Checkpoint.Keys,
					CommitAfter: matching.Checkpoint.CommitAfter,
				},
			}},
			want: false,
		},
		{
			name:      "matching explicit executor is eligible",
			connector: changefeedTestExecutor{descriptor: matching},
			want:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := HasImplementedChangefeed(tc.connector, declaration); got != tc.want {
				t.Fatalf("HasImplementedChangefeed() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestChangefeedDescriptorRejectsUnsupportedExecutionClaims(t *testing.T) {
	base := ChangefeedDescriptor{
		Status:    ChangefeedStatusUnsupported,
		Mechanism: ChangefeedMechanismLogicalReplication,
		Source: ChangefeedSource{
			ArtifactURL:     "https://example.test/changefeed",
			ArtifactVersion: "v1",
			RetrievedAt:     "2026-08-05",
		},
		Reason: "not implemented",
	}

	for _, tc := range []struct {
		name   string
		mutate func(*ChangefeedDescriptor)
	}{
		{
			name: "executor",
			mutate: func(d *ChangefeedDescriptor) {
				d.Executor = &ChangefeedExecutorRef{Kind: "native", ID: "unsupported"}
			},
		},
		{
			name: "checkpoint",
			mutate: func(d *ChangefeedDescriptor) {
				d.Checkpoint = &ChangefeedCheckpoint{Kind: "lsn", Keys: []string{"lsn"}, CommitAfter: "downstream_ack", OnInvalid: "resnapshot_required"}
			},
		},
		{
			name: "delivery",
			mutate: func(d *ChangefeedDescriptor) {
				d.Delivery = &ChangefeedDelivery{Ordering: "source_ordered", Duplicates: "at_least_once", Deletes: "tombstone", DedupeKey: []string{"lsn"}}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			descriptor := base
			tc.mutate(&descriptor)
			if err := descriptor.Validate(); err == nil {
				t.Fatal("Validate() = nil, want unsupported descriptor execution claim rejected")
			}
		})
	}
}

func TestChangefeedDescriptorRejectsInvalidSourceURL(t *testing.T) {
	descriptor := ChangefeedDescriptor{
		Status:    ChangefeedStatusUnsupported,
		Mechanism: ChangefeedMechanismLogicalReplication,
		Source: ChangefeedSource{
			ArtifactURL:     "not-a-provider-artifact-url",
			ArtifactVersion: "v1",
			RetrievedAt:     "2026-08-05",
		},
		Reason: "not implemented",
	}
	if err := descriptor.Validate(); err == nil {
		t.Fatal("Validate() = nil, want invalid evidence URL rejected")
	}
}

func TestChangefeedDescriptorUsesClosedStatusAndMechanism(t *testing.T) {
	base := ChangefeedDescriptor{
		Status:    ChangefeedStatusUnsupported,
		Mechanism: ChangefeedMechanismLogicalReplication,
		Source: ChangefeedSource{
			ArtifactURL:     "https://example.test/changefeed",
			ArtifactVersion: "v1",
			RetrievedAt:     "2026-08-05",
		},
		Reason: "not implemented",
	}

	for _, tc := range []struct {
		name   string
		mutate func(*ChangefeedDescriptor)
	}{
		{
			name: "unknown status",
			mutate: func(d *ChangefeedDescriptor) {
				d.Status = "preview"
			},
		},
		{
			name: "unknown mechanism",
			mutate: func(d *ChangefeedDescriptor) {
				d.Mechanism = "snapshot_cursor"
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			descriptor := base
			tc.mutate(&descriptor)
			if err := descriptor.Validate(); err == nil {
				t.Fatal("Validate() = nil, want closed-vocabulary rejection")
			}
		})
	}
}

func TestDefinitionOfDerivesCDCFromMatchingChangefeedExecutor(t *testing.T) {
	declaration := &ChangefeedDescriptor{
		Status:    ChangefeedStatusImplemented,
		Mechanism: ChangefeedMechanismPollingWatermark,
		Source: ChangefeedSource{
			ArtifactURL:     "https://example.test/changes",
			ArtifactVersion: "v1",
			RetrievedAt:     "2026-08-05",
		},
		Executor: &ChangefeedExecutorRef{Kind: "engine", ID: "polling_watermark"},
		Checkpoint: &ChangefeedCheckpoint{
			Kind:        "watermark",
			Keys:        []string{"updated_at", "id"},
			CommitAfter: "downstream_ack",
			OnInvalid:   "resnapshot_required",
		},
		Delivery: &ChangefeedDelivery{
			Ordering:   "provider_declared_stable_order",
			Duplicates: "at_least_once",
			Deletes:    "not_available",
			DedupeKey:  []string{"id", "updated_at"},
		},
		Streams: []string{"widgets"},
		PollingWatermark: &PollingWatermarkSpec{
			Watermark:        PollingWatermarkValue{Kind: "timestamp", Path: "updated_at"},
			TieBreaker:       PollingWatermarkField{Path: "id"},
			Boundary:         "inclusive",
			SafetyLagSeconds: 0,
			PageSize:         2,
			MaxPages:         1,
			RequestBudget:    1,
		},
	}
	connector := changefeedDefinitionExecutor{
		changefeedTestExecutor: changefeedTestExecutor{descriptor: ChangefeedExecutorDescriptor{
			Status:    ChangefeedStatusImplemented,
			Mechanism: ChangefeedMechanismPollingWatermark,
			Executor:  ChangefeedExecutorRef{Kind: "engine", ID: "polling_watermark"},
			Checkpoint: ChangefeedCheckpoint{
				Kind:        "watermark",
				Keys:        []string{"updated_at", "id"},
				CommitAfter: "downstream_ack",
				OnInvalid:   "resnapshot_required",
			},
		}},
		definition: Definition{
			Name:         "changefeed-test",
			Capabilities: Capabilities{CDC: false},
			Changefeed:   declaration,
		},
	}

	definition, ok := DefinitionOf(connector)
	if !ok {
		t.Fatal("DefinitionOf() = false, want true")
	}
	if !definition.Capabilities.CDC {
		t.Fatalf("DefinitionOf().Capabilities.CDC = false, want true: %+v", definition)
	}
	if !MetadataOf(connector).Capabilities.CDC {
		t.Fatal("MetadataOf().Capabilities.CDC = false, want true")
	}
}

func TestRegistryDoesNotTrustLegacyCDCMetadataOrReader(t *testing.T) {
	registry := NewEmptyRegistry()
	registry.Register(changefeedTestReader{})

	listed := registry.List()
	if len(listed) != 1 {
		t.Fatalf("List() returned %d connectors, want 1", len(listed))
	}
	if listed[0].Capabilities.CDC {
		t.Fatalf("List() advertised CDC from legacy metadata: %+v", listed[0].Capabilities)
	}
	if manifest := ManifestOf(changefeedTestReader{}); manifest.Metadata.Capabilities.CDC {
		t.Fatalf("ManifestOf() advertised CDC from legacy metadata: %+v", manifest.Metadata.Capabilities)
	}

	catalog := registry.CatalogEntries()
	if len(catalog) != 1 {
		t.Fatalf("CatalogEntries() returned %d connectors, want 1", len(catalog))
	}
	if catalog[0].Capabilities.CDC {
		t.Fatalf("CatalogEntries() advertised CDC from legacy metadata/CDCReader: %+v", catalog[0].Capabilities)
	}
}
