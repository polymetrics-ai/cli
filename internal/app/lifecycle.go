package app

import (
	"context"
	"os"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/coordination"
)

// RuntimeOptions are borrowed execution capabilities installed before durable
// parking starts. The caller owns SharedRateLimits and closes it after all Apps.
type RuntimeOptions struct {
	// Instance construction seam for package tests; never an ambient callback.
	newParking       func(*App, coordination.RateParkingStore) *coordination.RateParkingCoordinator
	Context          context.Context
	SharedRateLimits connectors.SharedRateLimitCoordinator
}

func OpenWithRuntime(root string, options RuntimeOptions) (*App, error) {
	return openWithRegistryWithStat(root, false, bundleregistry.NewRegistry, os.Stat, options)
}
func OpenForReverseExecutionWithRuntime(root string, options RuntimeOptions) (*App, error) {
	return openWithRegistryWithStat(root, true, bundleregistry.NewRegistry, os.Stat, options)
}
func OpenWithRegistryAndRuntime(root string, registry *connectors.Registry, options RuntimeOptions) (*App, error) {
	return openWithRegistryWithStat(root, false, func() (*connectors.Registry, error) { return registry, nil }, os.Stat, options)
}
func OpenForReverseExecutionWithRegistryAndRuntime(root string, registry *connectors.Registry, options RuntimeOptions) (*App, error) {
	return openWithRegistryWithStat(root, true, func() (*connectors.Registry, error) { return registry, nil }, os.Stat, options)
}

// ExecutableRuntime copies this App's borrowed capability without opening a
// client or contacting a service. Anonymous commands use the same injection as
// credential-backed execution; metadata/preflight configs remain inert.
func (a *App) ExecutableRuntime(runtime connectors.RuntimeConfig) connectors.RuntimeConfig {
	runtime.SharedRateLimits = a.sharedRateLimits
	return runtime
}

// Close stops new parked resumes, cancels and joins existing callbacks, retaining
// durable work. Call after foreground operations finish. No borrowed capability
// is closed, and this does not make arbitrary concurrent App mutation safe.
func (a *App) Close() error {
	if a == nil {
		return nil
	}
	a.closeOnce.Do(func() {
		if a.lifecycleCancel != nil {
			a.lifecycleCancel()
		}
		if a.rateParking != nil {
			a.rateParking.Close()
		}
	})
	return nil
}
