package engine

import (
	"context"
	"fmt"
	"net/url"
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

	rt, err := newRuntime(b, req.Config, h)
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
	paginator, err := newPaginator(specForPaginator, pageSize)
	if err != nil {
		return &Error{Connector: b.Name, Stream: stream.Name, Page: -1, RecordIndex: -1, Err: err}
	}
	if nu, ok := paginator.(*nextURL); ok {
		nu.BaseHost = requesterHost(rt.Requester.BaseURL)
	}

	baseQuery, err := buildInitialQuery(stream, req)
	if err != nil {
		return &Error{Connector: b.Name, Stream: stream.Name, Page: -1, RecordIndex: -1, Err: err}
	}

	rateLimit := b.HTTP.RateLimit
	rl := newRateLimiter(rateLimit, rt.Requester.Sleep)

	lowerBound, err := incrementalLowerBoundValue(stream, req)
	if err != nil {
		return &Error{Connector: b.Name, Stream: stream.Name, Page: -1, RecordIndex: -1, Err: err}
	}

	maxPages := specForPaginator.MaxPages

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

		reqPath := stream.Path
		if page.URL != "" {
			reqPath = page.URL
		}
		query := mergeQuery(baseQuery, page.Query)

		resp, err := rt.Requester.Do(ctx, methodOrDefault(stream.Method), reqPath, query, nil)
		if err != nil {
			class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
			return &Error{Connector: b.Name, Stream: stream.Name, Page: pageNum, RecordIndex: -1, Class: class, Hint: hint, Err: err}
		}

		rawRecords, err := connsdk.RecordsAt(resp.Body, recordsPathOf(stream.Records))
		if err != nil {
			return &Error{Connector: b.Name, Stream: stream.Name, Page: pageNum, RecordIndex: -1, Err: err}
		}

		for _, raw := range rawRecords {
			if err := ctx.Err(); err != nil {
				return err
			}
			if !passesFilter(raw, stream.Records.Filter) {
				continue
			}
			if lowerBound != "" && stream.Incremental != nil && stream.Incremental.ClientFiltered {
				if !clientFilterKeeps(raw, stream.Incremental.CursorField, lowerBound) {
					continue
				}
			}

			projected := projectRecord(raw, schema, stream.Projection)
			if err := applyComputedFields(projected, raw, stream.ComputedFields); err != nil {
				return &Error{Connector: b.Name, Stream: stream.Name, Page: pageNum, RecordIndex: -1, Err: err}
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

// newRuntime builds the shared *Runtime (Requester + Bundle + Config) used
// by both the declarative path and any dispatched hooks.
func newRuntime(b Bundle, cfg connectors.RuntimeConfig, h Hooks) (*Runtime, error) {
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
		auth, err = selectAuth(cfg, b.HTTP.Auth, h)
		if err != nil {
			return nil, fmt.Errorf("engine: %w", err)
		}
	}

	headers, err := resolveHeaders(b.HTTP.Headers, cfg)
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

// resolveHeaders interpolates every declared header value, omitting any
// header whose resolved value is empty (matches stripe's conditional
// Stripe-Account header, SPEC §1.1).
func resolveHeaders(headers map[string]string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if len(headers) == 0 {
		return nil, nil
	}
	out := make(map[string]string, len(headers))
	for k, tmpl := range headers {
		val, err := InterpolateHeader(tmpl, requestVars(cfg, nil, ""))
		if err != nil {
			// An unresolved config/secret key (the value is simply absent —
			// e.g. an optional per-account header like Stripe-Account with no
			// account_id configured) is treated the same as an empty resolved
			// value: the header is omitted. Any other interpolation failure
			// (CRLF injection, unknown namespace/filter) still propagates.
			if isUnresolvedKey(err) {
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

func isUnresolvedKey(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unresolved key")
}

// requestVars builds the interpolation environment shared by base URL,
// headers, query, and path resolution.
func requestVars(cfg connectors.RuntimeConfig, record map[string]any, cursor string) Vars {
	return Vars{Config: cfg.Config, Secrets: cfg.Secrets, Record: record, Cursor: cursor}
}

// buildInitialQuery resolves stream.Query (static, templated values) plus
// the incremental lower bound (state cursor, falling back to
// start_config_key) formatted per param_format.
func buildInitialQuery(stream StreamSpec, req connectors.ReadRequest) (url.Values, error) {
	q := url.Values{}
	vars := requestVars(req.Config, nil, "")
	for k, tmpl := range stream.Query {
		val, err := Interpolate(tmpl, vars)
		if err != nil {
			return nil, fmt.Errorf("engine: resolve query %q: %w", k, err)
		}
		q.Set(k, val)
	}

	lower, err := incrementalLowerBoundValue(stream, req)
	if err != nil {
		return nil, err
	}
	if lower != "" && stream.Incremental != nil && stream.Incremental.RequestParam != "" {
		formatted, err := formatParam(lower, stream.Incremental.ParamFormat)
		if err != nil {
			return nil, err
		}
		q.Set(stream.Incremental.RequestParam, formatted)
	}
	return q, nil
}

// incrementalLowerBoundValue returns the raw (unformatted, always RFC3339
// when present) incremental lower bound: the state cursor if set, else the
// start_config_key config value, else "" (full sync / no lower bound).
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

// formatParam formats an RFC3339 lower-bound value per param_format
// (rfc3339|unix_seconds|date|github_date_range). An empty format defaults to
// rfc3339 (send the value verbatim).
func formatParam(value, format string) (string, error) {
	switch format {
	case "", "rfc3339":
		return value, nil
	case "unix_seconds":
		t, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return "", fmt.Errorf("engine: param_format unix_seconds: invalid RFC3339 value %q: %w", value, err)
		}
		return strconv.FormatInt(t.Unix(), 10), nil
	case "date":
		t, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return "", fmt.Errorf("engine: param_format date: invalid RFC3339 value %q: %w", value, err)
		}
		return t.Format("2006-01-02"), nil
	case "github_date_range":
		// GitHub's date-range query-qualifier shape for a lower-bound-only
		// range (design doc's workflow_runs example declares no upper bound).
		return ">=" + value, nil
	default:
		return "", fmt.Errorf("engine: unknown param_format %q", format)
	}
}

// recordsPathOf returns the dotted path RecordsAt should use, defaulting the
// empty path (equivalent to ".") for streams that only set single_object.
func recordsPathOf(spec RecordsSpec) string {
	return spec.Path
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
func applyComputedFields(projected, raw map[string]any, computed map[string]string) error {
	if len(computed) == 0 {
		return nil
	}
	for name, tmpl := range computed {
		val, err := Interpolate(tmpl, Vars{Record: raw})
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

func isUnresolvedRecordPath(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unresolved key") && strings.Contains(err.Error(), "in record")
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

func requesterHost(baseURL string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	return u.Host
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

	rt, err := newRuntime(b, cfg, h)
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
	_, err = rt.Requester.Do(ctx, methodOrDefault(b.HTTP.Check.Method), b.HTTP.Check.Path, nil, nil)
	if err != nil {
		class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
		return &Error{Connector: b.Name, Page: -1, RecordIndex: -1, Class: class, Hint: hint, Err: err}
	}
	return nil
}
