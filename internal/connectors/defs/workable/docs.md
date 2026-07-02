# Overview

Workable is a wave2 fan-out declarative-HTTP migration. It reads Workable jobs, candidates, and
members through the Workable SPI v3 API (`GET {base_url}/jobs|candidates|members`). This bundle
migrates `internal/connectors/workable` (the hand-written connector it replaces); the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Workable API access token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`), matching legacy's `connsdk.Bearer(key)` (`workable.go:159`),
and is never logged. `base_url` must be the full SPI v3 base URL (e.g.
`https://example.workable.com/spi/v3`) — see Known limits for why legacy's
`account_subdomain`-only shorthand is not modeled.

## Streams notes

All three streams (`jobs`, `candidates`, `members`) hit their respective Workable list endpoints;
records live at the top-level key matching the stream name (`jobs`/`candidates`/`members`),
matching legacy's per-stream `recordsPath`. Pagination follows Workable's own `paging.next`
absolute-URL convention (`pagination.type: next_url`, `next_url_path: "paging.next"`), matching
legacy's own loop (`Read`'s `next, err := connsdk.StringAt(resp.Body, "paging.next")`). `limit=100`
(legacy's own `defaultPageSize`) is sent as a static per-stream query value, matching stripe's
`limit=100` static-query precedent.

`created_after` is sent whenever `start_date` is configured (`stream.Query`'s optional-query
dialect, `omit_when_absent: true`), matching legacy's blanket
`if start := ...; start != "" { q.Set("created_after", start) }` applied identically to every
stream's request (legacy's own `Read` builds this filter once and reuses it regardless of which
stream is being read — reproduced here per-stream since the declarative dialect has no
once-per-Read shared-query-builder concept, but the resulting request is identical either way).

## Write actions & risks

None. This connector is read-only in both legacy and this bundle (`capabilities.write: false`); no
`writes.json` is shipped.

## Known limits

- **`account_subdomain`-derived base URL shorthand is not modeled.** Legacy accepts either an
  explicit `base_url` override or a bare `account_subdomain`, deriving
  `https://<subdomain>.workable.com/spi/v3` in code (`baseURL`). The engine's `spec.json` `default`
  materialization only fills a fixed literal for a genuinely-absent key; it has no mechanism to
  derive one config value's default from another config value's runtime content. This bundle
  therefore requires `base_url` directly (the full SPI v3 URL) and does not declare
  `account_subdomain` at all (a declared-but-unwireable key is worse than an absent one, per
  `docs/migration/conventions.md`'s F6 precedent). This is a documented config-surface narrowing: a
  caller who previously configured only `account_subdomain` must now supply the equivalent full
  `base_url`; the emitted request and data are byte-identical once they do.
- **The declarative `Check` request cannot reproduce legacy's `limit=1` query param.** The engine's
  `RequestSpec` (used for `base.check`) has only `method`/`path` fields, no `query` — this is a
  structural constraint of the engine dialect shared by every migrated connector (see stripe's and
  bitly's own check blocks, neither of which declares a query either), not something specific to
  Workable. Legacy's check additionally scopes the request to 1 record (`url.Values{"limit":
  []string{"1"}}`); this bundle's check requests the same `jobs` endpoint without that limit
  param — both requests succeed identically against a real Workable API (the limit is a bandwidth
  optimization for the check call, not a behavior legacy depends on for correctness) and are
  functionally equivalent connectivity/auth probes.
- **`jobs`/`members` schemas omit `x-cursor-field` despite legacy declaring `created_at` as the
  catalog cursor field for every stream.** Legacy's `streams()` stamps `CursorFields:
  []string{"created_at"}` uniformly across all three streams regardless of whether that stream's
  own declared `Fields` list actually includes a `created_at` property — `jobs` (id/shortcode/
  title/state) and `members` (id/name/email) never declare one. The engine's `cursor_field_missing`
  validate rule requires `x-cursor-field` to name a property that exists in that same schema, so
  reproducing legacy's inconsistent declaration verbatim is not possible without either adding a
  field neither stream's real API response carries (a schema-shape deviation) or accepting the
  validate failure. This bundle declares `x-cursor-field: created_at` only on `candidates` (which
  genuinely has the field); `jobs`/`members` support only `full_refresh[_deduped]` sync modes as a
  result. Legacy itself performs no actual per-record cursor-based incremental filtering on any of
  these three streams (there is no `InitialState`/cursor-comparison logic anywhere in
  `workable.go` — only a blanket `created_after` query param gated on `start_date`), so no reader
  behavior is lost; this is a catalog-metadata-only correction of a real legacy inconsistency.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`pageSize`/`maxPages`, `workable.go:212-243`; `max_pages` defaults to `1` page,
  notably NOT unbounded like every other connector's convention). `PaginationSpec.MaxPages` and
  every pagination field in this dialect are static integers/strings fixed at bundle-load time —
  there is no templating mechanism on any `pagination` field (unlike `stream.Query`), so neither
  value can be wired to a runtime config key. `page_size` is therefore not declared in `spec.json`
  at all (F6, REVIEW.md: a declared-but-unwireable key is worse than an absent one) and its value
  (`limit=100`, legacy's own `defaultPageSize`) is sent as a static per-stream query literal,
  matching bitly's and stripe's identical precedent. `max_pages` is likewise not declared;
  pagination is bounded only by the `next_url` paginator's own short/empty-`paging.next` stop
  signal, matching Workable's real termination behavior once a caller's page count naturally runs
  out — this bundle does not reproduce legacy's unusual `max_pages=1`-by-default early stop, a
  behavior narrowing (more records may be read per sync than legacy's default would have fetched,
  never fewer, and never different data for any record legacy would have read).
