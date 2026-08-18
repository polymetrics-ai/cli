package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pglogrepl"

	"polymetrics.ai/internal/synccontract"
)

func TestPostgresBootstrapCheckpointBindsBarrierAndSchemaFingerprint(t *testing.T) {
	source := postgresCDCSource{
		identity: synccontract.SourceIdentity{
			Engine:           "postgres",
			AccountOrCluster: "system-one:analytics",
			ObjectScope:      "public.events",
		},
		generation:        synccontract.OpaqueToken("7\npm_events"),
		publication:       "pm_events",
		schemaFingerprint: "schema-fingerprint-one",
		system: pglogrepl.IdentifySystemResult{
			SystemID: "system-one",
			DBName:   "analytics",
			Timeline: 7,
		},
	}
	barrier := pglogrepl.LSN(0x16B6C50)

	bootstrap, err := newPostgresBootstrapBarrier(source, barrier)
	if err != nil {
		t.Fatalf("newPostgresBootstrapBarrier() error = %v", err)
	}
	source.bootstrap = &bootstrap
	candidate := postgresCDCCheckpointForLSNs(source, barrier, barrier, barrier, barrier)
	if err := candidate.Validate(); err != nil {
		t.Fatalf("bootstrap checkpoint Validate() = %v", err)
	}
	if got := string(candidate.SnapshotBarrier.Token); got == barrier.String() {
		t.Fatal("bootstrap checkpoint persisted only a bare LSN, losing source/schema binding")
	}

	decoded, ok, err := postgresBootstrapBarrierFromCheckpoint(&candidate)
	if err != nil || !ok {
		t.Fatalf("postgresBootstrapBarrierFromCheckpoint() = (%#v, %t, %v), want complete metadata", decoded, ok, err)
	}
	if decoded.InitialLSN != barrier.String() || decoded.SystemID != source.system.SystemID || decoded.Timeline != source.system.Timeline || decoded.Publication != source.publication || decoded.Relation != source.identity.ObjectScope || decoded.SchemaFingerprint != source.schemaFingerprint {
		t.Fatalf("bootstrap metadata = %#v, want barrier/source/schema binding", decoded)
	}

	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement("bootstrap-test", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	var committed synccontract.CheckpointEnvelope
	if err := synccontract.CommitAfterDownstreamAcknowledgement(candidate, acknowledgement, func(checkpoint synccontract.CheckpointEnvelope) error {
		committed = checkpoint
		return nil
	}); err != nil {
		t.Fatalf("CommitAfterDownstreamAcknowledgement() = %v", err)
	}
	if err := validateCDCResume(&committed, source); err != nil {
		t.Fatalf("validateCDCResume() = %v, want matching bootstrap state accepted", err)
	}

	for _, test := range []struct {
		name    string
		mutate  func(*postgresCDCSource)
		outcome synccontract.RecoveryOutcome
	}{
		{
			name: "schema",
			mutate: func(drifted *postgresCDCSource) {
				drifted.schemaFingerprint = "schema-fingerprint-two"
			},
			outcome: synccontract.RecoveryOutcomeSourceGenerationChanged,
		},
		{
			name: "publication",
			mutate: func(drifted *postgresCDCSource) {
				drifted.publication = "pm_events_two"
				drifted.generation = synccontract.OpaqueToken("7\npm_events_two")
			},
			outcome: synccontract.RecoveryOutcomeSourceGenerationChanged,
		},
		{
			name: "timeline",
			mutate: func(drifted *postgresCDCSource) {
				drifted.system.Timeline = 8
				drifted.generation = synccontract.OpaqueToken("8\npm_events")
			},
			outcome: synccontract.RecoveryOutcomeSourceGenerationChanged,
		},
		{
			name: "system",
			mutate: func(drifted *postgresCDCSource) {
				drifted.system.SystemID = "system-two"
				drifted.identity.AccountOrCluster = "system-two:analytics"
			},
			outcome: synccontract.RecoveryOutcomeSourceIdentityIncompatible,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			drifted := source
			test.mutate(&drifted)
			err := validateCDCResume(&committed, drifted)
			if !errors.Is(err, synccontract.ErrRebootstrapRequired) {
				t.Fatalf("validateCDCResume(%s drift) = %v, want rebootstrap requirement", test.name, err)
			}
			var recovery *synccontract.RebootstrapRequiredError
			if !errors.As(err, &recovery) || recovery.Outcome != test.outcome {
				t.Fatalf("validateCDCResume(%s drift) recovery = %#v, want %q", test.name, recovery, test.outcome)
			}
		})
	}
}
