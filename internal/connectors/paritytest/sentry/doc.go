package sentryparity_test

// Engine-vs-legacy parity suite for the sentry pilot migration (PLAN.md P-5,
// SPEC.md §5.3). Drives BOTH connectors live against the SAME httptest
// server, asserting RAW connectors.Record equality (reflect.DeepEqual) plus
// legacy-side sanity assertions, following the wave0 goldens
// (internal/connectors/engine/parity_stripe_test.go, parity_searxng_test.go).
//
// Red-first protocol (conventions.md, PLAN.md P-1..P-10): this file was
// written FIRST and failed because internal/connectors/defs/sentry and
// internal/connectors/hooks/sentry did not exist yet — see
// traces/p5-sentry-ledger.md for the recorded RED output.
//
// Resolution ladder outcome (SPEC.md §5.3, decided from evidence per the
// ladder's own instruction, not assumed): Sentry's list endpoints ALWAYS
// emit a Link header rel="next" entry (sentry.go:7-9,144-152) — even on the
// truly-last page — with the real "more pages" signal carried in that
// entry's results="true"/"false" attribute, which the engine's link_header
// paginator type does not parse at all (paginate.go's linkHeaderPaginator
// follows rel="next" unconditionally). Rung 1 (Tier-1 link_header + document
// an "at most one extra trailing request with identical records" deviation)
// was tested and REJECTED: conformance's checkPaginationTerminates
// (conformance/dynamic.go) requires the fixture replay server to receive
// EXACTLY one request per recorded fixture page, and any request beyond the
// last fixture page 404s (replay.go's newStreamReplayServer) — connsdk's
// Requester.Do treats any non-2xx response as a hard error (connsdk/http.go),
// so the "extra trailing request" a Tier-1 link_header paginator would
// always issue against Sentry's always-present rel="next" link is not a
// benign extra fetch, it is a hard read failure. There is no fixture shape
// that satisfies both "engine consumes each page exactly once" (conformance's
// invariant) and "Tier-1 pagination stops without engine-side knowledge of
// results=". Rung 1 does not hold. This lands on rung 2: Tier-2 StreamHook
// (hooks/sentry/hooks.go) porting legacy's exact Link/results handling.
//
// Because conformance/dynamic.go's dynamic checks call engine.Read with
// hooks=nil (they exercise the declarative fallback path only, never
// dispatching through a StreamHook), this bundle declares
// streams.json base.pagination as {"type":"none"} — the declarative path is
// never actually taken in production (the StreamHook handles every stream,
// always returning handled=true), so its own conformance-fixture pagination
// contract is trivially satisfied by a single-page fixture. The StreamHook
// is what this parity suite exercises for real 2-page Link-header
// termination, matching PLAN.md P-5's "termination MUST be proven by 2-page
// fixture parity" requirement.
//
// Legacy read in full (internal/connectors/sentry/{sentry,streams}.go,
// read-only reference):
//   - auth: connsdk.Bearer(secret) where secret is the "auth_token" secret
//     (sentry.go:243) -> engine bearer auth mode, token
//     "{{ secrets.auth_token }}".
//   - base URL: https://<hostname> (default hostname sentry.io) or an
//     explicit base_url override; every endpoint path carries the "api/0"
//     prefix (sentry.go:32-34,248-276).
//   - 4 streams, routed by scope (streams.go's sentryStreamEndpoints):
//     projects (scopeGlobal, "projects/"), issues (scopeProject,
//     "projects/{org}/{project}/issues/"), events (scopeProject,
//     "projects/{org}/{project}/events/"), releases (scopeOrg,
//     "organizations/{org}/releases/"). Every endpoint returns a top-level
//     JSON array (records.path: "").
//   - pagination: RFC 5988 Link header, rel="next" ALWAYS present,
//     results="true"/"false" is the real continuation signal (sentry.go's
//     harvest/nextCursor) — ported into the StreamHook exactly.
//   - no incremental cursor is ever sent as a request param anywhere in
//     legacy (CursorFields are published on the Stream/Catalog surface only,
//     e.g. "lastSeen"/"dateCreated" — sentry/streams.go — but neither
//     sentry.go's Read nor harvest ever forwards a state cursor into a
//     request; there is no incrementalLowerBound-equivalent call anywhere in
//     the package). This bundle matches that by declaring NO
//     incremental block at all (full_refresh only, matching legacy's real
//     behavior byte-for-byte — declaring one would be new, unrequested
//     capability, not a migration).
//   - read-only: Write returns connectors.ErrUnsupportedOperation
//     (sentry.go:97-99) -> capabilities.write: false, no writes.json.
//   - Check: GET api/0/projects/?per_page=1, requires organization/project
//     NOT required for Check itself (sentry.go:70-93) — only secret
//     auth_token and a resolvable base URL are required.
import _ "polymetrics.ai/internal/connectors/hooks/sentry" // triggers engine.RegisterHooks("sentry", ...) init side effect for parity, mirroring SPEC.md §6's "Tier-2 pilots blank-import their own hooks/<name> package from the parity test" rule
