package bundleregistry

import (
	"context"
	"fmt"
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/hooks/hookset"
	"polymetrics.ai/internal/connectors/manifestindex"
	"polymetrics.ai/internal/connectors/manifeststore"
	"polymetrics.ai/internal/connectors/native/nativeset"
)

// Construction owns one immutable, validated executor catalog for a caller.
// It is created before App project/vault work so invalid generated selections
// cannot cross a local or provider-I/O boundary.
type Construction struct {
	index         manifestindex.Index
	hooks         hookFactories
	extensions    map[string]string
	database      nativeset.DatabaseAdapter
	compatibility nativeset.CompatibilityAdapter
	factories     ExecutorFactories
	store         *manifeststore.BundleStore
	leaseMu       sync.Mutex
	leases        map[string]*manifeststore.GenerationLease
}

// NewConstruction validates generated execution selections and the complete
// explicit factory inventory without loading a connector bundle.
func NewConstruction() (*Construction, error) {
	entries := manifestindex.GeneratedEntries()
	index, err := manifestindex.New(entries, len(entries))
	if err != nil {
		return nil, fmt.Errorf("load generated manifest index: %w", err)
	}
	hooks, err := newHookFactories(hookset.Factories())
	if err != nil {
		return nil, err
	}
	construction := &Construction{
		index:         index,
		hooks:         hooks,
		database:      nativeset.NewDatabaseAdapter(),
		compatibility: nativeset.NewCompatibilityAdapter(),
		leases:        make(map[string]*manifeststore.GenerationLease),
	}
	factories, err := NewExecutorFactories(
		ExecutorFactory{
			ID: "api_engine.v1",
			Construct: func(bundle engine.Bundle) (connectors.Connector, error) {
				hooks, err := construction.hooks.construct(construction.extensions[bundle.Name], bundle.Name)
				if err != nil {
					return nil, err
				}
				return engine.New(bundle, hooks), nil
			},
		},
		nativeDatabaseExecutorFactory(construction.database, "native_database/dynamodb.v1"),
		nativeDatabaseExecutorFactory(construction.database, "native_database/mysql.v1"),
		nativeDatabaseExecutorFactory(construction.database, "native_database/postgres.v1"),
	)
	if err != nil {
		return nil, err
	}
	construction.factories = factories
	if err := construction.validateEntries(index.List()); err != nil {
		return nil, err
	}
	store, err := manifeststore.NewBundleStore(index, manifeststore.Limits{Entries: 16, Bytes: 64 << 20}, func(ctx context.Context, entry manifestindex.Entry) (manifeststore.LoadedBundle, error) {
		if err := ctx.Err(); err != nil {
			return manifeststore.LoadedBundle{}, err
		}
		bundle, err := engine.Load(defs.FS, entry.Connector)
		if err != nil {
			return manifeststore.LoadedBundle{}, fmt.Errorf("load execution bundle %q: %w", entry.Connector, err)
		}
		return manifeststore.LoadedBundle{Bundle: &bundle, Identity: bundle.Identity}, nil
	})
	if err != nil {
		return nil, err
	}
	construction.store = store
	return construction, nil
}

func nativeDatabaseExecutorFactory(adapter nativeset.DatabaseAdapter, id string) ExecutorFactory {
	return ExecutorFactory{
		ID: id,
		Construct: func(bundle engine.Bundle) (connectors.Connector, error) {
			connector, selected, err := adapter.Construct(id, bundle)
			if err != nil {
				return nil, err
			}
			if !selected {
				return nil, fmt.Errorf("native database executor %q is unavailable", id)
			}
			return connector, nil
		},
	}
}

// NewRegistry constructs the selected production registry through one explicit
// factory catalog. It returns typed errors for invalid generated selections.
func NewRegistry() (*connectors.Registry, error) {
	construction, err := NewConstruction()
	if err != nil {
		return nil, err
	}
	return construction.BuildRegistry()
}

// BuildRegistry returns a metadata-complete registry without decoding
// definition bundles. Its resolver acquires exactly the selected indexed bundle
// when a caller needs a connector implementation.
func (c *Construction) BuildRegistry() (*connectors.Registry, error) {
	entries := c.index.List()
	if err := c.validateEntries(entries); err != nil {
		return nil, err
	}
	metadata := make([]connectors.Metadata, 0, len(entries))
	summaries := make([]connectors.CommandSummary, 0, len(entries))
	for _, entry := range entries {
		if entry.Metadata.Name != entry.Connector {
			return nil, fmt.Errorf("manifest entry %q has incomplete metadata", entry.Connector)
		}
		metadata = append(metadata, entry.Metadata)
		if entry.CommandUsage != "" {
			summaries = append(summaries, connectors.CommandSummary{
				Connector: entry.Connector,
				Usage:     entry.CommandUsage,
				Tagline:   entry.CommandTagline,
			})
		}
	}
	registry, err := connectors.NewLazyRegistry(metadata, c.construct, summaries...)
	if err != nil {
		return nil, fmt.Errorf("create lazy connector registry: %w", err)
	}
	registry.RegisterBuiltins()
	if err := registry.ValidateIconCoverage(); err != nil {
		return nil, err
	}
	return registry, nil
}
func (c *Construction) construct(ctx context.Context, name string) (connectors.Connector, error) {
	entry, ok := c.index.Lookup(name)
	if !ok {
		return nil, fmt.Errorf("manifest entry %q not found", name)
	}
	handle, err := c.store.AcquireEntry(ctx, entry)
	if err != nil {
		return nil, fmt.Errorf("acquire connector %q: %w", entry.Connector, err)
	}
	lease := handle.HoldGeneration()
	if lease == nil {
		handle.Release()
		return nil, fmt.Errorf("hold selected generation for %q", entry.Connector)
	}
	bundle := *handle.Bundle()
	connector, selected, err := c.compatibility.Construct(entry.Executor, bundle)
	if err != nil {
		lease.Release()
		handle.Release()
		return nil, fmt.Errorf("construct compatibility connector %q: %w", entry.Connector, err)
	}
	if selected {
		if connector.Name() != entry.Connector {
			lease.Release()
			handle.Release()
			return nil, fmt.Errorf("compatibility executor %q returned %q for %q", entry.Executor, connector.Name(), entry.Connector)
		}
		c.retainLease(entry.Connector, lease)
		handle.Release()
		return connector, nil
	}
	connector, err = c.factories.Construct(entry, bundle)
	if err != nil {
		lease.Release()
		handle.Release()
		return nil, fmt.Errorf("construct connector %q: %w", entry.Connector, err)
	}
	c.retainLease(entry.Connector, lease)
	handle.Release()
	return connector, nil
}

// Close releases every selected generation hold owned by this construction.
func (c *Construction) Close() {
	if c == nil {
		return
	}
	c.leaseMu.Lock()
	leases := c.leases
	c.leases = nil
	c.leaseMu.Unlock()
	for _, lease := range leases {
		lease.Release()
	}
}

func (c *Construction) retainLease(connector string, lease *manifeststore.GenerationLease) {
	c.leaseMu.Lock()
	if c.leases == nil {
		c.leases = make(map[string]*manifeststore.GenerationLease)
	}
	previous := c.leases[connector]
	c.leases[connector] = lease
	c.leaseMu.Unlock()
	if previous != nil {
		previous.Release()
	}
}

func (c *Construction) validateEntries(entries []manifestindex.Entry) error {
	extensions := make(map[string]string, len(entries))
	for _, entry := range entries {
		if isBuiltinConnector(entry.Connector) {
			return fmt.Errorf("manifest entry %q collides with a reserved builtin", entry.Connector)
		}
		if entry.Extension != "" {
			if entry.Executor != "api_engine.v1" {
				return fmt.Errorf("manifest entry %q selects extension %q with non-engine executor %q", entry.Connector, entry.Extension, entry.Executor)
			}
			if err := c.hooks.validate(entry.Extension, entry.Connector); err != nil {
				return err
			}
			extensions[entry.Connector] = entry.Extension
		}
		if c.compatibility.Has(entry.Executor) {
			continue
		}
		if !c.factories.Has(entry.Executor) {
			return fmt.Errorf("%w: %q", ErrUnknownExecutor, entry.Executor)
		}
	}
	c.extensions = extensions
	return nil
}

func isBuiltinConnector(name string) bool {
	switch name {
	case "sample", "file", "warehouse", "outbox":
		return true
	default:
		return false
	}
}

// New preserves the test and presentation convenience API. Executable App
// construction must use NewRegistry so selection failures stay returned.
func New() *connectors.Registry {
	registry, err := NewRegistry()
	if err != nil {
		panic(err)
	}
	return registry
}
