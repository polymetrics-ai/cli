package app

import (
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestPostgresManagedTargetContractModeRequiresExactDeclaredPair(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "postgres")
	if err != nil {
		t.Fatalf("load PostgreSQL bundle: %v", err)
	}
	postgres := engine.New(bundle, nil)
	if !postgresManagedTargetContractMode(postgres, postgres, "incremental_append") {
		t.Fatal("PostgreSQL declared pair did not select its incremental contract mode")
	}
	if postgresManagedTargetContractMode(postgres, postgres, "full_refresh_append") {
		t.Fatal("legacy full-refresh spelling selected PostgreSQL incremental contract mode")
	}
	githubBundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load GitHub bundle: %v", err)
	}
	if postgresManagedTargetContractMode(engine.New(githubBundle, nil), engine.New(githubBundle, nil), "incremental_append") {
		t.Fatal("non-PostgreSQL pair selected PostgreSQL contract mode")
	}
}
