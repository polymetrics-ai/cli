# Overview

noCRM is a wave2 fan-out declarative-HTTP migration of `internal/connectors/nocrm` (the
hand-written legacy connector this bundle migrates; the legacy package stays registered and
unchanged until wave6's registry flip). It reads noCRM.io leads, pipelines, users, teams, and
prospecting lists through the noCRM REST API v2 (`https://api.nocrm.io/api/v2`). Read-only.

## Auth setup

Provide a noCRM API key via the `api_key` secret; it is sent as the `X-API-KEY` header
(`connsdk.APIKeyHeader("X-API-KEY", secret, "")` in legacy, `mode: api_key_header` here) and is
never logged. `base_url` defaults to `https://api.nocrm.io/api/v2` and may be overridden for a
per-account subdomain deployment or for tests/proxies.

**Documented config-surface deviation (subdomain)**: legacy additionally accepts a bare
`subdomain` config value and derives `https://<subdomain>.nocrm.io/api/v2` itself
(`nocrm.go`'s `nocrmSubdomain`/`nocrmBaseURL`), falling back to a `base_url` override only when
set. The engine's declarative `url` template is a single static string with no conditional
branching between two config keys or string-derivation support (conventions.md §3's
`materializeConfigDefaults` mechanism fills in a fixed literal default only; it cannot derive a
URL from another config value), so this bundle requires `base_url` directly instead of declaring
an unwireable `subdomain` property (a declared-but-unwireable key is worse than an absent one —
see searxng's `subreddit` precedent). Every legacy-accepted `subdomain`-only configuration is
still reachable by an operator supplying the equivalent `https://<subdomain>.nocrm.io/api/v2` as
`base_url` instead — no request shape or emitted record ever differs between the two config
styles for the same effective account.

## Streams notes

All 5 streams (`leads`, `pipelines`, `users`, `teams`, `prospecting_lists`) share the same shape:
`GET` against the noCRM list endpoint, records extracted from the response's top-level JSON array
(`records.path: ""`), primary key `["id"]`. No stream declares an `incremental` block — legacy
never implements `InitialState`/exposes a cursor field for any of these 5 objects
(`nocrm/streams.go`'s doc comment: "the API supports only full_refresh syncs so no incremental
cursor field is published"), so every sync (fresh or resumed) reads the full, unfiltered
collection, matching legacy exactly.

Pagination is `offset_limit` (`offset`/`limit` query params, `page_size: 100` matching legacy's
`nocrmDefaultPageSize`/`nocrmMaxPageSize`, both fixed at 100). The engine's `OffsetPaginator`
stops on a short page (fewer than `page_size` records returned), the same primary stop signal
legacy's own `harvest` loop uses first. Legacy additionally consults the `X-TOTAL-COUNT` response
header as a secondary stop signal (`parseTotalCount`) to end pagination one page early when the
running record count already reaches the server-advertised total; the engine's `offset_limit`
paginator has no header-driven stop signal at all, relying solely on the short-page rule. This
is a **behavior-preserving optimization difference, not a data difference**: the short-page rule
alone is sufficient to terminate correctly and emit the identical record set — the header check
in legacy only ever saves at most one extra (empty) trailing request in the case where the final
page happens to be exactly `page_size` records long and the total is already known; the engine
will issue that one extra request (which returns zero records) before stopping. No record is
duplicated, dropped, or reordered either way.

## Write actions & risks

None. Legacy `nocrm` is read-only (`Write` returns `connectors.ErrUnsupportedOperation`);
`metadata.json` declares `capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- Full noCRM API surface (activities, calls, custom fields, webhooks, tags) is out of scope for
  wave2; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability
  expansion"}` entries. Only the 5 legacy-parity streams are implemented.
- **`X-TOTAL-COUNT` early-stop header is not modeled** (see "Streams notes" above): the engine's
  `offset_limit` paginator has no body/header-driven stop signal beyond the short-page rule. In
  the worst case (a final page exactly `page_size` records long) this bundle issues one extra,
  empty-body trailing request that legacy's header check would have skipped. No emitted record
  differs; this affects only page-count efficiency, never data.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`nocrmPageSize`/`nocrmMaxPages`, capped at 100 for `page_size`). The engine's
  `offset_limit` paginator reads `PaginationSpec.PageSize` as a fixed value declared in
  `streams.json`, not a templated reference to `config.*` — there is no mechanism to wire a
  runtime config value into a pagination spec field. `page_size` is therefore fixed at 100
  (legacy's own default and cap, so every legacy-valid request this bundle would ever need to
  reproduce is already covered) and `max_pages` is unbounded (legacy's own default of 0/unlimited),
  matching legacy's default behavior; neither key is declared in `spec.json` (a
  declared-but-unwireable config key is worse than an absent one, per conventions.md §3).
