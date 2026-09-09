package bundleregistry

import (
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestindex"
)

func TestNewExecutorFactoriesRejectsDuplicateIDsBeforeConstruction(t *testing.T) {
	calls := 0
	factory := func(engine.Bundle) (connectors.Connector, error) {
		calls++
		return nil, nil
	}

	_, err := NewExecutorFactories(
		ExecutorFactory{ID: "api_engine.v1", Construct: factory},
		ExecutorFactory{ID: "api_engine.v1", Construct: factory},
	)
	if err == nil {
		t.Fatal("NewExecutorFactories accepted duplicate executor IDs")
	}
	if calls != 0 {
		t.Fatalf("constructors called = %d, want 0 before duplicate rejection", calls)
	}
}

func TestExecutorFactoriesRejectUnknownOrEmptySelectionBeforeConstruction(t *testing.T) {
	calls := 0
	factories, err := NewExecutorFactories(ExecutorFactory{
		ID: "api_engine.v1",
		Construct: func(engine.Bundle) (connectors.Connector, error) {
			calls++
			return nil, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	bundle := engine.Bundle{Name: "alpha"}
	for _, executor := range []string{"", "unknown.v1"} {
		t.Run(executor, func(t *testing.T) {
			_, err := factories.Construct(manifestindex.Entry{Connector: "alpha", Executor: executor}, bundle)
			if err == nil {
				t.Fatalf("Construct accepted executor %q", executor)
			}
		})
	}
	if calls != 0 {
		t.Fatalf("constructors called = %d, want 0 for invalid selections", calls)
	}
}

func TestExecutorFactoriesConstructExactSelectedBundle(t *testing.T) {
	calls := 0
	factories, err := NewExecutorFactories(ExecutorFactory{
		ID: "api_engine.v1",
		Construct: func(bundle engine.Bundle) (connectors.Connector, error) {
			calls++
			return engine.New(bundle, nil), nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	connector, err := factories.Construct(
		manifestindex.Entry{Connector: "alpha", Executor: "api_engine.v1"},
		engine.Bundle{Name: "alpha"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if connector.Name() != "alpha" {
		t.Fatalf("constructed connector = %q, want alpha", connector.Name())
	}
	if calls != 1 {
		t.Fatalf("constructors called = %d, want 1", calls)
	}
}
