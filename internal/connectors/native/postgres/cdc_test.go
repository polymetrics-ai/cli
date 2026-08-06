package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgconn"

	"polymetrics.ai/internal/connectors"
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

func TestCDCResumeRejectsPublicationScopeDrift(t *testing.T) {
	source := postgresCDCSource{
		identity: synccontract.SourceIdentity{
			Engine:           "postgres",
			AccountOrCluster: "system-one:database-one",
			ObjectScope:      "public.events",
		},
		generation: synccontract.OpaqueToken([]byte("timeline-one")),
	}
	scope := validCDCPublicationScope()
	source = source.withPublicationScope(scope)
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

	changed := scope
	changed.publicationVersion = "42"
	err = validateCDCResume(&checkpoint, postgresCDCSource{
		identity:   source.identity,
		generation: synccontract.OpaqueToken([]byte("timeline-one")),
	}.withPublicationScope(changed))
	if !errors.Is(err, synccontract.ErrRebootstrapRequired) {
		t.Fatalf("validateCDCResume(publication scope drift) = %v, want rebootstrap error", err)
	}
	var recovery *synccontract.RebootstrapRequiredError
	if !errors.As(err, &recovery) || recovery.Outcome != synccontract.RecoveryOutcomeSourceGenerationChanged {
		t.Fatalf("validateCDCResume(publication scope drift) recovery = %#v, want source generation changed", recovery)
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

func TestCDCSlotReuseRequiresDurableCheckpoint(t *testing.T) {
	if err := validateCDCSlotReuse(nil, postgresCDCSlotSetup{created: true}); err != nil {
		t.Fatalf("validateCDCSlotReuse(new slot) = %v", err)
	}
	err := validateCDCSlotReuse(nil, postgresCDCSlotSetup{barrier: pglogrepl.LSN(16)})
	if !errors.Is(err, synccontract.ErrRebootstrapRequired) {
		t.Fatalf("validateCDCSlotReuse(existing slot) = %v, want rebootstrap error", err)
	}
	var recovery *synccontract.RebootstrapRequiredError
	if !errors.As(err, &recovery) || recovery.Outcome != synccontract.RecoveryOutcomeInvalidCheckpoint {
		t.Fatalf("validateCDCSlotReuse(existing slot) recovery = %#v, want invalid checkpoint", recovery)
	}
}

func TestCDCPublicationPluginArgumentQuotesIdentifier(t *testing.T) {
	if got, want := cdcPublicationPluginArgument("CDCFeed"), `publication_names '"CDCFeed"'`; got != want {
		t.Fatalf("cdcPublicationPluginArgument() = %q, want %q", got, want)
	}
}

func TestCDCLifecycleRejectsFixtureMode(t *testing.T) {
	c := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{
		"mode":            "fixture",
		"cdc_publication": "pm_cdc",
	}}
	if _, err := c.CDCSlotName(context.Background(), cfg, "public.users"); !errors.Is(err, errCDCFixtureMode) {
		t.Fatalf("CDCSlotName(fixture) = %v, want fixture rejection", err)
	}
	if err := c.TeardownCDC(context.Background(), cfg, "public.users"); !errors.Is(err, errCDCFixtureMode) {
		t.Fatalf("TeardownCDC(fixture) = %v, want fixture rejection", err)
	}
}

func TestCDCRelationHierarchyRejectsDescendants(t *testing.T) {
	if err := validateCDCRelationHierarchy(false); err != nil {
		t.Fatalf("validateCDCRelationHierarchy(no descendants) = %v", err)
	}
	if err := validateCDCRelationHierarchy(true); !errors.Is(err, errCDCRelationHasDescendants) {
		t.Fatalf("validateCDCRelationHierarchy(descendants) = %v, want descendant rejection", err)
	}
}

func TestCDCPublicationScopeRequiresSelectedRelation(t *testing.T) {
	if err := validateCDCPublicationScope(true); err != nil {
		t.Fatalf("validateCDCPublicationScope(published) = %v", err)
	}
	if err := validateCDCPublicationScope(false); !errors.Is(err, errCDCRelationNotPublished) {
		t.Fatalf("validateCDCPublicationScope(unpublished) = %v, want publication rejection", err)
	}
}

func TestCDCPublicationScopeRequiresAllDML(t *testing.T) {
	for _, tc := range []struct {
		name                         string
		publishesInsert, update, del bool
		wantErr                      bool
	}{
		{name: "all DML", publishesInsert: true, update: true, del: true},
		{name: "missing insert", update: true, del: true, wantErr: true},
		{name: "missing update", publishesInsert: true, del: true, wantErr: true},
		{name: "missing delete", publishesInsert: true, update: true, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCDCPublicationDML(tc.publishesInsert, tc.update, tc.del)
			if tc.wantErr && !errors.Is(err, errCDCPublicationMissingDML) {
				t.Fatalf("validateCDCPublicationDML() = %v, want missing DML rejection", err)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("validateCDCPublicationDML() = %v, want nil", err)
			}
		})
	}
}

func TestCDCPublicationScopeRejectsTruncate(t *testing.T) {
	if err := validateCDCPublicationTruncate(false); err != nil {
		t.Fatalf("validateCDCPublicationTruncate(false) = %v", err)
	}
	if err := validateCDCPublicationTruncate(true); !errors.Is(err, errCDCPublicationPublishesTruncate) {
		t.Fatalf("validateCDCPublicationTruncate(true) = %v, want truncate rejection", err)
	}
}

func TestCDCPublicationScopeRejectsPartialTablePublications(t *testing.T) {
	for _, tc := range []struct {
		name string
		set  func(*postgresCDCPublicationScope)
		want error
	}{
		{name: "row filter", set: func(scope *postgresCDCPublicationScope) { scope.hasRowFilter = true }, want: errCDCPublicationHasRowFilter},
		{name: "column list", set: func(scope *postgresCDCPublicationScope) { scope.hasColumnList = true }, want: errCDCPublicationHasColumnList},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scope := validCDCPublicationScope()
			tc.set(&scope)
			if err := scope.validate(); !errors.Is(err, tc.want) {
				t.Fatalf("scope.validate() = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestCDCPublicationScopeChangeRequiresRebootstrap(t *testing.T) {
	expected := validCDCPublicationScope()
	if err := validateCDCPublicationScopeChange(expected, expected); err != nil {
		t.Fatalf("validateCDCPublicationScopeChange(same scope) = %v", err)
	}
	for _, tc := range []struct {
		name string
		set  func(*postgresCDCPublicationScope)
	}{
		{name: "direct membership", set: func(scope *postgresCDCPublicationScope) { scope.membershipVersion = "43" }},
		{name: "schema membership", set: func(scope *postgresCDCPublicationScope) { scope.namespaceMembershipVersion = "44" }},
		{name: "all tables", set: func(scope *postgresCDCPublicationScope) { scope.publicationAllTables = true }},
		{name: "row filter", set: func(scope *postgresCDCPublicationScope) { scope.hasRowFilter = true }},
		{name: "column list", set: func(scope *postgresCDCPublicationScope) { scope.hasColumnList = true }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := expected
			tc.set(&changed)
			err := validateCDCPublicationScopeChange(expected, changed)
			if !errors.Is(err, synccontract.ErrRebootstrapRequired) {
				t.Fatalf("validateCDCPublicationScopeChange(changed scope) = %v, want rebootstrap error", err)
			}
			var recovery *synccontract.RebootstrapRequiredError
			if !errors.As(err, &recovery) || recovery.Outcome != synccontract.RecoveryOutcomeSourceGenerationChanged {
				t.Fatalf("validateCDCPublicationScopeChange(changed scope) recovery = %#v, want source generation changed", recovery)
			}
		})
	}
}

func TestCDCRelationScopeQueriesAvoidExpandedPublicationTables(t *testing.T) {
	for _, tc := range []struct {
		name    string
		version int
		modern  bool
	}{
		{name: "PostgreSQL 12", version: 120022},
		{name: "PostgreSQL 15", version: cdcPublicationFeaturesVersion, modern: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			query := cdcRelationScopeQuery(tc.version)
			if strings.Contains(query, "pg_publication_tables") {
				t.Fatal("scope query expands pg_publication_tables")
			}
			if !strings.Contains(query, "pg_publication_rel") || !strings.Contains(query, "puballtables") {
				t.Fatal("scope query does not use direct publication membership")
			}
			if tc.modern {
				if !strings.Contains(query, "pg_publication_namespace") || !strings.Contains(query, "prqual") || !strings.Contains(query, "prattrs") {
					t.Fatal("modern scope query does not inspect schema membership and partial-table settings")
				}
				return
			}
			if strings.Contains(query, "pg_publication_namespace") || strings.Contains(query, "prqual") || strings.Contains(query, "prattrs") {
				t.Fatal("legacy scope query requires a modern publication catalog")
			}
		})
	}
}

func validCDCPublicationScope() postgresCDCPublicationScope {
	return postgresCDCPublicationScope{
		publicationOID:     "101",
		publicationVersion: "102",
		relationOID:        "103",
		membershipVersion:  "104",
		published:          true,
		publishesInsert:    true,
		publishesUpdate:    true,
		publishesDelete:    true,
	}
}

func TestClassifyCDCSlotDropErrorRefusesActiveSlot(t *testing.T) {
	err := classifyCDCSlotDropError(&pgconn.PgError{Code: "55006"})
	if !errors.Is(err, ErrCDCSlotActive) {
		t.Fatalf("classifyCDCSlotDropError = %v, want active slot refusal", err)
	}
}
