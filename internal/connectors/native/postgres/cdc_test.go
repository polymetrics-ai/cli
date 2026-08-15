package postgres

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgconn"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

func TestCDCRequiresCrashRecoverableStageBeforeValidatingSource(t *testing.T) {
	err := New().ReadCDC(context.Background(), connectors.CDCReadRequest{}, nil)
	if !errors.Is(err, errCDCProjectDirectory) {
		t.Fatalf("ReadCDC = %v, want missing stage project directory", err)
	}
}

func TestCDCFailClosedBeforeOpeningSourceConnection(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	accepted := make(chan struct{}, 1)
	acceptDone := make(chan struct{})
	go func() {
		defer close(acceptDone)
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		_ = conn.Close()
		accepted <- struct{}{}
	}()
	t.Cleanup(func() {
		_ = listener.Close()
		select {
		case <-acceptDone:
		case <-time.After(time.Second):
			t.Error("listener did not stop")
		}
	})

	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = New().ReadCDC(ctx, connectors.CDCReadRequest{
		Stream: "public.users",
		Config: connectors.RuntimeConfig{
			Config: map[string]string{
				"host":            host,
				"port":            port,
				"database":        "analytics",
				"username":        "reader",
				"sslmode":         "disabled",
				"cdc_publication": "pm_publication",
			},
			Secrets: map[string]string{"password": t.Name()},
		},
		DurableCheckpointCommitter: cdcDiscardingCommitter{},
	}, func(connectors.CDCEvent) error {
		return nil
	})
	if !errors.Is(err, errCDCProjectDirectory) {
		t.Fatalf("ReadCDC = %v, want a local stage-directory error before source access", err)
	}

	select {
	case <-accepted:
		t.Fatal("ReadCDC opened a source connection before it could create its durable stage")
	case <-time.After(50 * time.Millisecond):
		t.Log("ReadCDC rejected the missing durable stage directory without opening the local source listener")
	}
}

type cdcDiscardingCommitter struct{}

func (cdcDiscardingCommitter) CommitDurableChangefeedCheckpoint(context.Context, synccontract.CheckpointEnvelope) error {
	return nil
}

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
	var recovery *synccontract.RebootstrapRequiredError
	if !errors.As(err, &recovery) || recovery.Outcome != synccontract.RecoveryOutcomeRetentionGap {
		t.Fatalf("cdcStartLSN(unretained) recovery = %#v, want retention gap", recovery)
	}
}

func TestCDCServerVersionRequiresPostgreSQL14OrNewer(t *testing.T) {
	if err := validateCDCServerVersion(139999); !errors.Is(err, errCDCServerVersion) {
		t.Fatalf("validateCDCServerVersion(139999) = %v, want PostgreSQL 14 requirement", err)
	}
	if err := validateCDCServerVersion(140000); err != nil {
		t.Fatalf("validateCDCServerVersion(140000) = %v, want nil", err)
	}
}

func TestCDCResumeRejectsIncompatibleNativeCheckpointShape(t *testing.T) {
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

	for _, mutate := range []func(*synccontract.CheckpointEnvelope){
		func(checkpoint *synccontract.CheckpointEnvelope) { checkpoint.SchemaVersion = "other-schema" },
		func(checkpoint *synccontract.CheckpointEnvelope) { checkpoint.ProtocolVersion = "other-protocol" },
		func(checkpoint *synccontract.CheckpointEnvelope) { checkpoint.SnapshotBarrier.Kind = "other-barrier" },
	} {
		mutated := checkpoint.Clone()
		mutate(&mutated)
		err := validateCDCResume(&mutated, source)
		if !errors.Is(err, synccontract.ErrRebootstrapRequired) {
			t.Fatalf("validateCDCResume(incompatible native shape) = %v, want rebootstrap error", err)
		}
		var recovery *synccontract.RebootstrapRequiredError
		if !errors.As(err, &recovery) || recovery.Outcome != synccontract.RecoveryOutcomeInvalidCheckpoint {
			t.Fatalf("validateCDCResume(incompatible native shape) recovery = %#v, want invalid checkpoint", recovery)
		}
	}
}

func TestSlotRetentionLSNUsesRestartPosition(t *testing.T) {
	got, err := slotRetentionLSN(postgresReplicationSlot{
		confirmedLSN: "0/40",
		restartLSN:   "0/20",
	})
	if err != nil {
		t.Fatalf("slotRetentionLSN: %v", err)
	}
	if want := pglogrepl.LSN(0x20); got != want {
		t.Fatalf("slotRetentionLSN = %s, want %s", got, want)
	}
}

func TestStandbyStatusAcknowledgesOnlyTheDurablePosition(t *testing.T) {
	position := pglogrepl.LSN(0x1234)
	status := standbyStatusUpdate(position)
	if status.WALWritePosition != position || status.WALFlushPosition != position || status.WALApplyPosition != position {
		t.Fatalf("standby status = %+v, want all positions at %s", status, position)
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

func TestCDCReplicationConfigForcesAndVerifiesUTF8(t *testing.T) {
	config := &pgconn.Config{RuntimeParams: map[string]string{"client_encoding": "LATIN1"}}
	configureCDCReplicationConfig(config)
	if got := config.RuntimeParams["replication"]; got != "database" {
		t.Fatalf("replication runtime parameter = %q, want database", got)
	}
	if got := config.RuntimeParams["client_encoding"]; got != cdcClientEncoding {
		t.Fatalf("client_encoding runtime parameter = %q, want %q", got, cdcClientEncoding)
	}
	for _, tc := range []struct {
		encoding string
		wantErr  bool
	}{
		{encoding: "UTF8"},
		{encoding: "utf8"},
		{encoding: "LATIN1", wantErr: true},
		{wantErr: true},
	} {
		err := validateCDCClientEncoding(tc.encoding)
		if tc.wantErr && !errors.Is(err, errCDCClientEncoding) {
			t.Fatalf("validateCDCClientEncoding(%q) = %v, want UTF-8 rejection", tc.encoding, err)
		}
		if !tc.wantErr && err != nil {
			t.Fatalf("validateCDCClientEncoding(%q) = %v, want nil", tc.encoding, err)
		}
	}
}

func TestCDCReplicaIdentityModesFailClosed(t *testing.T) {
	for _, tc := range []struct {
		name          string
		identity      string
		hasPrimaryKey bool
		want          error
	}{
		{name: "default primary key", identity: "d", hasPrimaryKey: true},
		{name: "default no primary key", identity: "d", want: errCDCReplicaIdentityDefault},
		{name: "full", identity: "f", hasPrimaryKey: true, want: errCDCReplicaIdentityFull},
		{name: "using index", identity: "i", hasPrimaryKey: true, want: errCDCReplicaIdentityIndex},
		{name: "nothing", identity: "n", want: errCDCReplicaIdentityNothing},
		{name: "unknown", identity: "x", want: errCDCReplicaIdentityUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCDCReplicaIdentity(tc.identity, tc.hasPrimaryKey)
			if tc.want == nil && err != nil {
				t.Fatalf("validateCDCReplicaIdentity(%q) = %v, want nil", tc.identity, err)
			}
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("validateCDCReplicaIdentity(%q) = %v, want %v", tc.identity, err, tc.want)
			}
		})
	}
}
