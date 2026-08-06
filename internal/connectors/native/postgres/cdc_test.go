package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgconn"

	"polymetrics.ai/internal/synccontract"
)

func TestCDCSlotNameIsStableAndSourceBound(t *testing.T) {
	first := synccontract.SourceIdentity{
		Engine:           "postgres",
		AccountOrCluster: "system-one:database-one",
		ObjectScope:      "public.events",
	}
	otherSource := first
	otherSource.AccountOrCluster = "system-two:database-two"
	otherStream := first
	otherStream.ObjectScope = "audit.events"

	name := cdcSlotName(first)
	if name != cdcSlotName(first) {
		t.Fatal("derived CDC slot name is not stable")
	}
	if name == cdcSlotName(otherSource) || name == cdcSlotName(otherStream) {
		t.Fatal("derived CDC slot name is not source-bound")
	}
	if len(name) > 63 || len(name) <= len(cdcSlotPrefix) {
		t.Fatal("derived CDC slot name is not a valid PostgreSQL identifier length")
	}
}

func TestCanonicalCDCStreamBindsDefaultSchema(t *testing.T) {
	stream, err := canonicalCDCStream("audit", "events")
	if err != nil {
		t.Fatalf("canonicalCDCStream: %v", err)
	}
	if stream != "audit.events" {
		t.Fatalf("canonicalCDCStream = %q, want audit.events", stream)
	}
	if _, err := canonicalCDCStream("public", "events.bad.extra"); err == nil {
		t.Fatal("canonicalCDCStream accepted a non-relation stream")
	}
}

func TestCDCResumeRejectsDifferentSource(t *testing.T) {
	source := postgresCDCSource{
		identity: synccontract.SourceIdentity{
			Engine:           "postgres",
			AccountOrCluster: "system-one:database-one",
			ObjectScope:      "public.events",
		},
		generation: synccontract.OpaqueToken([]byte("timeline-one")),
	}
	candidate := postgresCDCCheckpoint(source, pglogrepl.LSN(16), 0, &pglogrepl.CommitMessage{
		CommitLSN:         pglogrepl.LSN(24),
		TransactionEndLSN: pglogrepl.LSN(32),
	})
	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement("cdc-test-sink", time.Now().UTC())
	if err != nil {
		t.Fatalf("NewDurableDownstreamAcknowledgement: %v", err)
	}
	var checkpoint synccontract.CheckpointEnvelope
	if err := synccontract.CommitAfterDownstreamAcknowledgement(candidate, acknowledgement, func(committed synccontract.CheckpointEnvelope) error {
		checkpoint = committed
		return nil
	}); err != nil {
		t.Fatalf("CommitAfterDownstreamAcknowledgement: %v", err)
	}
	if err := validateCDCResume(&checkpoint, source); err != nil {
		t.Fatalf("validateCDCResume(same source): %v", err)
	}

	other := source
	other.identity.AccountOrCluster = "system-two:database-two"
	err = validateCDCResume(&checkpoint, other)
	if !errors.Is(err, synccontract.ErrRebootstrapRequired) {
		t.Fatalf("validateCDCResume(different source) = %v, want rebootstrap error", err)
	}
	var recovery *synccontract.RebootstrapRequiredError
	if !errors.As(err, &recovery) || recovery.Outcome != synccontract.RecoveryOutcomeSourceIdentityIncompatible {
		t.Fatalf("validateCDCResume(different source) recovery = %#v, want source identity incompatible", recovery)
	}
}

func TestCDCStartLSNRejectsUnretainedCheckpoint(t *testing.T) {
	checkpoint := &synccontract.CheckpointEnvelope{
		Position: synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken([]byte("0/10"))},
	}
	_, err := cdcStartLSN(checkpoint, pglogrepl.LSN(0x20))
	if !errors.Is(err, synccontract.ErrRebootstrapRequired) {
		t.Fatalf("cdcStartLSN(unretained) = %v, want rebootstrap error", err)
	}
}

func TestClassifyCDCStartErrorRequiresRebootstrapForLostWAL(t *testing.T) {
	err := classifyCDCStartError(&synccontract.CheckpointEnvelope{}, &pgconn.PgError{Code: "58P01"})
	if !errors.Is(err, synccontract.ErrRebootstrapRequired) {
		t.Fatalf("classifyCDCStartError = %v, want rebootstrap error", err)
	}
	var recovery *synccontract.RebootstrapRequiredError
	if !errors.As(err, &recovery) || recovery.Outcome != synccontract.RecoveryOutcomeRetentionGap {
		t.Fatalf("classifyCDCStartError recovery = %#v, want retention gap", recovery)
	}
}
