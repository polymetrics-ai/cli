package nativeset

import (
	"fmt"
	"sort"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	bingads "polymetrics.ai/internal/connectors/native/bing-ads"
	nativedynamodb "polymetrics.ai/internal/connectors/native/dynamodb"
	nativefaker "polymetrics.ai/internal/connectors/native/faker"
	nativehubspot "polymetrics.ai/internal/connectors/native/hubspot"
	nativemysql "polymetrics.ai/internal/connectors/native/mysql"
	nativepostgres "polymetrics.ai/internal/connectors/native/postgres"
	tallyprime "polymetrics.ai/internal/connectors/native/tally-prime"
)

type nativeFactory struct {
	connector string
	executor  string
	new       func(engine.Bundle) (connectors.Connector, error)
}

var protectedDatabaseFactories = []nativeFactory{
	{connector: "dynamodb", executor: "native_database/dynamodb.v1", new: func(bundle engine.Bundle) (connectors.Connector, error) {
		return nativedynamodb.NewFromBundle(bundle), nil
	}},
	{connector: "mysql", executor: "native_database/mysql.v1", new: func(bundle engine.Bundle) (connectors.Connector, error) {
		return nativemysql.NewFromBundle(bundle), nil
	}},
	{connector: "postgres", executor: "native_database/postgres.v1", new: func(bundle engine.Bundle) (connectors.Connector, error) { return nativepostgres.NewFromBundle(bundle) }},
}

var protectedCompatibilityFactories = []nativeFactory{
	{connector: "bing-ads", executor: "closed_typed/bing-ads.v1", new: func(bundle engine.Bundle) (connectors.Connector, error) { return bingads.NewFromBundle(bundle), nil }},
	{connector: "faker", executor: "closed_typed/faker.v1", new: func(bundle engine.Bundle) (connectors.Connector, error) {
		return nativefaker.NewFromBundle(bundle), nil
	}},
	{connector: "hubspot", executor: "closed_typed/hubspot.v1", new: func(bundle engine.Bundle) (connectors.Connector, error) { return nativehubspot.NewFromBundle(bundle) }},
	{connector: "tally-prime", executor: "closed_typed/tally-prime.v1", new: func(bundle engine.Bundle) (connectors.Connector, error) { return tallyprime.NewFromBundle(bundle), nil }},
}

// ManifestSelection is one closed generator-time executor selection. Runtime
// construction validates this ID against the adapter inventories again.
type ManifestSelection struct {
	Connector string
	Executor  string
}

// ManifestSelections returns the complete, sorted native/compatibility
// selection inventory used by connectorgen without a connector-name branch in
// shared generator or runtime code.
func ManifestSelections() ([]ManifestSelection, error) {
	factories := append(append([]nativeFactory(nil), protectedDatabaseFactories...), protectedCompatibilityFactories...)
	selections := make([]ManifestSelection, 0, len(factories))
	seenConnectors := make(map[string]struct{}, len(factories))
	seenExecutors := make(map[string]struct{}, len(factories))
	for _, factory := range factories {
		if factory.connector == "" || factory.executor == "" || factory.new == nil {
			return nil, fmt.Errorf("native manifest selection is incomplete")
		}
		if _, exists := seenConnectors[factory.connector]; exists {
			return nil, fmt.Errorf("duplicate native manifest connector %q", factory.connector)
		}
		if _, exists := seenExecutors[factory.executor]; exists {
			return nil, fmt.Errorf("duplicate native manifest executor %q", factory.executor)
		}
		seenConnectors[factory.connector] = struct{}{}
		seenExecutors[factory.executor] = struct{}{}
		selections = append(selections, ManifestSelection{Connector: factory.connector, Executor: factory.executor})
	}
	sort.Slice(selections, func(i, j int) bool { return selections[i].Connector < selections[j].Connector })
	return selections, nil
}

type factoryAdapter struct {
	factories map[string]func(engine.Bundle) (connectors.Connector, error)
}

// DatabaseAdapter is the closed executor-keyed native database inventory.
type DatabaseAdapter struct {
	factoryAdapter
}

// CompatibilityAdapter is the closed executor-keyed legacy inventory.
type CompatibilityAdapter struct {
	factoryAdapter
}

// NewDatabaseAdapter constructs the three permitted native database executors.
func NewDatabaseAdapter() DatabaseAdapter {
	adapter, err := newFactoryAdapter(protectedDatabaseFactories)
	if err != nil {
		panic(err)
	}
	return DatabaseAdapter{factoryAdapter: adapter}
}

// NewCompatibilityAdapter constructs the only permitted legacy compatibility inventory.
func NewCompatibilityAdapter() CompatibilityAdapter {
	adapter, err := newFactoryAdapter(protectedCompatibilityFactories)
	if err != nil {
		panic(err)
	}
	return CompatibilityAdapter{factoryAdapter: adapter}
}

func newFactoryAdapter(factories []nativeFactory) (factoryAdapter, error) {
	out := factoryAdapter{factories: make(map[string]func(engine.Bundle) (connectors.Connector, error), len(factories))}
	for _, factory := range factories {
		if factory.connector == "" || factory.executor == "" || factory.new == nil {
			return factoryAdapter{}, fmt.Errorf("native executor factory is incomplete")
		}
		if _, exists := out.factories[factory.executor]; exists {
			return factoryAdapter{}, fmt.Errorf("duplicate native executor factory %q", factory.executor)
		}
		out.factories[factory.executor] = factory.new
	}
	return out, nil
}

// Has reports whether the sealed adapter owns executor without constructing it.
func (a factoryAdapter) Has(executor string) bool {
	_, ok := a.factories[executor]
	return ok
}

// Construct returns the connector for one exact sealed executor identity using
// the caller's already-selected execution bundle.
func (a factoryAdapter) Construct(executor string, bundle engine.Bundle) (connectors.Connector, bool, error) {
	factory, ok := a.factories[executor]
	if !ok {
		return nil, false, nil
	}
	connector, err := factory(bundle)
	if err != nil {
		return nil, true, err
	}
	if connector == nil || connector.Name() != bundle.Name {
		return nil, true, fmt.Errorf("native executor %q returned the wrong connector", executor)
	}
	return connector, true, nil
}
