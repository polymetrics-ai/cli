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
	budget        connsdk.BudgetCoordinator
}

// Hooks is the base interface every hook set implements. A concrete hook set
// additionally implements any subset of AuthHook/RecordHook/StreamHook/
// WriteHook/CheckHook for the specific extension points it needs (design
// §B.7 Tier 2); the engine type-asserts for each at the relevant dispatch
// point.
type Hooks interface {
	ConnectorName() string
}

// CommandBindingTransportHook exposes the fixed physical endpoint selected by
// a registered executor hook for one declaration binding. Admission uses it
// only to prove a documented canonical endpoint maps to that exact transport;
// callers cannot provide or alter either value.
type CommandBindingTransportHook interface {
	CommandBindingTransport(binding connectors.CommandBindingIdentity) (method, path string, handled bool)
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
	clone.Auth = nil
	headers := make(map[string]string, len(requester.DefaultHeaders)+len(request.Headers))
	for key, value := range requester.DefaultHeaders {
		headers[key] = value
	}
	for key, value := range request.Headers {
		headers[key] = value
	}
	clone.DefaultHeaders = headers
	clone.DisableRetries = true

	lease, err := r.runtime.reserveDeclaredRouteBudget(ctx, method, path)
	if err != nil {
		return nil, err
	}
	response, requestErr := clone.Do(ctx, method, path, nil, request.Body)
	if lease == "" {
		return response, requestErr
	}
	// A provider attempt may complete at the same moment its caller cancels.
	// Completion must still reach the lease owner so an admitted token request
	// cannot strand an in-flight reservation. The observation is deliberately
	// limited to the safe requester's vocabulary and contains no provider body,
	// headers, route, or auth material.
	finishErr := r.runtime.budget.Finish(context.WithoutCancel(ctx), lease, connsdk.CompletionObservation{Attempted: true})
	if requestErr != nil {
		return nil, requestErr
	}
	if finishErr != nil {
		return nil, fmt.Errorf("auth declared route budget completion: %w", finishErr)
	}
	return response, nil
}

// reserveDeclaredRouteBudget derives an opaque batch only after the auth hook
// has named its own declared request. It owns the one Decide call for an
// eventual physical auth send; a non-grant has no lease and therefore no
// Finish call.
func (rt *Runtime) reserveDeclaredRouteBudget(ctx context.Context, method, path string) (connsdk.RateBudgetLease, error) {
	if rt == nil || rt.budget == nil || rt.rateLimits == nil {
		return "", nil
	}
	batch, err := rt.rateLimits.reservationBatch(ctx, method, path)
	if err != nil {
		return "", err
	}
	if len(batch.Policies) == 0 {
		return "", nil
	}
	decision, err := rt.budget.Decide(ctx, batch)
	if err != nil {
		return "", fmt.Errorf("auth declared route budget decision: %w", err)
	}
	if !decision.Granted || decision.Lease == "" {
		return "", &connsdk.RateBudgetRefusalError{
			Code:   connsdk.RateBudgetRefusalReservationDenied,
			Reason: "lifecycle coordinator did not grant a lease",
		}
	}
	return decision.Lease, nil
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
// actions such as github's create_pull_request + reviewer follow-up). Every
// physical provider response, including the terminal failed one, is returned
// to the engine for bounded receipt projection. handled=false tells the engine
// to fall back to the declarative write path and must return no responses.
type WriteHook interface {
	ExecuteWrite(ctx context.Context, action WriteAction, rec connectors.Record, rt *Runtime) (handled bool, responses []*connsdk.Response, err error)
}

// PreparedWriteHook is the only compound-write extension admitted by the
// approval path. A hook may select ordered, declaration-owned action names and
// sealed typed records; it cannot supply a method, URL, headers, or body. The
// engine validates and prepares every selected physical request before a
// preview is produced, then executes that same private plan after approval.
//
// WriteHook is retained for direct unit compatibility, but a legacy hook that
// claims an action is refused by prepareDeclarativeWrite rather than being
// allowed to choose an unpreviewed request at execution time.
type PreparedWriteHook interface {
	PrepareWrite(action WriteAction, records []connectors.Record) (PreparedWriteHookPlan, bool, error)
}

// PreparedWriteHookPlan groups each source record's physical provider
// requests. Records may have no steps only when the named compound action
// resolves to a declaration-owned no-op; every non-empty step becomes exactly
// one PreparedRequest in the preview's ordered digest projection.
type PreparedWriteHookPlan struct {
	Records []PreparedWriteHookRecord
}

type PreparedWriteHookRecord struct {
	Steps []PreparedWriteHookStep
}

// PreparedWriteHookStep names an existing declaration-owned write action.
// Record is validated against that declaration and sealed by the engine. A
// response binding is intentionally limited to a previous step's one named
// JSON field becoming one declared path field of this step; it is not a raw
// response/body projection language. ResolvedDeclarative allows the root hook
// to explicitly attest that a selected hook-backed action is already a fully
// resolved declarative request, so the engine will not recurse into that hook.
type PreparedWriteHookStep struct {
	Action              string
	Record              connectors.Record
	ResponseBinding     *PreparedWriteResponseBinding
	ResolvedDeclarative bool
}

// PreparedWriteResponseBinding supplies one follow-up path field from one
// earlier physical response in the same source record. SourceStep is a
// zero-based index into that record's ordered hook plan.
type PreparedWriteResponseBinding struct {
	SourceStep  int    `json:"source_step"`
	Field       string `json:"field"`
	TargetField string `json:"target_field"`
}

// PreparedWriteResponseValidator lets a native hook retain an established
// typed provider-envelope rule while the engine owns transport, planning, and
// receipt persistence. It receives only the named action, sealed record, and
// bounded response already captured by the engine.
type PreparedWriteResponseValidator interface {
	ValidatePreparedWriteResponse(action WriteAction, rec connectors.Record, response *connsdk.Response) error
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
