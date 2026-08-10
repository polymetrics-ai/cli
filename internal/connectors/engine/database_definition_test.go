package engine

import (
	"testing"

	"polymetrics.ai/internal/connectors/defs"
)

func TestBundleLoadPostgresDatabaseDefinitionWithoutCapabilityPromotion(t *testing.T) {
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
	if modes := bundle.Database.AdmittedModes(); len(modes) != 0 {
		t.Fatalf("database definition admitted modes = %v, want no unimplemented operation claim", modes)
	}
	if bundle.Metadata.Capabilities.Write || bundle.Metadata.Capabilities.CDC {
		t.Fatalf("PostgreSQL metadata promoted capabilities: %+v", bundle.Metadata.Capabilities)
	}
}
