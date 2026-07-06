package engine

import (
	"context"
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

// Runtime is passed to Stream/Write/Check hooks so they can reuse the
// engine's already-built HTTP plumbing and bundle context instead of
// re-deriving it. It is populated by engine.Connector at dispatch time;
// wave0 hooks are exercised only via in-test fakes (SPEC §1.3), so the
// fields below are the minimal shape those fakes and future wave callers
// need.
type Runtime struct {
	Requester *connsdk.Requester
	Bundle    *Bundle
	Config    connectors.RuntimeConfig
}

// Hooks is the base interface every hook set implements. A concrete hook set
// additionally implements any subset of AuthHook/RecordHook/StreamHook/
// WriteHook/CheckHook for the specific extension points it needs (design
// §B.7 Tier 2); the engine type-asserts for each at the relevant dispatch
// point.
type Hooks interface {
	ConnectorName() string
}

// AuthHook resolves a connsdk.Authenticator for an AuthSpec whose mode is
// "custom" (e.g. GitHub App JWT->installation-token exchange, AWS SigV4).
type AuthHook interface {
	Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, spec AuthSpec) (connsdk.Authenticator, error)
}

// RecordHook post-processes a single record beyond the declarative
// projection: raw is the untouched extracted record, projected is the
// result after schema-driven projection/computed_fields. Returning
// keep=false drops the record; keep=true with a possibly-mutated record
// emits it.
type RecordHook interface {
	MapRecord(stream string, raw, projected connsdk.Record) (connsdk.Record, bool, error)
}

// StreamHook overrides the entire read of one stream (async report jobs,
// CSV downloads, sub-resource fan-out). handled=false tells the engine to
// fall back to the declarative read path.
type StreamHook interface {
	ReadStream(ctx context.Context, stream StreamSpec, req connectors.ReadRequest, rt *Runtime, emit func(connectors.Record) error) (handled bool, err error)
}

// WriteHook overrides execution of one write action (compound/multi-request
// actions such as github's create_pull_request + reviewer follow-up).
// handled=false tells the engine to fall back to the declarative write path.
type WriteHook interface {
	ExecuteWrite(ctx context.Context, action WriteAction, rec connectors.Record, rt *Runtime) (handled bool, err error)
}

// CheckHook overrides the connector's Check(). handled=false tells the
// engine to fall back to the declarative check request.
type CheckHook interface {
	Check(ctx context.Context, cfg connectors.RuntimeConfig, rt *Runtime) (handled bool, err error)
}

// hookRegistry is the process-global hook registry. It lives in engine
// (rather than a separate package) to avoid an import cycle: hooks/<name>
// packages need to reference engine types (AuthSpec, StreamSpec, ...) in
// their method signatures, so engine cannot import them back.
var hookRegistry = struct {
	mu        sync.RWMutex
	factories map[string]func() Hooks
}{factories: make(map[string]func() Hooks)}

// RegisterHooks registers a hook-set factory under name. It is intended to
// be called from a hooks/<name> package's init(); the generated
// hooks/hookset/hookset_gen.go blank-imports each hooks package to run those
// init() side effects. Re-registering an existing name overwrites its factory:
// the most recently registered factory wins.
func RegisterHooks(name string, factory func() Hooks) {
	hookRegistry.mu.Lock()
	defer hookRegistry.mu.Unlock()
	hookRegistry.factories[name] = factory
}

// HooksFor returns a freshly constructed Hooks for name, or nil when no
// hook set is registered under that name. Callers (selectAuth, connector.go)
// must treat a nil return as "no hooks available" rather than an error by
// itself; a hook-requiring spec that finds no hooks is the caller's error to
// raise (e.g. auth.go's missing-hook error).
func HooksFor(name string) Hooks {
	hookRegistry.mu.RLock()
	factory, ok := hookRegistry.factories[name]
	hookRegistry.mu.RUnlock()
	if !ok {
		return nil
	}
	return factory()
}

// unregisterHooks removes a previously registered hook factory. It exists
// for test cleanup so process-global registration does not leak between
// tests.
func unregisterHooks(name string) {
	hookRegistry.mu.Lock()
	defer hookRegistry.mu.Unlock()
	delete(hookRegistry.factories, name)
}
