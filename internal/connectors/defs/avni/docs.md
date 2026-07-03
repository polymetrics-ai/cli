# Overview

Avni is a read-only source connector. It reads Avni subjects and encounters through Avni's REST
API (`https://app.avniproject.org` by default, overridable for self-hosted instances) using HTTP
Basic authentication. This bundle migrates `internal/connectors/avni` (the hand-written legacy
connector); the legacy package stays registered and unchanged until wave6's registry flip. Legacy
is pure `connsdk`-HTTP with no signature auth, no custom stream handling, and no writes, so it maps
to a Tier-1 declarative bundle with zero Go.

## Auth setup

Provide `username` (plain config) and `password` (`x-secret`) for HTTP Basic authentication —
`Authorization: Basic base64(username:password)` — matching legacy's
`connsdk.Basic(cfg.Config["username"], secret(cfg, "password"))` exactly.

## Streams notes

Both streams (`subjects`, `encounters`) share the same shape: `GET` against an Avni list endpoint
returning the `{items:[...], next_page}` envelope (`records.path: "items"`), primary key `["id"]`.
Pagination is `cursor`/`token_path` (`cursor_param: page`, `token_path: next_page`): the response
body's `next_page` field carries the literal next page number as a string, and pagination stops
when `next_page` is empty — identical to legacy's `readPaged` loop (`page = next; if next == ""
{ return nil }`). Every request sends `page_size` (default `100`, matching legacy's
`defaultPageSize`) and an optional `start_date` query parameter, sent only when the `start_date`
config value is set (`omit_when_absent`) — legacy only calls `query.Set("start_date", start)` when
`strings.TrimSpace(cfg.Config["start_date"]) != ""`. `start_date` is a passthrough config value,
not a computed incremental lower bound (legacy never reads it back from a persisted sync cursor),
so no `incremental` block is declared here — `updated_at` stays a schema-only cursor candidate
(`x-cursor-field`), matching legacy's own `CursorFields: []string{"updated_at"}` catalog
declaration with no server-side incremental filter wired to it (§8 rule 2).

`check` issues a single bounded `GET /api/subjects?page_size=1`, mirroring legacy's `Check`
implementation exactly (a 1-record probe confirms auth and connectivity without mutating
anything).

## Write actions & risks

None. Avni is read-only (`capabilities.write: false`); legacy's own `Write` always returns
`connectors.ErrUnsupportedOperation`.

## Known limits

- Only the 2 legacy-parity read streams (`subjects`, `encounters`) are implemented; the full Avni
  API surface (programs, forms, media, etc.) is out of scope until Pass B.
- Documented parity deviation: legacy accepts a runtime `max_pages` config override
  (`intConfig(req.Config, "max_pages", defaultMaxPages)`); the engine's `pagination.max_pages` field
  is a static integer with no template support (no `PaginationSpec` field is ever resolved via
  `{{ }}` interpolation — see `conventions.md` §3's pagination table), so a per-request runtime
  override cannot be expressed declaratively. This bundle instead declares a fixed
  `pagination.max_pages: 100`, matching legacy's own `defaultMaxPages` constant exactly — the
  request-count cap every existing caller actually observes (nothing in the repo's config surface
  overrides `max_pages` from its default today). A caller that previously set a non-default
  `max_pages` would see a behavior change (this bundle always caps at 100); this is judged
  ACCEPTABLE as a documented scope narrowing rather than an `ENGINE_GAP`, since `max_pages` is a
  defensive request-count ceiling, not data-shaping logic, and every default-configured caller
  (the common case) is unaffected.
- `start_date` is sent verbatim as a static query parameter on every page of every read (matching
  legacy); it is not a computed, cursor-state-driven incremental lower bound.
