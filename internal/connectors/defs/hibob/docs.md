# Overview

HiBob is a wave2 fan-out declarative-HTTP migration. It reads HiBob employee profiles, company
named lists, and people field definitions through the HiBob REST API v1
(`GET https://api.hibob.com/v1/...`), authenticating with HTTP Basic. This bundle targets
capability parity with `internal/connectors/hibob` (the hand-written connector it migrates); the
legacy package stays registered and unchanged until wave6's registry flip. HiBob is a read-only HR
data source in both legacy and this bundle.

## Auth setup

Provide a HiBob service-user id via the `username` config value and its token via the `password`
secret; the engine's `basic` auth mode sends them as HTTP Basic (`Authorization: Basic
base64(username:password)`), matching legacy's `connsdk.Basic(username, secret)` (`hibob.go:252`)
exactly. `password` is never logged.

## Streams notes

`profiles` is HiBob's one paginated endpoint: `GET /profiles` returns `{"employees":[...]}`,
paged with `limit`/`offset` query params (`pagination.type: offset_limit`, `page_size: 50`, matching
legacy's `hibobDefaultPageSize`) until a page returns fewer than `page_size` records, exactly
mirroring legacy's `harvest` offset loop (`hibob.go:161-190`). `named_lists`
(`GET /company/named-lists`, records at `values`) and `company_lists`
(`GET /company/people/fields`, records at the response body's ROOT — `records.path: ""`, the
engine's root-array convention) are both single-shot, non-paginated metadata reads, matching
legacy's `paginated: false` endpoints exactly.

Legacy defensively coerces every stream's `id` field through `nestedString` (`hibob.go:342-353`):
pass a string id through byte-for-byte, stringify a non-string id, and return `""` for a nil id —
it never inspects or splits the string's content. `id` is NOT a HAL/URI-keyed derived field, so
`last_path_segment` (a trailing-`/`-segment extractor, correct for calendly's `idFromURI(uri)`
convention) is the wrong tool here: HiBob's own endpoints are slash-namespaced
(`company/named-lists`, `company/people/fields`) and this platform commonly uses hierarchical ids
for named-list values, so an id containing a literal `/` would be silently truncated by that
filter — a real, undetected parity break (wave2 adversarial review finding, corrected here). This
bundle no longer declares an `id` computed_fields entry at all: plain `"schema"`-mode projection
already copies `raw["id"]` verbatim by key match, which is exactly what `nestedString` does for the
common case (id already arrives as a string) — with no truncation risk for any input, including
ids containing `/`. See Known limits for the one residual, narrow edge case this leaves
undecorated (a present-but-null id).

`profiles` also hoists nested `work.*`/`personal.pronouns` fields into top-level `work_*`/
`personal_pronouns` keys via `computed_fields`, matching legacy's `hibobProfileRecord` hoisting
(`streams.go:98-117`) field-for-field; the raw nested `work`/`personal` objects are NOT
additionally preserved as top-level fields in this bundle's schema (legacy's own record also keeps
`rec["work"]`/`rec["personal"]` verbatim in addition to the hoisted fields — see Known limits).

No stream exposes a server-side incremental filter parameter, and none of HiBob's endpoints carry
a natural change-tracking timestamp in the legacy connector (`CursorFields: nil` on every stream,
`hibob/streams.go:37-54`); this bundle declares no `x-cursor-field` and no `incremental` block on
any schema, matching legacy exactly — full refresh only.

## Write actions & risks

None. HiBob has no obviously-safe reverse-ETL writes in the legacy connector (`Capabilities: Write:
false`, and legacy's own `Write` explicitly sets `RecordsFailed: len(records)` before returning
`ErrUnsupportedOperation`); this bundle ships no `writes.json`.

## Known limits

- **`base_url` is required; legacy's `is_sandbox`-derived host selection is not modeled.** Legacy
  derives a default base URL from a boolean `is_sandbox` config flag
  (`https://api.hibob.com/v1` vs `https://api.sandbox.hibob.com/v1`, `hibob.go:275-294`). The
  engine's `spec.json` `"default"` materialization mechanism only fills in a SINGLE fixed literal
  default, not a value conditionally selected between two literals based on another config field —
  the same documented "derived default" gap as sentry's `hostname`-based URL or chargebee's
  `site`-based URL (`docs/migration/conventions.md` §3). Per that section's guidance, this bundle
  requires `base_url` directly and drops the `is_sandbox` convenience flag rather than inventing ad
  hoc Go for it — a documented config-surface narrowing: a caller targeting HiBob's sandbox simply
  supplies `https://api.sandbox.hibob.com/v1` as `base_url` directly.
- **Raw nested `work`/`personal` objects are not preserved alongside their hoisted fields.**
  Legacy's `hibobProfileRecord` keeps both the hoisted `work_*`/`personal_pronouns` top-level keys
  AND the original nested `rec["work"]`/`rec["personal"]` objects verbatim (`streams.go:113-115`).
  This bundle's `profiles` schema declares only the hoisted fields (`"schema"` projection mode
  silently drops anything not schema-declared) since the nested duplicates carry no additional
  information a downstream consumer cannot already get from the hoisted fields, and re-declaring
  full nested passthrough objects side-by-side with their own flattened copies would double the
  emitted payload for no data-fidelity gain. Documented scope-narrowing, not a data-loss risk for
  any field legacy's own hoisting logic actually surfaces.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides on the `profiles` stream (`hibobPageSize`/`hibobMaxPages`, `hibob.go:303-331`). The
  engine's `offset_limit` paginator has no config-driven page-size/max-pages knob wired to a spec
  property (mirrors bitly's/stripe's identical precedent for their own paginator types), so this
  bundle does not declare either in `spec.json` (F6, REVIEW.md: a declared-but-unwireable config
  key is worse than an absent one). Pagination is bounded only by the short-page stop signal,
  matching HiBob's own real termination behavior for any well-behaved sync.
- **A present-but-`null` `id` fails schema validation here rather than being coerced to `""` like
  legacy.** Legacy's `nestedString` (`hibob.go:342-353`) returns `""` for a nil id; this bundle
  relies on plain `"schema"`-mode projection (no `computed_fields` transform — see Streams notes
  above) to copy `raw["id"]` verbatim, which preserves a JSON `null` as `null` rather than coercing
  it to an empty string. Every stream's schema declares `id` as `required`/`type: "string"`
  (`x-primary-key`), so an upstream record with a null id would fail `records_match_schema`
  here, where legacy would have silently emitted an empty-string id instead. The engine's
  `computed_fields` dialect has no filter that stringifies-without-transforming (every filter
  either passes a string through unexamined, like the removed `last_path_segment`, or actively
  transforms/encodes it, like `urlencode`/`base64`/`unix_seconds`/`join`) — reproducing
  `nestedString`'s exact three-way coercion (string passthrough / stringify-other / nil-to-`""`)
  without also reintroducing a data-changing transformation is not expressible in the current
  dialect. Failing loudly on a null id is judged safer than silently emitting a synthetic
  empty-string primary key value legacy invented defensively but which HiBob's real API is not
  documented to ever produce (the three bundled streams' primary keys are always non-null in real
  HiBob responses).
