package engine

import (
	"reflect"
	"testing"

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
	}
	if modes := bundle.Database.AdmittedModes(); !reflect.DeepEqual(modes, wantModes) {
		t.Fatalf("database definition admitted modes = %v, want the five private managed-target modes %v", modes, wantModes)
	}
	if bundle.Metadata.Capabilities.Write || !bundle.Metadata.Capabilities.CDC {
		t.Fatalf("PostgreSQL metadata capabilities = %+v, want write=false and cdc=true", bundle.Metadata.Capabilities)
	}
}
