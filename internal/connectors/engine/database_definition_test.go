package engine

import (
	"context"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/synccontract"
)

func TestBundleLoadPostgresDatabaseDefinitionWithProvenCDCCapability(t *testing.T) {
	bundle, err := Load(defs.FS, "postgres")
	if err != nil {
		t.Fatalf("Load(defs.FS, postgres) error = %v", err)
	}
	if bundle.Database == nil {
		t.Fatal("postgres bundle has no loaded database.json definition")
	}
	if got := bundle.Database.Driver(); got.ID != "postgres" || got.Protocol != "postgres-wire" || got.APIVersion != 1 {
		t.Fatalf("database driver declaration = %#v, want postgres wire API v1", got)
	}
	wantModes := []synccontract.Mode{
		synccontract.ModeFullOverwrite,
		synccontract.ModeFullAppend,
		synccontract.ModeIncrementalAppend,
		synccontract.ModeIncrementalUpsert,
		synccontract.ModeIncrementalDedupe,
		synccontract.ModeIncrementalDedupeHistory,
	}
	if modes := bundle.Database.AdmittedModes(); !reflect.DeepEqual(modes, wantModes) {
		t.Fatalf("PostgreSQL database definition admitted modes = %v, want the six private managed-target modes including history %v", modes, wantModes)
	}
	if bundle.Metadata.Capabilities.Write || !bundle.Metadata.Capabilities.CDC {
		t.Fatalf("PostgreSQL metadata capabilities = %+v, want write=false and cdc=true", bundle.Metadata.Capabilities)
	}
}

func TestNonPostgresDatabaseDefinitionRetainsFivePrivateManagedTargetModes(t *testing.T) {
	postgresDefinition, err := fs.ReadFile(defs.FS, "postgres/database.json")
	if err != nil {
		t.Fatalf("ReadFile(postgres/database.json) error = %v", err)
	}
	// PostgreSQL is the only shipped database definition. This independent
	// non-PostgreSQL declaration keeps the pre-history five-mode contract;
	// database/history_route_test.go separately proves every non-PostgreSQL
	// route rejects history before a fake driver, session, ledger, or target can
	// perform an operation.
	nonPostgresDefinition := strings.NewReplacer(
		`"id": "postgres"`, `"id": "mysql"`,
		`"protocol": "postgres-wire"`, `"protocol": "mysql-wire"`,
		",\n    \"incremental_dedupe_history\"", "",
	).Replace(string(postgresDefinition))
	definition, err := database.Load(context.Background(), fstest.MapFS{
		"database.json": &fstest.MapFile{Data: []byte(nonPostgresDefinition)},
	})
	if err != nil {
		t.Fatalf("Load(non-PostgreSQL database definition) error = %v", err)
	}
	if got := definition.Driver(); got.ID != "mysql" || got.Protocol != "mysql-wire" || got.APIVersion != 1 {
		t.Fatalf("non-PostgreSQL database driver declaration = %#v, want mysql wire API v1", got)
	}
	wantModes := []synccontract.Mode{
		synccontract.ModeFullOverwrite,
		synccontract.ModeFullAppend,
		synccontract.ModeIncrementalAppend,
		synccontract.ModeIncrementalUpsert,
		synccontract.ModeIncrementalDedupe,
	}
	if modes := definition.AdmittedModes(); !reflect.DeepEqual(modes, wantModes) {
		t.Fatalf("non-PostgreSQL database definition admitted modes = %v, want only the five pre-history private modes %v", modes, wantModes)
	}
}
