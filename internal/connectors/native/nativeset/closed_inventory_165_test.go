package nativeset

import (
	"slices"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestProtectedFactoryInventory165(t *testing.T) {
	expected := []ManifestSelection{
		{Connector: "bing-ads", Executor: "closed_typed/bing-ads.v1"},
		{Connector: "dynamodb", Executor: "native_database/dynamodb.v1"},
		{Connector: "faker", Executor: "closed_typed/faker.v1"},
		{Connector: "hubspot", Executor: "closed_typed/hubspot.v1"},
		{Connector: "mysql", Executor: "native_database/mysql.v1"},
		{Connector: "postgres", Executor: "native_database/postgres.v1"},
		{Connector: "tally-prime", Executor: "closed_typed/tally-prime.v1"},
	}
	actual, err := ManifestSelections()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(actual, expected) {
		t.Fatalf("protected membership changed: got=%v want=%v", actual, expected)
	}
	database := NewDatabaseAdapter()
	compatibility := NewCompatibilityAdapter()
	for _, selection := range expected {
		t.Run(selection.Connector, func(t *testing.T) {
			bundle, err := engine.Load(defs.FS, selection.Connector)
			if err != nil {
				t.Fatal(err)
			}
			selectedAdapter := compatibility.factoryAdapter
			other := database.factoryAdapter
			if strings.HasPrefix(selection.Executor, "native_database/") {
				selectedAdapter, other = other, selectedAdapter
			}
			connector, selected, err := selectedAdapter.Construct(selection.Executor, bundle)
			if err != nil || !selected || connector == nil || connector.Name() != selection.Connector {
				t.Fatalf("protected constructor failed: selected=%t error=%v connector=%T", selected, err, connector)
			}
			if other.Has(selection.Executor) {
				t.Fatal("protected executor crossed adapter ownership")
			}
			for _, unknown := range []string{"", selection.Connector, selection.Executor + ".unknown", "api_engine.v1"} {
				connector, selected, err := selectedAdapter.Construct(unknown, bundle)
				if err != nil || selected || connector != nil {
					t.Fatalf("unknown selection %q fell back: %T %t %v", unknown, connector, selected, err)
				}
			}
		})
	}
	wrong := slices.Clone(actual)
	wrong[0].Executor = wrong[1].Executor
	if slices.Equal(wrong, expected) {
		t.Fatal("same-count executor substitution escaped inventory oracle")
	}
}
