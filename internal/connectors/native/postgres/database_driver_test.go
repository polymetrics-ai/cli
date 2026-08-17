package postgres_test

import (
	"testing"

	"polymetrics.ai/internal/connectors/database"
	native "polymetrics.ai/internal/connectors/native/postgres"
)

func TestPostgresDatabaseDriverReferenceSeam(t *testing.T) {
	var _ database.Driver = native.DatabaseDriver{}
	var _ database.ManagedTargetProvisioningDriver = (*native.DatabaseDriver)(nil)
	var _ database.ManagedTargetDeliveryLedgerStore = (*native.DatabaseDriver)(nil)
	var _ database.DatabaseWriteDriver = (*native.DatabaseDriver)(nil)

	descriptor := (native.DatabaseDriver{}).DatabaseDriverDescriptor()
	if descriptor.ID != "postgres" || descriptor.Protocol != "postgres-wire" || descriptor.APIVersion != 1 {
		t.Fatalf("PostgreSQL database driver descriptor = %#v, want stable postgres wire v1 reference", descriptor)
	}

	caps := native.New().Metadata().Capabilities
	if caps.Write || !caps.CDC || caps.Query {
		t.Fatalf("PostgreSQL capabilities = %+v, want write=false cdc=true query=false; the certified managed target is a closed transport", caps)
	}
}

func TestPostgresDatabaseDriverRequiresPinnedConnection(t *testing.T) {
	if _, err := native.NewDatabaseDriver(nil); err == nil {
		t.Fatal("NewDatabaseDriver(nil) succeeded, want a refused unpinned connection")
	}
}
