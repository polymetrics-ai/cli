package engine

import (
	"context"
	"fmt"
	"strings"
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
	Requester     *connsdk.Requester
	baseRequester *connsdk.Requester
	Bundle        *Bundle
	Config        connectors.RuntimeConfig
	rateLimits    *rateLimitResolver
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

// DeclaredRouteRequest describes the one JSON request an AuthHook needs to
// make through the engine. DeclaredPath is the connector declaration; Path is
// the hook's already-interpolated provider path. It deliberately has no
// coordinator, URL, query, raw transport, or generic request-writer escape
// hatch.
type DeclaredRouteRequest struct {
	Method       string
	DeclaredPath string
	Path         string
	Headers      map[string]string
	Body         any
}

// DeclaredRouteRequester permits a custom AuthHook to make one
// declaration-aware JSON request. The engine owns its implementation so every
// request enters the same admission and observation lifecycle as declarative
// traffic.
type DeclaredRouteRequester interface {
	DoJSON(ctx context.Context, request DeclaredRouteRequest) (*connsdk.Response, error)
}

// DeclaredRouteAuthHook is the optional custom-auth extension for hooks that
// need a provider request during authenticator construction. AuthHook remains
// compatible for hooks that derive an authenticator locally.
type DeclaredRouteAuthHook interface {
	AuthHook
	AuthenticatorWithDeclaredRoute(ctx context.Context, cfg connectors.RuntimeConfig, spec AuthSpec, requester DeclaredRouteRequester) (connsdk.Authenticator, error)
}

// RateLimitAuthProfileHook reports the declared non-secret rate-limit profile
// for a matched authentication spec. The engine asks before constructing the
// authenticator, so network-capable custom auth is admitted before its own
// request.
type RateLimitAuthProfileHook interface {
	RateLimitAuthProfile(cfg connectors.RuntimeConfig, spec AuthSpec) (string, bool)
}

type declaredRouteRequester struct {
	runtime *Runtime
}

// DoJSON resolves a declaration before materializing the physical request.
// Header values are copied into a requester clone and never retained by the
// runtime, resolver, or coordinator. The actual path is intentionally passed
// to connsdk.Requester.Do: endpoint policy admission happens at that physical
// send boundary, including after the #3754 resolved-path change. Automatic
// retries are disabled so a token exchange does not repeat a quota call.
func (r declaredRouteRequester) DoJSON(ctx context.Context, request DeclaredRouteRequest) (*connsdk.Response, error) {
	if r.runtime == nil {
		return nil, fmt.Errorf("auth declared route requester is unavailable")
	}
	method := strings.ToUpper(strings.TrimSpace(request.Method))
	declaredPath := strings.TrimSpace(request.DeclaredPath)
	path := strings.TrimSpace(request.Path)
	if method == "" || !strings.HasPrefix(declaredPath, "/") {
		return nil, fmt.Errorf("auth declared route request requires a method and declaration path")
	}
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return nil, fmt.Errorf("auth declared route request requires a relative absolute path")
	}
	requester, err := r.runtime.RequesterFor(method, declaredPath)
	if err != nil {
		return nil, err
	}
	clone := *requester
	headers := make(map[string]string, len(requester.DefaultHeaders)+len(request.Headers))
	for key, value := range requester.DefaultHeaders {
		headers[key] = value
	}
	for key, value := range request.Headers {
		headers[key] = value
	}
	clone.DefaultHeaders = headers
	clone.DisableRetries = true
	return clone.Do(ctx, method, path, nil, request.Body)
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

type WriteHookClassifier interface {
	HandlesWriteAction(action WriteAction) bool
}

// WriteRecordHook pins the body constants an action's own name implies onto the
// record, before the declarative body is built. handled=false tells the engine
// the action is untouched.
//
// It exists because a WriteHook that overrides execution cannot be destructive:
// prepareDeclarativeWrite refuses to prepare a destructive hook-executed action,
// since the preview an operator approves would not be the request that runs.
// Mapping the record instead keeps the declarative path — preview and execution
// build the same body from the same record — so an action that only needs a
// fixed body field can still carry a typed confirmation.
type WriteRecordHook interface {
	MapWriteRecord(action WriteAction, rec connectors.Record) (connectors.Record, bool, error)
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
