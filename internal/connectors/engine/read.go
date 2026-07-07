package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

// defaultPageSize is used when neither the stream's nor the base pagination
// spec declares an explicit page_size.
const defaultPageSize = 100

// Read executes a declarative stream read against b per design §B.4: resolve
// the StreamSpec, build a connsdk.Requester (base URL/headers/auth), build
// the initial query (static query + incremental lower bound), drive the
// paginator loop, and for every extracted record: filter -> project
// (+computed_fields) -> hook dispatch -> emit. A StreamHook that returns
// handled=true bypasses the declarative path entirely for this stream.
func Read(ctx context.Context, b Bundle, req connectors.ReadRequest, h Hooks, emit func(connectors.Record) error) error {
	return ReadWithSleeper(ctx, b, req, h, emit, nil)
}

// ReadWithSleeper is Read with an injectable rate-limit sleeper (nil uses the
// real connsdk default, i.e. context-aware time.Sleep). Exposed so tests can
// assert the rate-limit wait is invoked without incurring real sleeps.
func ReadWithSleeper(ctx context.Context, b Bundle, req connectors.ReadRequest, h Hooks, emit func(connectors.Record) error, sleeper func(context.Context, time.Duration) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	stream, err := findStream(b, req.Stream)
	if err != nil {
		return err
	}

	req.Config = materializeConfigDefaults(b, req.Config)

	rt, err := newRuntime(ctx, b, req.Config, h)
	if err != nil {
		return err
	}
	if sleeper != nil {
		rt.Requester.Sleep = sleeper
	}

	if sh, ok := h.(StreamHook); ok {
		handled, err := sh.ReadStream(ctx, stream, req, rt, emit)
		if err != nil {
			return &Error{Connector: b.Name, Stream: stream.Name, Page: -1, RecordIndex: -1, Err: err}
		}
		if handled {
			return nil
		}
	}

	return readDeclarative(ctx, b, stream, req, rt, h, emit)
}

func readDeclarative(ctx context.Context, b Bundle, stream StreamSpec, req connectors.ReadRequest, rt *Runtime, h Hooks, emit func(connectors.Record) error) error {
	if stream.FanOut != nil {
		return readFanOut(ctx, b, stream, req, rt, h, emit)
	}
	return readOneSequence(ctx, b, stream, req, rt, h, emit, fanoutContext{})
}

// fanoutContext carries the current fan-out id (if any) into a single
// request/pagination/incremental sequence (S4 engine mini-wave item 2):
// zero-valued (id == "") for an ordinary, non-fan-out stream read, so
// readOneSequence's behavior is byte-for-byte identical to the pre-fan-out
// implementation when a stream declares no fan_out block at all.
type fanoutContext struct {
	id         string
	queryParam string // FanOutInto.QueryParam, "" when unused
	stampField string // FanOutSpec.StampField, "" when unset
}

// readFanOut resolves the fan-out id list EXACTLY ONCE (config CSV or one
// preliminary paginated request), then runs the ENTIRE declarative
// request/pagination/incremental/filter/project/computed_fields/hook
// sequence unchanged, once per id, via readOneSequence. Pagination,
// incremental state, MaxPages, and rate-limiting are independent PER id
// sub-sequence — the fan-out itself introduces no shared page-count/cursor
// state across ids, mirroring how every quarantined connector's own
// per-parent-id harvest loop behaves.
func readFanOut(ctx context.Context, b Bundle, stream StreamSpec, req connectors.ReadRequest, rt *Runtime, h Hooks, emit func(connectors.Record) error) error {
	fo := stream.FanOut
	ids, err := resolveFanOutIDs(ctx, b, stream, req, rt)
	if err != nil {
		return &Error{Connector: b.Name, Stream: stream.Name, Page: -1, RecordIndex: -1, Err: err}
	}

	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return err
		}
		fc := fanoutContext{id: id, queryParam: fo.Into.QueryParam, stampField: fo.StampField}
		if err := readOneSequence(ctx, b, stream, req, rt, h, emit, fc); err != nil {
			return err
		}
	}
	return nil
}

// resolveFanOutIDs resolves stream.FanOut.IDsFrom's id list, exactly one of
// two mutually-exclusive forms: ConfigKey (a comma-separated config value,
// trimmed and empty-entries-dropped) or Request (one preliminary GET,
// paginated to exhaustion using the stream's OWN effective pagination spec,
// extracting IDField off every record found at RecordsPath). Declaring
// both, or neither, is a read-time error — mirroring
// PaginationSpec.token_path/last_record_field's identical mutual-exclusivity
// error shape (paginate.go's newCursorPaginator).
func resolveFanOutIDs(ctx context.Context, b Bundle, stream StreamSpec, req connectors.ReadRequest, rt *Runtime) ([]string, error) {
	fo := stream.FanOut
	hasConfigKey := fo.IDsFrom.ConfigKey != ""
	hasRequest := fo.IDsFrom.Request != nil

	switch {
	case hasConfigKey && hasRequest:
		return nil, fmt.Errorf("fan_out: ids_from: config_key and request are mutually exclusive")
	case hasConfigKey:
		raw := req.Config.Config[fo.IDsFrom.ConfigKey]
		return splitTrimmedCSV(raw), nil
	case hasRequest:
		return fanOutIDsFromRequest(ctx, b, stream, req, rt, fo.IDsFrom.Request)
	default:
		return nil, fmt.Errorf("fan_out: ids_from: exactly one of config_key or request is required")
	}
}

// splitTrimmedCSV splits a comma-separated config value into a trimmed,
// empty-entry-dropped id list — an empty/whitespace-only input yields a nil
// (zero-length) slice, not an error: an empty configured id list (e.g. no
// sub-resources configured yet) is a legitimate "emit nothing" outcome, not
// a failure.
func splitTrimmedCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// fanOutIDsFromRequest issues the preliminary id-listing GET declared by
// spec, paginated to exhaustion via the CHILD stream's own effective
// pagination spec (base or stream-level override — the id-listing request
// has no pagination block of its own; it reuses the surrounding stream's,
// exactly like the surrounding stream's own per-page requests do), and
// extracts spec.IDField off every record found at spec.RecordsPath.
func fanOutIDsFromRequest(ctx context.Context, b Bundle, stream StreamSpec, req connectors.ReadRequest, rt *Runtime, spec *FanOutIDsRequest) ([]string, error) {
	pag := stream.Pagination
	if pag == nil {
		pag = b.HTTP.Pagination
	}
	pageSize := defaultPageSize
	if pag != nil && pag.PageSize > 0 {
		pageSize = pag.PageSize
	}
	var specForPaginator PaginationSpec
	if pag != nil {
		specForPaginator = *pag
	}
	paginator, err := newPaginator(specForPaginator, pageSize, spec.RecordsPath)
	if err != nil {
		return nil, fmt.Errorf("fan_out: ids_from.request: %w", err)
	}
	if setter, ok := paginator.(baseHostSetter); ok {
		scheme, host := requesterOrigin(rt.Requester.BaseURL)
		setter.setBaseOrigin(scheme, host)
	}

	var ids []string
	page := paginator.Start()
	for pageNum := 0; page != nil; pageNum++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if specForPaginator.MaxPages > 0 && pageNum >= specForPaginator.MaxPages {
			break
		}

		var reqPath string
		if page.URL != "" {
			reqPath = page.URL
		} else {
			resolved, err := InterpolatePath(spec.Path, requestVars(req.Config, nil, ""))
			if err != nil {
				return nil, fmt.Errorf("fan_out: ids_from.request: resolve path: %w", err)
			}
			reqPath = resolved
		}

		resp, err := rt.Requester.Do(ctx, "GET", reqPath, page.Query, nil)
		if err != nil {
			return nil, fmt.Errorf("fan_out: ids_from.request: %w", err)
		}
		records, err := connsdk.RecordsAt(resp.Body, spec.RecordsPath)
		if err != nil {
			return nil, fmt.Errorf("fan_out: ids_from.request: %w", err)
		}
		for _, rec := range records {
			v, present := rec[spec.IDField]
			if !present || v == nil {
				continue
			}
			ids = append(ids, stringify(v))
		}

		page = paginator.Next(resp, len(records))
	}
	if guard, ok := paginator.(interface{ Err() error }); ok {
		if err := guard.Err(); err != nil {
			return nil, fmt.Errorf("fan_out: ids_from.request: %w", err)
		}
	}
	return ids, nil
}

// readOneSequence runs ONE full declarative request/pagination/incremental/
// filter/project/computed_fields/hook sequence for stream — the entire body
// of the pre-fan-out readDeclarative, unchanged in every respect except for
// two additions gated on fc.id != "" (S4 engine mini-wave item 2): (1) the
// path/query Vars carry fc.id as "{{ fanout.id }}" and, when fc.queryParam is
// set, that param is added to every page's query; (2) fc.stampField, when
// set, is written onto every emitted record after projection/
// computed_fields — implemented as an ENGINE-added computed field the
// bundle author never declares twice, applied via the exact same
// applyComputedFields code path as any other computed_fields entry.
func readOneSequence(ctx context.Context, b Bundle, stream StreamSpec, req connectors.ReadRequest, rt *Runtime, h Hooks, emit func(connectors.Record) error, fc fanoutContext) error {
	schema := b.Schemas[stream.Name]

	pag := stream.Pagination
	if pag == nil {
		pag = b.HTTP.Pagination
	}
	pageSize := defaultPageSize
	if pag != nil && pag.PageSize > 0 {
		pageSize = pag.PageSize
	}
	var specForPaginator PaginationSpec
	if pag != nil {
		specForPaginator = *pag
	}
	paginator, err := newPaginator(specForPaginator, pageSize, recordsPathOf(stream.Records))
	if err != nil {
		return &Error{Connector: b.Name, Stream: stream.Name, Page: -1, RecordIndex: -1, Err: err}
	}
	if setter, ok := paginator.(baseHostSetter); ok {
		scheme, host := requesterOrigin(rt.Requester.BaseURL)
		setter.setBaseOrigin(scheme, host)
	}

	baseQuery, err := buildInitialQuery(stream, req)
	if err != nil {
		return &Error{Connector: b.Name, Stream: stream.Name, Page: -1, RecordIndex: -1, Err: err}
	}
	if fc.id != "" && fc.queryParam != "" {
		baseQuery = cloneAndSetQuery(baseQuery, fc.queryParam, fc.id)
	}

	rateLimit := b.HTTP.RateLimit
	rl := newRateLimiter(rateLimit, rt.Requester.Sleep)

	lowerBound, err := incrementalLowerBoundValue(stream, req)
	if err != nil {
		return &Error{Connector: b.Name, Stream: stream.Name, Page: -1, RecordIndex: -1, Err: err}
	}
	formattedLowerBound := ""
	if lowerBound != "" && stream.Incremental != nil {
		formattedLowerBound, err = formatParam(lowerBound, stream.Incremental.ParamFormat)
		if err != nil {
			return &Error{Connector: b.Name, Stream: stream.Name, Page: -1, RecordIndex: -1, Err: err}
		}
	}

	maxPages := specForPaginator.MaxPages

	pathVars := requestVars(req.Config, nil, "")
	pathVars.FanoutID = fc.id

	page := paginator.Start()
	for pageNum := 0; page != nil; pageNum++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		// Hard request-count cap, independent of page fullness: mirrors
		// connsdk.Harvest's own maxPages parameter (paginate.go:201), checked
		// before issuing the request for this page number. maxPages<=0 (the
		// zero value, i.e. absent/unset) means unbounded — pagination is
		// bounded only by the paginator's own short/empty-page stop signal,
		// same as before this cap existed.
		if maxPages > 0 && pageNum >= maxPages {
			break
		}

		if pageNum > 0 {
			if err := rl.wait(ctx); err != nil {
				return err
			}
		}

		var reqPath string
		if page.URL != "" {
			reqPath = page.URL
		} else {
			// F1/SECURITY-REVIEW.md m3: a stream path is a request path just
			// like a write action's, and must go through InterpolatePath
			// (urlencode-by-default) instead of being sent literally — a
			// paginator-supplied absolute URL (page.URL, above) is already
			// fully resolved and must NOT be re-interpolated.
			resolved, err := InterpolatePath(stream.Path, pathVars)
			if err != nil {
				return &Error{Connector: b.Name, Stream: stream.Name, Page: pageNum, RecordIndex: -1, Err: fmt.Errorf("resolve stream path: %w", err)}
			}
			reqPath = resolved
		}
		query := mergeQuery(baseQuery, page.Query)
		body, err := buildStreamRequestBody(stream, req.Config, req.Query, page, specForPaginator, formattedLowerBound, fc)
		if err != nil {
			return &Error{Connector: b.Name, Stream: stream.Name, Page: pageNum, RecordIndex: -1, Err: err}
		}

		resp, err := rt.Requester.Do(ctx, methodOrDefault(stream.Method), reqPath, query, body)
		if err != nil {
			class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
			return &Error{Connector: b.Name, Stream: stream.Name, Page: pageNum, RecordIndex: -1, Class: class, Hint: hint, Err: err}
		}
		if stream.GraphQL != nil {
			if err := graphQLErrors(resp.Body); err != nil {
				return &Error{Connector: b.Name, Stream: stream.Name, Page: pageNum, RecordIndex: -1, Err: err}
			}
		}

		rawRecords, err := extractRecords(resp.Body, stream.Records)
		if err != nil {
			return &Error{Connector: b.Name, Stream: stream.Name, Page: pageNum, RecordIndex: -1, Err: err}
		}
		responseFields, err := extractResponseFields(resp.Body, stream.ResponseFields)
		if err != nil {
			return &Error{Connector: b.Name, Stream: stream.Name, Page: pageNum, RecordIndex: -1, Err: err}
		}

		for _, raw := range rawRecords {
			if err := ctx.Err(); err != nil {
				return err
			}
			raw = mergeResponseFields(raw, responseFields)
			if !passesFilter(raw, stream.Records.Filter) {
				continue
			}
			if lowerBound != "" && stream.Incremental != nil && stream.Incremental.ClientFiltered {
				if !clientFilterKeeps(raw, stream.Incremental.CursorField, lowerBound) {
					continue
				}
			}

			projected := projectRecord(raw, schema, stream.Projection)
			if err := applyComputedFields(projected, raw, req.Config.Config, stream.ComputedFields); err != nil {
				return &Error{Connector: b.Name, Stream: stream.Name, Page: pageNum, RecordIndex: -1, Err: err}
			}
			if fc.id != "" && fc.stampField != "" {
				projected[fc.stampField] = fc.id
			}

			out := connectors.Record(projected)
			if rh, ok := h.(RecordHook); ok {
				mapped, keep, err := rh.MapRecord(stream.Name, connsdk.Record(raw), connsdk.Record(projected))
				if err != nil {
					return &Error{Connector: b.Name, Stream: stream.Name, Page: pageNum, RecordIndex: -1, Err: err}
				}
				if !keep {
					continue
				}
				out = connectors.Record(mapped)
			}

			if err := emit(out); err != nil {
				return err
			}
		}

		page = paginator.Next(resp, len(rawRecords))
	}

	if guard, ok := paginator.(interface{ Err() error }); ok {
		if err := guard.Err(); err != nil {
			return &Error{Connector: b.Name, Stream: stream.Name, Page: -1, RecordIndex: -1, Err: err}
		}
	}

	return nil
}

func buildStreamRequestBody(stream StreamSpec, cfg connectors.RuntimeConfig, query map[string]string, page *connsdk.NextPage, pag PaginationSpec, formattedLowerBound string, fc fanoutContext) (any, error) {
	if stream.GraphQL == nil {
		return nil, nil
	}
	var cursor string
	if page != nil && pag.CursorParam != "" {
		cursor = page.Query.Get(pag.CursorParam)
	}
	vars := requestVars(cfg, nil, cursor, query)
	vars.IncrementalLowerBound = formattedLowerBound
	vars.FanoutID = fc.id
	return buildGraphQLPayload(stream.GraphQL, vars)
}

// cloneAndSetQuery returns a copy of base with key set to value — used to
// add the fan-out id as a query param onto baseQuery without mutating a
// url.Values a caller might otherwise still hold a reference to.
func cloneAndSetQuery(base url.Values, key, value string) url.Values {
	out := url.Values{}
	for k, vs := range base {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	out.Set(key, value)
	return out
}

// findStream resolves name against b.Streams. An empty name is only valid
// when the bundle declares exactly one stream (mirrors the pattern of
// single-stream connectors defaulting to their sole stream).
func findStream(b Bundle, name string) (StreamSpec, error) {
	for _, s := range b.Streams {
		if s.Name == name {
			return s, nil
		}
	}
	if name == "" && len(b.Streams) == 1 {
		return b.Streams[0], nil
	}
	return StreamSpec{}, fmt.Errorf("engine: stream %q not found in bundle %q", name, b.Name)
}

// materializeConfigDefaults returns cfg with every spec.json-declared
// "default" value filled in for a key genuinely ABSENT from cfg.Config (gap-
// loop cycle-1 item 6, REVIEW-A.md C3): the single mechanism for every
// legacy connector's in-code base-URL default/derivation (github
// api.github.com, gmail base+token URLs, monday api.monday.com/v2,
// chargebee/sentry site/hostname-derived, stripe's own base_url default) —
// before this, EVERY migrated connector hard-errored on a config shape
// legacy accepted (an unset base_url), since the engine never read
// spec.json's "default" annotation back out anywhere.
//
// A key already present in cfg.Config (even as an explicit empty string —
// "explicitly set to empty" is a caller decision, not absence) is NEVER
// overridden: defaults only fill genuinely missing keys, exactly like every
// other "default" semantics (JSON Schema's own definition, and legacy's own
// in-code `if cfg.BaseURL == "" { cfg.BaseURL = defaultBaseURL }` pattern —
// though note legacy's pattern typically treated an EMPTY string the same as
// absent, which this narrower "only when the key itself is missing" rule
// deliberately does not replicate; an explicit empty-string config value is
// preserved as-is rather than silently defaulted, since RuntimeConfig.Config
// is a plain map with no separate "was this key set" bit beyond key presence).
// b.Spec == nil (no compiled spec.json at all, e.g. many ad hoc test
// bundles) is a no-op — cfg is returned unchanged.
func materializeConfigDefaults(b Bundle, cfg connectors.RuntimeConfig) connectors.RuntimeConfig {
	if b.Spec == nil {
		return cfg
	}
	defaults := b.Spec.Defaults()
	if len(defaults) == 0 {
		return cfg
	}
	merged := make(map[string]string, len(defaults)+len(cfg.Config))
	for k, v := range defaults {
		merged[k] = v
	}
	for k, v := range cfg.Config {
		merged[k] = v
	}
	cfg.Config = merged
	return cfg
}

// newRuntime builds the shared *Runtime (Requester + Bundle + Config) used
// by both the declarative path and any dispatched hooks. ctx is threaded
// into selectAuth (F8) so a "custom" AuthHook's Authenticator resolution
// (e.g. a network token exchange) honors the caller's context instead of
// running under context.Background().
func newRuntime(ctx context.Context, b Bundle, cfg connectors.RuntimeConfig, h Hooks) (*Runtime, error) {
	baseURL, err := Interpolate(b.HTTP.URL, requestVars(cfg, nil, ""))
	if err != nil {
		return nil, fmt.Errorf("engine: resolve base url: %w", err)
	}

	// An empty auth list means the bundle declares no authentication scheme at
	// all (e.g. a fully public API, or a test double) — selectAuth itself
	// requires at least one candidate spec, so that case is handled here
	// rather than forcing every caller to declare a trivial "none" rule.
	var auth connsdk.Authenticator
	if len(b.HTTP.Auth) > 0 {
		auth, err = selectAuth(ctx, cfg, b.HTTP.Auth, h)
		if err != nil {
			return nil, fmt.Errorf("engine: %w", err)
		}
	}

	headers, err := resolveHeaders(b.HTTP.Headers, cfg, b.Spec)
	if err != nil {
		return nil, err
	}

	requester := &connsdk.Requester{
		BaseURL:        baseURL,
		Auth:           auth,
		UserAgent:      b.HTTP.UserAgent,
		DefaultHeaders: headers,
	}

	return &Runtime{Requester: requester, Bundle: &b, Config: cfg}, nil
}

// resolveHeaders interpolates every declared header value, omitting a
// header whose resolved value is empty (matches stripe's conditional
// Stripe-Account header, SPEC §1.1) or whose template references an
// absent config key that spec DECLARES as optional (not in required[]).
//
// F4 (REVIEW.md flag / SECURITY-REVIEW.md finding): the prior version
// swallowed ANY unresolved-key error uniformly, so
// `Authorization: Bearer {{ secrets.token }}` with no token secret
// configured was silently omitted — the request went out unauthenticated
// instead of failing. The decision table now is:
//   - an unresolved key in the "secrets" namespace is ALWAYS a hard error
//     (never send a request that was meant to carry a secret-derived header
//     but silently didn't);
//   - an unresolved key in the "config" namespace is omitted ONLY when spec
//     declares that key as an optional property (present in spec.Properties()
//     but absent from spec.RequiredKeys()) — the Stripe-Account/account_id
//     pattern; a REQUIRED key, or a key spec doesn't declare at all, is a
//     hard error;
//   - when spec is nil (a bundle/test with no compiled spec.json at all —
//     common in ad hoc test bundles), a config-namespace absence is still
//     omitted, preserving prior behavior for that case: there is no spec
//     surface to know required-vs-optional, and a bundle with literally no
//     declared config keys is not the target of this fix (secrets.* is
//     still always a hard error even with a nil spec);
//   - any other interpolation failure (CRLF injection, unknown
//     namespace/filter) still propagates unchanged.
func resolveHeaders(headers map[string]string, cfg connectors.RuntimeConfig, spec *Schema) (map[string]string, error) {
	if len(headers) == 0 {
		return nil, nil
	}

	var optionalConfigKeys map[string]bool
	if spec != nil {
		required := make(map[string]bool, len(spec.RequiredKeys()))
		for _, k := range spec.RequiredKeys() {
			required[k] = true
		}
		optionalConfigKeys = make(map[string]bool, len(spec.Properties()))
		for _, k := range spec.Properties() {
			if !required[k] {
				optionalConfigKeys[k] = true
			}
		}
	}

	out := make(map[string]string, len(headers))
	for k, tmpl := range headers {
		val, err := InterpolateHeader(tmpl, requestVars(cfg, nil, ""))
		if err != nil {
			omit, hardErr := classifyHeaderResolutionError(err, spec, optionalConfigKeys)
			if hardErr != nil {
				return nil, fmt.Errorf("engine: resolve header %q: %w", k, hardErr)
			}
			if omit {
				continue
			}
			return nil, fmt.Errorf("engine: resolve header %q: %w", k, err)
		}
		if val == "" {
			continue
		}
		out[k] = val
	}
	return out, nil
}

// classifyHeaderResolutionError decides, for a single header's interpolation
// error, whether the header should be silently OMITTED (omit=true), or the
// error should hard-fail (hardErr non-nil), or neither applies (both false/
// nil — caller wraps err as-is; this happens for any non-unresolvedKeyError
// interpolation failure, e.g. CRLF injection or an unknown filter/namespace).
func classifyHeaderResolutionError(err error, spec *Schema, optionalConfigKeys map[string]bool) (omit bool, hardErr error) {
	var unresolved *unresolvedKeyError
	if !errors.As(err, &unresolved) {
		return false, nil
	}
	switch {
	case unresolved.Namespace == "secrets":
		return false, fmt.Errorf("%w (a header referencing an absent secret is never silently omitted)", err)
	case unresolved.Namespace == "config" && spec == nil:
		// No spec surface to know required-vs-optional: preserve prior
		// omit-on-absent-config behavior for bundles/tests with no compiled
		// spec.json at all.
		return true, nil
	case unresolved.Namespace == "config" && optionalConfigKeys[unresolved.Key]:
		// Declared-optional config key, absent at runtime: omit (the
		// Stripe-Account/account_id pattern).
		return true, nil
	default:
		return false, err
	}
}

// requestVars builds the interpolation environment shared by base URL,
// headers, query, and path resolution.
func requestVars(cfg connectors.RuntimeConfig, record map[string]any, cursor string, query ...map[string]string) Vars {
	var q map[string]string
	if len(query) > 0 {
		q = query[0]
	}
	return Vars{Config: cfg.Config, Secrets: cfg.Secrets, Record: record, Cursor: cursor, Query: q}
}

// buildInitialQuery resolves the incremental lower bound (state cursor,
// falling back to start_config_key) formatted per param_format FIRST, then
// stream.Query (static, templated values) against Vars that already carry
// that resolved value as "{{ incremental.lower_bound }}" (S3 engine
// mini-wave item 1, wave1-pilot SUMMARY.md carried queue / REVIEW-A.md
// re-review R2): this is what makes a param legacy sends ONLY when the
// incremental lower bound resolves (chargebee's sort_by[asc]=updated_at,
// chargebee.go:151-155 — set in the SAME branch as updated_at[after], on
// every state-cursor-driven repeat sync as well as a start_date-seeded fresh
// sync) expressible via the existing optional-query dialect. Ordering matters
// here: the lower bound MUST be computed before stream.Query's own resolution
// loop runs, since that loop's Vars is what a "{{ incremental.lower_bound }}"
// template reads.
//
// Per-entry dialect (gap-loop item 3, extended by S3 item 1): a
// plain-string-sourced QueryParam (OmitWhenAbsent false, Default "") keeps
// the exact pre-existing semantics — Interpolate errors hard on any
// unresolved config/secrets/incremental key, with no tolerance at all. An
// object-form entry gets exactly one of two resolutions when its template
// hits an unresolved config/secrets/incremental key: OmitWhenAbsent true
// means the param is left off entirely (no error); otherwise a non-empty
// Default is sent verbatim instead of erroring. An object-form entry with
// NEITHER set behaves identically to a plain string entry (hard error on
// unresolved key) — declaring the object shape alone changes nothing.
func buildInitialQuery(stream StreamSpec, req connectors.ReadRequest) (url.Values, error) {
	lower, err := incrementalLowerBoundValue(stream, req)
	if err != nil {
		return nil, err
	}

	var formattedLower string
	if lower != "" && stream.Incremental != nil {
		formattedLower, err = formatParam(lower, stream.Incremental.ParamFormat)
		if err != nil {
			return nil, err
		}
	}

	vars := requestVars(req.Config, nil, "")
	vars.IncrementalLowerBound = formattedLower
	q, err := resolveQueryParams(stream.Query, vars)
	if err != nil {
		return nil, err
	}

	for k, v := range req.Query {
		q.Set(k, v)
	}
	if formattedLower != "" && stream.Incremental.RequestParam != "" {
		q.Set(stream.Incremental.RequestParam, formattedLower)
	}
	return q, nil
}

// resolveQueryParams resolves every entry of params against vars, applying
// the SAME per-entry dialect stream.Query has always used (bundle.go's
// QueryParam doc comment): a plain-string-sourced entry (OmitWhenAbsent
// false, Default "") hard-errors on any unresolved config/secrets/
// incremental key; an object-form entry tolerates an unresolved key via
// OmitWhenAbsent (param dropped, no error) or Default (literal value sent
// instead of erroring). Shared by buildInitialQuery (stream.Query) and
// buildCheckQuery (RequestSpec.Query, checkquery-ledger.md) so both surfaces
// resolve query templates identically by construction, not by convention.
func resolveQueryParams(params map[string]QueryParam, vars Vars) (url.Values, error) {
	q := url.Values{}
	for k, param := range params {
		val, err := Interpolate(param.Template, vars)
		if err != nil {
			if isUnresolvedConfigSecretOrIncremental(err) {
				switch {
				case param.OmitWhenAbsent:
					continue
				case param.Default != "":
					q.Set(k, param.Default)
					continue
				}
			}
			return nil, fmt.Errorf("engine: resolve query %q: %w", k, err)
		}
		q.Set(k, val)
	}
	return q, nil
}

// buildCheckQuery resolves check.query (RequestSpec.Query) against cfg using
// the identical resolveQueryParams semantics stream.Query uses — see that
// function's doc comment. A nil/empty Query returns an empty url.Values
// (Check() sends no query string at all, exactly as before this dialect
// existed).
func buildCheckQuery(check *RequestSpec, cfg connectors.RuntimeConfig) (url.Values, error) {
	if len(check.Query) == 0 {
		return nil, nil
	}
	vars := requestVars(cfg, nil, "")
	q, err := resolveQueryParams(check.Query, vars)
	if err != nil {
		return nil, fmt.Errorf("engine: resolve check query: %w", err)
	}
	return q, nil
}

// incrementalLowerBoundValue returns the raw (unformatted) incremental lower
// bound: the state cursor if set, else the start_config_key config value,
// else "" (full sync / no lower bound). This value is NOT always RFC3339
// (N4, wave0 REVIEW.md re-review): the state cursor may be an all-digits
// Unix-seconds string (the app-persisted shape for a numeric cursor field,
// B1) or an RFC3339 timestamp (a string cursor field, or a config
// start_date); formatParam/parseLowerBoundTime accept both shapes.
// client_filtered streams (no server-side request_param) still need this
// value to drop old records client-side.
func incrementalLowerBoundValue(stream StreamSpec, req connectors.ReadRequest) (string, error) {
	if stream.Incremental == nil {
		return "", nil
	}
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor, nil
	}
	if stream.Incremental.StartConfigKey == "" {
		return "", nil
	}
	return strings.TrimSpace(req.Config.Config[stream.Incremental.StartConfigKey]), nil
}

// formatParam formats a lower-bound cursor value per param_format
// (rfc3339|unix_seconds|date|github_date_range). An empty format defaults to
// rfc3339 (send the value verbatim).
//
// The state cursor this function receives is NOT always RFC3339: the app
// layer (internal/app/sync_modes.go recordCursor -> toComparableString)
// persists a stream's cursor as the STRINGIFIED RECORD FIELD VALUE, which
// for a numeric field (e.g. Stripe's Unix-seconds "created") is a bare
// digits string like "1700000100", never converted to RFC3339 (B1,
// REVIEW.md). formatParam therefore accepts BOTH input shapes for every
// format that needs to parse a timestamp: an all-digits value is treated as
// already-Unix-seconds (matching legacy's verbatim-forward semantics, and
// the ONLY shape production code actually produces for a numeric cursor
// field); any other value is parsed as RFC3339, exactly as before. rfc3339
// itself never parses at all (verbatim passthrough either way).
func formatParam(value, format string) (string, error) {
	switch format {
	case "", "rfc3339":
		return value, nil
	case "unix_seconds":
		t, err := parseLowerBoundTime(value)
		if err != nil {
			return "", fmt.Errorf("engine: param_format unix_seconds: %w", err)
		}
		return strconv.FormatInt(t.Unix(), 10), nil
	case "date":
		t, err := parseLowerBoundTime(value)
		if err != nil {
			return "", fmt.Errorf("engine: param_format date: %w", err)
		}
		return t.Format("2006-01-02"), nil
	case "github_date_range":
		// GitHub's date-range query-qualifier shape for a lower-bound-only
		// range (design doc's workflow_runs example declares no upper bound).
		// A digits-only (Unix-seconds) input is normalized to RFC3339 first so
		// the emitted qualifier is always a valid GitHub date-range value
		// regardless of which shape the state cursor arrived in.
		t, err := parseLowerBoundTime(value)
		if err != nil {
			return "", fmt.Errorf("engine: param_format github_date_range: %w", err)
		}
		return ">=" + t.UTC().Format(time.RFC3339), nil
	default:
		return "", fmt.Errorf("engine: unknown param_format %q", format)
	}
}

// dateOnlyLayout is the bare YYYY-MM-DD layout (S4 engine mini-wave item 5):
// marketstack's real wire cursor value for its "date" param_format streams
// (eod/splits/dividends) is this shape with no time/offset component at all
// — matches Go's time.DateOnly constant value, spelled out here since this
// package's minimum Go version predates that constant's introduction being
// guaranteed available everywhere this repo builds.
const dateOnlyLayout = "2006-01-02"

// parseLowerBoundTime parses value as one of three shapes, tried in this
// order so no shape masks another (a valid RFC3339 string is never
// all-digits; a valid bare-date string is never RFC3339-parseable — RFC3339
// always requires at least a "T" time-of-day separator):
//  1. a bare Unix-seconds digits string (the app-persisted cursor shape for
//     a numeric cursor field, B1);
//  2. a full RFC3339 timestamp (a string cursor field, or a config
//     start_date);
//  3. a bare YYYY-MM-DD date-only string (S4 engine mini-wave item 5:
//     marketstack's real wire cursor shape for its "date" param_format
//     streams — no time/offset component at all), parsed as midnight UTC
//     that date.
//
// This applies uniformly across every param_format that calls
// parseLowerBoundTime (unix_seconds/date/github_date_range), not just
// "date" — a date-only lower bound is equally valid input for any of them.
func parseLowerBoundTime(value string) (time.Time, error) {
	if isAllDigits(value) {
		secs, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid unix-seconds value %q: %w", value, err)
		}
		return time.Unix(secs, 0).UTC(), nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	if t, err := time.Parse(dateOnlyLayout, value); err == nil {
		return t.UTC(), nil
	}
	return time.Time{}, fmt.Errorf("invalid RFC3339 or date-only (YYYY-MM-DD) value %q", value)
}

// isAllDigits reports whether s is non-empty and consists only of ASCII
// digits (optionally a leading '-' for a pre-epoch Unix timestamp).
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	i := 0
	if s[0] == '-' {
		i = 1
	}
	if i == len(s) {
		return false
	}
	for ; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// recordsPathOf returns the dotted path RecordsAt should use, defaulting the
// empty path (equivalent to ".") for streams that only set single_object.
func recordsPathOf(spec RecordsSpec) string {
	return spec.Path
}

// extractRecords extracts records from a page body per spec: connsdk's
// ordinary array/single-object extraction (RecordsAt), or — when
// spec.KeyedObject is set (S4 engine mini-wave item 3: appfigures/
// alpha-vantage/exchange-rates-shaped APIs whose list endpoint returns a
// JSON OBJECT keyed by an arbitrary id rather than an array) — the
// keyed-object flatten (recordsAtKeyed) instead.
func extractRecords(body []byte, spec RecordsSpec) ([]connsdk.Record, error) {
	if spec.KeyedObject {
		return recordsAtKeyed(body, recordsPathOf(spec), spec.KeyField)
	}
	return connsdk.RecordsAt(body, recordsPathOf(spec))
}

func extractResponseFields(body []byte, fields map[string]string) (map[string]any, error) {
	if len(fields) == 0 {
		return nil, nil
	}
	root, err := decodeJSONKeyed(body)
	if err != nil {
		return nil, err
	}
	out := make(map[string]any, len(fields))
	for name, path := range fields {
		if val := selectPathKeyed(root, path); val != nil {
			out[name] = val
		}
	}
	return out, nil
}

func mergeResponseFields(raw map[string]any, fields map[string]any) map[string]any {
	if len(fields) == 0 {
		return raw
	}
	out := make(map[string]any, len(raw)+len(fields))
	for k, v := range raw {
		out[k] = v
	}
	for k, v := range fields {
		out[k] = v
	}
	return out
}

// recordsAtKeyed selects the JSON object found at path in body and explodes
// EACH VALUE into its own connsdk.Record — {"111":{...},"222":{...}} becomes
// 2 records — instead of connsdk.RecordsAt's ordinary behavior of returning
// a bare object as ONE record. A value that does not itself decode as a JSON
// object (a scalar, array, or null) is silently skipped, mirroring
// RecordsAt's own tolerance for non-object array elements. When keyField is
// non-empty, the source map key is stamped onto that field of the record
// BEFORE it is returned, so it participates in ordinary schema projection/
// computed_fields like any other raw field. Records are emitted in
// ascending sorted-key order for deterministic output — Go map iteration
// order is randomized, and parity/test stability requires a fixed order.
//
// path selects the object using the SAME dotted-path convention
// connsdk.RecordsAt uses ("" or "." selects the root); connsdk itself is not
// modified (read-only in this task), so this is an engine-local
// reimplementation of connsdk's decode+selectPath plumbing rather than an
// exported connsdk addition.
func recordsAtKeyed(body []byte, path, keyField string) ([]connsdk.Record, error) {
	root, err := decodeJSONKeyed(body)
	if err != nil {
		return nil, err
	}
	node := selectPathKeyed(root, path)
	obj, ok := node.(map[string]any)
	if !ok {
		if node == nil {
			return nil, nil
		}
		return nil, fmt.Errorf("engine: keyed_object: value at path %q is not a JSON object", path)
	}

	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]connsdk.Record, 0, len(keys))
	for _, k := range keys {
		valObj, ok := obj[k].(map[string]any)
		if !ok {
			continue
		}
		rec := make(map[string]any, len(valObj)+1)
		for f, v := range valObj {
			rec[f] = v
		}
		if keyField != "" {
			rec[keyField] = k
		}
		out = append(out, connsdk.Record(rec))
	}
	return out, nil
}

// decodeJSONKeyed decodes body into a generic value, preserving numbers as
// json.Number — an engine-local duplicate of connsdk/extract.go's unexported
// decodeJSON (connsdk is read-only in this task, and that helper is not
// exported).
func decodeJSONKeyed(body []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("engine: decode json: %w", err)
	}
	return v, nil
}

// selectPathKeyed walks a decoded JSON value along a dotted path — an
// engine-local duplicate of connsdk/extract.go's unexported selectPath
// (same read-only-connsdk rationale as decodeJSONKeyed).
func selectPathKeyed(root any, path string) any {
	path = strings.TrimSpace(path)
	if path == "" || path == "." {
		return root
	}
	cur := root
	for _, seg := range strings.Split(path, ".") {
		if seg == "" {
			continue
		}
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur, ok = obj[seg]
		if !ok {
			return nil
		}
	}
	return cur
}

// passesFilter applies filter.field_absent / filter.field_equals to a raw
// record. A nil filter always passes.
func passesFilter(raw map[string]any, filter *FilterSpec) bool {
	if filter == nil {
		return true
	}
	if filter.FieldAbsent != "" {
		if v, present := raw[filter.FieldAbsent]; present && v != nil {
			return false
		}
	}
	for field, want := range filter.FieldEquals {
		if !valuesEqual(raw[field], want) {
			return false
		}
	}
	return true
}

func valuesEqual(a, b any) bool {
	return fmt.Sprint(a) == fmt.Sprint(b)
}

// clientFilterKeeps reports whether raw's cursorField value is strictly
// greater than lowerBound (client_filtered incremental streams with no
// server-side filtering support).
func clientFilterKeeps(raw map[string]any, cursorField, lowerBound string) bool {
	v, ok := raw[cursorField]
	if !ok || v == nil {
		return true
	}
	current := fmt.Sprint(v)
	return connsdk.MaxCursor(lowerBound, current) == current && current != lowerBound
}

// projectRecord builds the emitted record from a raw extracted record: in
// "schema" mode (the default) only schema-declared properties survive;
// "passthrough" keeps every raw field. computed_fields are added afterward
// by applyComputedFields regardless of projection mode.
func projectRecord(raw map[string]any, schema *StreamSchema, projection string) map[string]any {
	if projection == "passthrough" || schema == nil {
		out := make(map[string]any, len(raw))
		for k, v := range raw {
			out[k] = v
		}
		return out
	}
	props := schema.Properties()
	out := make(map[string]any, len(props))
	for _, name := range props {
		if v, ok := raw[name]; ok {
			out[name] = v
		}
	}
	return out
}

// applyComputedFields resolves each computed_fields template against raw
// (not the already-projected record, since computed fields commonly reach
// into nested raw JSON that projection would have dropped) and writes the
// result into projected. A missing intermediate path resolves to an absent
// (nil-safe) value rather than erroring, matching interpolate.go's
// resolveRecordPath semantics for record. references.
//
// cfg is the runtime's Config map (design-adjudication A3/G0, REVIEW-A.md):
// wired into the template Vars so a computed field can stamp a config-scoped
// marker onto every record (e.g. github's dropped `repository` field —
// "{{ config.owner }}/{{ config.repo }}"). Secrets is deliberately NEVER
// passed here (threat-model line, A3): a computed field must never be able
// to copy a secret value into emitted record data, so a `{{ secrets.* }}`
// reference in a computed_fields template continues to hard-error exactly as
// before (resolveRefValue's "secrets" case looks up vars.Secrets, which is
// always nil/empty in this Vars — it therefore always misses and returns the
// same unresolved-key error a template author would see if they mistakenly
// tried it).
//
// Typed extraction (adjudication A1, REVIEW-A.md): when a computed_fields
// template is a SINGLE bare `{{ record.<path> }}` reference — no filter
// stage, no surrounding literal text, no second `{{ }}` segment — the RAW
// (pre-stringify) JSON value found at that record path is copied directly
// into the projected record, preserving its native type (number/bool/null/
// object/array), instead of being forced through Interpolate's
// always-string return type. Any other shape (a filter chain, a mixed
// template with literal text or multiple references, a static literal) is
// UNCHANGED and keeps producing a string via Interpolate, exactly as
// before — this is what makes the increment backward compatible with every
// existing bundle's `join:`/rename/marker-stamp computed_fields.
func applyComputedFields(projected, raw map[string]any, cfg map[string]string, computed map[string]string) error {
	if len(computed) == 0 {
		return nil
	}
	vars := Vars{Record: raw, Config: cfg}
	for name, tmpl := range computed {
		if paths, ok, err := bareCoalesceRecordReferences(tmpl); ok || err != nil {
			if err != nil {
				return fmt.Errorf("engine: computed_fields %q: %w", name, err)
			}
			val, found, err := resolveCoalesceRecordValue(raw, paths)
			if err != nil {
				return fmt.Errorf("engine: computed_fields %q: %w", name, err)
			}
			if !found {
				delete(projected, name)
				continue
			}
			projected[name] = val
			continue
		}

		if path, ok := bareRecordLengthReference(tmpl); ok {
			val, err := resolveRecordPathValue(raw, strings.Split(path, "."))
			if err != nil {
				if isUnresolvedRecordPath(err) {
					// Absent path: omit the field, matching legacy's guarded
					// `if arr, ok := item[k].([]any); ok { rec[name] = len(arr) }`
					// — legacy never stamps a count when the source array key is
					// missing, so emitting 0 here would be a new divergence.
					continue
				}
				return fmt.Errorf("engine: computed_fields %q: %w", name, err)
			}
			// A present array yields its length (0 for an empty array); any
			// present-but-non-array value (including JSON null) is omitted, both
			// mirroring legacy's type-asserted guard exactly.
			if arr, ok := val.([]any); ok {
				projected[name] = len(arr)
			}
			continue
		}

		if path, ok := bareRecordPathReference(tmpl); ok {
			val, err := resolveRecordPathValue(raw, strings.Split(path, "."))
			if err != nil {
				if isUnresolvedRecordPath(err) {
					continue
				}
				return fmt.Errorf("engine: computed_fields %q: %w", name, err)
			}
			projected[name] = val
			continue
		}

		val, err := Interpolate(tmpl, vars)
		if err != nil {
			// A computed field whose source path is absent is intentionally
			// left out of the projected record (not an error): dotted record
			// paths regularly reach into optional nested objects (e.g. a PR
			// vs. issue's differing shape). Any other interpolation error
			// (unknown namespace, CRLF, bad filter) still propagates.
			if isUnresolvedRecordPath(err) {
				continue
			}
			return fmt.Errorf("engine: computed_fields %q: %w", name, err)
		}
		projected[name] = val
	}
	return nil
}

func bareCoalesceRecordReferences(tmpl string) ([]string, bool, error) {
	inner, ok := bareTemplateInner(tmpl)
	if !ok {
		return nil, false, nil
	}
	return coalesceRecordPathsExpression(inner)
}

func bareRecordLengthReference(tmpl string) (path string, ok bool) {
	inner, ok := bareTemplateInner(tmpl)
	if !ok {
		return "", false
	}
	parts := strings.Split(inner, "|")
	if len(parts) != 2 || strings.TrimSpace(parts[1]) != "length" {
		return "", false
	}
	ref := strings.TrimSpace(parts[0])
	const prefix = "record."
	if !strings.HasPrefix(ref, prefix) || ref == prefix {
		return "", false
	}
	return strings.TrimPrefix(ref, prefix), true
}

// bareRecordPathReference reports whether tmpl is EXACTLY one `{{ ... }}`
// template covering the whole string (no surrounding literal text, no second
// `{{ }}` occurrence) whose inner expression is a plain `record.<path>`
// reference with NO filter stage (`|`). When it is, ok is true and path is
// the dotted path after "record." (e.g. "meta.count"). A static literal (no
// `{{ }}` at all), a `cursor`/`config.*`/`secrets.*` bare reference, a
// filtered reference, or any mixed/multi-part template returns ok=false so
// the caller falls through to ordinary string Interpolate.
func bareRecordPathReference(tmpl string) (path string, ok bool) {
	inner, ok := bareTemplateInner(tmpl)
	if !ok {
		return "", false
	}
	if strings.Contains(inner, "|") {
		return "", false
	}
	const prefix = "record."
	if !strings.HasPrefix(inner, prefix) {
		return "", false
	}
	return strings.TrimPrefix(inner, prefix), true
}

func bareTemplateInner(tmpl string) (string, bool) {
	matches := templatePattern.FindAllStringSubmatchIndex(tmpl, -1)
	if len(matches) != 1 {
		return "", false
	}
	m := matches[0]
	if m[0] != 0 || m[1] != len(tmpl) {
		// The single {{ }} match doesn't span the entire template — there is
		// surrounding literal text (a mixed template like "count={{ record.count }}").
		return "", false
	}
	return strings.TrimSpace(tmpl[m[2]:m[3]]), true
}

// isUnresolvedRecordPath reports whether err is the typed unresolvedKeyError
// with Namespace "record" (F4/REVIEW.md: replaced brittle
// strings.Contains(err.Error(), ...) classification with errors.As against
// the typed sentinel from interpolate.go).
func isUnresolvedRecordPath(err error) bool {
	var unresolved *unresolvedKeyError
	return errors.As(err, &unresolved) && unresolved.Namespace == "record"
}

// isUnresolvedConfigSecretOrIncremental reports whether err is the typed
// unresolvedKeyError for the "config", "secrets", or "incremental" namespace
// (gap-loop item 3, extended by S3 item 1: buildInitialQuery's object-form
// QueryParam tolerance is scoped to an absent config/secrets/incremental-
// lower-bound key specifically — any OTHER interpolation failure, e.g. CRLF
// injection or an unknown filter/namespace, still propagates as a hard error
// even for an omit_when_absent/default-bearing entry). "incremental" covers
// exactly one key today ("lower_bound"), which resolves as unresolved/absent
// whenever no incremental lower bound applies (fresh full sync, or no
// incremental spec at all) — see resolveIncrementalRef, interpolate.go.
func isUnresolvedConfigSecretOrIncremental(err error) bool {
	var unresolved *unresolvedKeyError
	if !errors.As(err, &unresolved) {
		return false
	}
	switch unresolved.Namespace {
	case "config", "secrets", "incremental":
		return true
	default:
		return false
	}
}

func methodOrDefault(method string) string {
	if method == "" {
		return "GET"
	}
	return method
}

func mergeQuery(base, extra url.Values) url.Values {
	out := url.Values{}
	for k, vs := range base {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	for k, vs := range extra {
		out.Del(k)
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}

// requesterOrigin returns baseURL's (scheme, host) pair for wiring an
// SSRF-guarded paginator's expected origin (M1/m2: the guard must compare
// scheme AND host, not host alone). An unparseable baseURL yields ("", "")
// — the guard treats an unset host as "no origin configured, do not
// enforce" (matching prior behavior for callers that never had a real base
// URL, e.g. some unit tests), which is safe because Check/Read already fail
// earlier if HTTP.URL itself cannot be interpolated/parsed as a request
// target.
func requesterOrigin(baseURL string) (scheme, host string) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", ""
	}
	return u.Scheme, u.Host
}

// rateLimiter enforces RateLimitSpec{requests_per_minute} as a fixed
// inter-request wait between successive page requests (a simple token-bucket
// approximation: wait exactly the per-request interval before every request
// after the first).
type rateLimiter struct {
	interval time.Duration
	sleep    func(context.Context, time.Duration) error
}

func newRateLimiter(spec *RateLimitSpec, sleep func(context.Context, time.Duration) error) *rateLimiter {
	if spec == nil || spec.RequestsPerMinute <= 0 {
		return &rateLimiter{}
	}
	interval := time.Minute / time.Duration(spec.RequestsPerMinute)
	if sleep == nil {
		sleep = ctxSleepFallback
	}
	return &rateLimiter{interval: interval, sleep: sleep}
}

func (rl *rateLimiter) wait(ctx context.Context) error {
	if rl == nil || rl.interval <= 0 || rl.sleep == nil {
		return nil
	}
	return rl.sleep(ctx, rl.interval)
}

// ctxSleepFallback is the real (non-test) sleep used when no injected
// sleeper is provided and a rate limit is configured.
func ctxSleepFallback(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// InitialState satisfies connectors.StatefulReader generically for every
// engine-backed connector: a fresh sync starts with an empty cursor (full
// sync), which the app layer's incremental-mode config (start_date,
// start_config_key) can raise at read time — mirrors
// stripe.Connector.InitialState (stripe/stripe.go:99).
func InitialState(ctx context.Context, b Bundle, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

// Check executes the bundle's declarative check request (HTTP.Check), or
// dispatches to a CheckHook first when the connector's Hooks implements one
// and reports handled=true.
func Check(ctx context.Context, b Bundle, cfg connectors.RuntimeConfig, h Hooks) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	cfg = materializeConfigDefaults(b, cfg)

	rt, err := newRuntime(ctx, b, cfg, h)
	if err != nil {
		return err
	}

	if ch, ok := h.(CheckHook); ok {
		handled, err := ch.Check(ctx, cfg, rt)
		if err != nil {
			return &Error{Connector: b.Name, Page: -1, RecordIndex: -1, Err: err}
		}
		if handled {
			return nil
		}
	}

	if b.HTTP.Check == nil {
		return nil
	}
	checkPath, err := InterpolatePath(b.HTTP.Check.Path, requestVars(cfg, nil, ""))
	if err != nil {
		return &Error{Connector: b.Name, Page: -1, RecordIndex: -1, Err: fmt.Errorf("resolve check path: %w", err)}
	}
	checkQuery, err := buildCheckQuery(b.HTTP.Check, cfg)
	if err != nil {
		return &Error{Connector: b.Name, Page: -1, RecordIndex: -1, Err: err}
	}
	_, err = rt.Requester.Do(ctx, methodOrDefault(b.HTTP.Check.Method), checkPath, checkQuery, nil)
	if err != nil {
		class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
		return &Error{Connector: b.Name, Page: -1, RecordIndex: -1, Class: class, Hint: hint, Err: err}
	}
	return nil
}
