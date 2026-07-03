# Overview

Stigg exposes ALL of its data through a single GraphQL endpoint (`POST /graphql`) whose request
body carries a GraphQL query string and whose response wraps records under a top-level `data`
object. This is a Tier-2 `StreamHook` migration (quarantine.json's original `ENGINE_GAP` finding):
`internal/connectors/engine/bundle.go`'s `StreamSpec.Body` field exists but
`internal/connectors/engine/read.go`'s declarative read path never sends a body (`read.go`'s
`readOneSequence` always issues `rt.Requester.Do(ctx, method, reqPath, query, nil)` — the body
argument is hard-coded `nil`), so a POST-body GraphQL read cannot be expressed in `streams.json`
alone. `internal/connectors/hooks/stigg/hooks.go` implements `StreamHook`, porting
`internal/connectors/stigg/stigg.go`'s GraphQL query construction and `data.<field>` record
extraction verbatim. This bundle is engine-vs-legacy parity-tested against
`internal/connectors/stigg` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Stigg authenticates via a single `api_key` secret sent as a standard **`Authorization: Bearer
<api_key>`** header on every GraphQL POST — matching legacy's `requester()`
(`stigg.go:108-118`: `connsdk.Bearer(key)`). This bundle wires the identical shape
**declaratively, with no AuthHook needed at all**: `streams.json`'s `base.auth` declares a single
`bearer` candidate (`{"mode":"bearer","token":"{{ secrets.api_key }}"}`). Auth is intentionally NOT
part of `hooks/stigg/hooks.go` — the engine's `connsdk.Requester` (built by `engine.newRuntime`
before any hook runs) already sets the `Authorization: Bearer <api_key>` header on every request
the hook issues via `rt.Requester.Do`, so the hook only needs to build the GraphQL query and POST
it; it never touches auth itself. This is why stigg needed only a `StreamHook`, not the `AuthHook`
interface, despite being filed as a Tier-2 connector in `quarantine.json`.

## Streams notes

Legacy defines 4 streams, each a distinct GraphQL root query field on the same `POST /graphql`
endpoint (`stigg.go`'s `streamEndpoints` map): `products`, `plans`, `customers`, `subscriptions`.
None of the 4 streams paginate — every query fetches the full result set in one request (legacy's
`Read`, `stigg.go:66-102`, issues exactly one `r.Do` call per stream, no loop, no cursor/page
argument in any query text). `hooks/stigg/hooks.go`'s `ReadStream` ports each stream's exact query
text and `data.<field>` record-extraction path verbatim:

- `products`: `query PolymetricsProducts { products { id refId displayName status } }` at
  `data.products`.
- `plans`: `query PolymetricsPlans { plans { id refId displayName status } }` at `data.plans`.
- `customers`: `query PolymetricsCustomers { customers { id refId displayName status } }` at
  `data.customers`.
- `subscriptions`: `query PolymetricsSubscriptions { subscriptions { id refId customerId status }
  }` at `data.subscriptions`.

Every record is projected to EXACTLY the fields legacy's `copyRecord` helper selects (`id`, `refId`,
`displayName`, `status` for products/plans/customers; `id`, `refId`, `customerId`, `status` for
subscriptions) — legacy discards every other field the raw GraphQL response might carry, matching
this bundle's schema-mode projection (no `computed_fields` renames needed since Stigg's GraphQL
field names are already the exact record keys legacy publishes, camelCase and all — unlike a REST
API this is not renamed to snake_case, since the schema is defined as literal field-for-field parity
with legacy's own `copyRecord` output, not a stylistic normalization).

Neither stream is incremental: legacy's `Read` never consults `req.State`/a cursor at all (grep
confirms no state read anywhere in `stigg.go`) — every read is a full re-fetch, matching this
bundle's schemas declaring no `x-cursor-field` and `streams.json` declaring no `incremental` block
on any stream.

**Deliberately NOT ported**: legacy's `Read` has no GraphQL-errors-in-HTTP-200 handling at all
(unlike monday's connector, which explicitly checks a top-level `errors` array even on a 200
response) — `stigg.go:85-92` decodes `resp.Body` directly at `endpoint.recordsPath` regardless of
whether the response also carries a GraphQL `errors` envelope. `hooks/stigg/hooks.go`'s `execute`
mirrors this exactly: it does NOT inspect the response for a GraphQL `errors` array, matching
legacy's behavior byte-for-byte (an error-shaped response with an empty/absent `data.<field>`
produces zero records, not a hard failure — see "Known limits").

### Declarative path (`streams.json`) vs. the live StreamHook path

`streams.json` still declares complete stream/schema metadata for all 4 streams — this is what
backs the catalog/manifest surface regardless of which path a read actually takes. Because
`hooks/stigg/hooks.go`'s `StreamHook.ReadStream` recognizes and handles all 4 stream names
unconditionally (`handled=true`), the declarative fallback in `streams.json` is **never exercised
by production traffic** — `engine.Read` only falls through to it when the `StreamHook` returns
`handled=false` (an unrecognized stream name) or no hooks are registered at all. Every stream
carries an explicit `"conformance": {"skip_dynamic": true, "reason": "..."}` marker
(conventions.md §4/§6): `internal/connectors/conformance/dynamic.go` honors this by Skipping every
dynamic fixture-replay check for these streams, since a declarative GET-shaped replay can never
faithfully exercise a GraphQL POST `StreamHook`. The authoritative substitute these markers name is
`internal/connectors/paritytest/stigg/parity_test.go` (drives the real, hook-dispatched connector
via `engine.HooksFor("stigg")`) and `hooks/stigg/hooks_test.go` — both assert stigg's real GraphQL
wire format (query text, `data.<field>` extraction, record mapping) byte-for-byte against legacy.
`fixtures/streams/<stream>/page_1.json` is retained purely as documentation of the real record
shape each stream emits (and to satisfy `fixtures_present`'s static "first stream ships a fixture"
requirement), not as a load-bearing replay contract.

## Write actions & risks

None. Stigg is a read-only source connector (`capabilities.write: false`, no `writes.json`),
matching legacy's `Write` returning `connectors.ErrUnsupportedOperation` unconditionally.

## Known limits

- **`StreamSpec.Body` is unwired (ENGINE_GAP, documented, non-blocking; same gap monday's and
  serpstat's bundles already ledger).** The engine's declarative read path never sends a request
  body, so a POST-body GraphQL read cannot be expressed in `streams.json` alone.
  `hooks/stigg/hooks.go`'s `StreamHook` implements the real GraphQL POST entirely within the
  sanctioned Tier-2 hook seam, reusing `rt.Requester` (the engine's already-built HTTP client/
  auth/base-URL plumbing, including the `Authorization: Bearer` header).
- **The declarative `streams.json` path is never live-dispatched** (see "Declarative path" above)
  — every stream carries a `conformance.skip_dynamic` marker naming
  `paritytest/stigg`/`hooks/stigg/hooks_test.go` as the authoritative substitute.
- **No incremental filtering, matching legacy exactly.** No stream declares
  `x-cursor-field`/`incremental` — legacy never filters or advances reads by any cursor; every read
  is a full, unpaginated re-fetch.
- **No GraphQL-errors-in-HTTP-200 detection, matching legacy exactly.** Unlike monday's connector,
  neither legacy's `stigg.go` nor this bundle's `StreamHook` inspects the response body for a
  top-level GraphQL `errors` array — a query that Stigg answers with `errors` (and an empty/absent
  `data.<field>`) is treated the same way on both sides: zero records emitted, no error surfaced.
  This is intentional parity with legacy's exact (arguably under-defensive) behavior, not an
  oversight; hardening this is Pass B scope if ever revisited, tracked here so it is not silently
  reintroduced as a "fix" that would diverge from legacy.
- **Legacy's `mode: fixture` credential-free affordance is NOT part of this bundle.** Legacy's
  `readFixture`/`fixtureMode` (`stigg.go:145-163`) emit synthetic records (stamping extra
  `connector`/`fixture` marker fields not present on any live record) without any network call when
  `config.mode == "fixture"` — a legacy-only testing convenience. Parity is asserted against
  legacy's LIVE (httptest-driven) read path only, matching the wave1-pilot convention (monday's
  docs.md carries the identical note). The `connector`/`fixture` marker fields are therefore
  correctly absent from this bundle's schemas.
- **`base_url` validation is stricter than a bare string default, matching legacy's own
  `baseURL()` helper** (`stigg.go:120-136`): must parse as a URL with an `http`/`https` scheme and a
  non-empty host; both legacy and `hooks/stigg/hooks.go` fail closed identically on a malformed
  override.
