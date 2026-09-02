package nativeset

import (
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestDatabaseConnectorForPreservesProtectedRegistrations(t *testing.T) {
	for _, name := range []string{"dynamodb", "mysql", "postgres"} {
		bundle, err := engine.Load(defs.FS, name)
		if err != nil {
			t.Fatalf("Load(%q): %v", name, err)
		}
		connector, ok := DatabaseConnectorFor(bundle.Name)
		if !ok {
			t.Fatalf("DatabaseConnectorFor(%q) = no connector", name)
		}
		if got := connector.Name(); got != name {
			t.Fatalf("DatabaseConnectorFor(%q).Name() = %q", name, got)
		}
	}
}

func TestDatabaseConnectorForLeavesAPIBundlesDeclarative(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("Load(github): %v", err)
	}
	if connector, ok := DatabaseConnectorFor(bundle.Name); ok || connector != nil {
		t.Fatalf("DatabaseConnectorFor(github) = %T, %t; want nil, false", connector, ok)
	}
}
