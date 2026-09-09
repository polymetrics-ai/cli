package nativeset

import (
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestDatabaseAdapterPreservesProtectedRegistrations(t *testing.T) {
	adapter := NewDatabaseAdapter()
	for executor, wantName := range map[string]string{
		"native_database/dynamodb.v1": "dynamodb",
		"native_database/mysql.v1":    "mysql",
		"native_database/postgres.v1": "postgres",
	} {
		bundle, err := engine.Load(defs.FS, wantName)
		if err != nil {
			t.Fatal(err)
		}
		connector, selected, err := adapter.Construct(executor, bundle)
		if err != nil {
			t.Fatalf("Construct(%q): %v", executor, err)
		}
		if !selected {
			t.Fatalf("Construct(%q) = no connector", executor)
		}
		if got := connector.Name(); got != wantName {
			t.Fatalf("Construct(%q).Name() = %q, want %q", executor, got, wantName)
		}
	}
}

func TestCompatibilityAdapterExcludesNativeDatabases(t *testing.T) {
	adapter := NewCompatibilityAdapter()
	for _, executor := range []string{"native_database/dynamodb.v1", "native_database/mysql.v1", "native_database/postgres.v1"} {
		if connector, selected, err := adapter.Construct(executor, engine.Bundle{}); err != nil || selected || connector != nil {
			t.Fatalf("Construct(%q) = %T, %t, %v; native database must not be compatibility", executor, connector, selected, err)
		}
	}
}

func TestCompatibilityAdapterLeavesAPIBundlesDeclarative(t *testing.T) {
	adapter := NewCompatibilityAdapter()
	if connector, selected, err := adapter.Construct("api_engine.v1", engine.Bundle{}); err != nil || selected || connector != nil {
		t.Fatalf("Construct(api_engine.v1) = %T, %t, %v; want nil, false, nil", connector, selected, err)
	}
}
