# Overview

Workday is a wave2 fan-out declarative-HTTP migration. It reads Workday tenant data — workers,
organizations, and positions — through a conservative subset of the Workday tenant API
(`GET {base_url}/ccx/api/v1/{tenant}/...`). This bundle is engine-vs-legacy parity-tested against
`internal/connectors/workday` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide Workday tenant credentials via the `username` and `password` secrets; both are required
(legacy hard-errors when either is empty: `"workday connector requires secrets username and
password"`, `workday.go:170-172`) and are sent via HTTP Basic auth
(`connsdk.Basic(username, password)`), matching this bundle's unconditional `auth: [{"mode":
"basic", ...}]` (no `when` gate — both are `required` in `spec.json`, so the engine's own
config-validation rejects a request before auth would ever be attempted with a missing secret).
Never logged.

`tenant` (config, required) is substituted as a path segment into every stream's resource URL
(`ccx/api/v1/{{ config.tenant }}/workers`, matching legacy's `resolveResource`'s `{tenant}`
placeholder substitution). The engine's `InterpolatePath` urlencodes the resolved value by default
and rejects path-traversal segments, matching legacy's own tenant validation (`tenant == "" ||
strings.ContainsAny(tenant, "/?#")` is rejected) with an equivalent, not byte-identical, error
classification (conventions.md §5 precedent: config-validation parity is bucketed by reason, not
exact string match).

`base_url` defaults to `https://wd2-impl-services1.workday.com` (legacy's own
`defaultBaseURL`) and may be overridden per-tenant or for tests/proxies.

## Streams notes

All 3 streams (`workers`, `organizations`, `positions`) are `GET` list endpoints returning records
under a top-level `data` key (`records.path: "data"`, matching legacy's `recordsPath: "data"`).
Pagination is `page_number` (`page_param: page`, `size_param: limit`, matching legacy's
`connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage: 1}`), stopping on a
short page; `page_size` defaults to 100 (legacy's `defaultPageSize`) and `max_pages` defaults to 1
(legacy's own default when `max_pages` is unset).

Every stream declares `x-cursor-field: updated_at` for catalog/sync-mode-classification parity with
legacy's `CursorFields: []string{"updated_at"}` (`workday.go:57,64,71`) — but, matching legacy
exactly, **no `incremental` block is declared and no request actually filters or advances by this
field**: legacy's `Read` never references `req.State` or sends any date-range query parameter at
all (verified: no `State` reference anywhere in `workday.go`). Every read is a full-refresh
page-through of the entire collection, identical on both sides.

All 3 streams declare `projection: "passthrough"` (§8 rule 1): legacy's `Read` calls
`connsdk.Harvest(...)` with an emit callback of `func(rec connsdk.Record) error { return
emit(connectors.Record(rec)) }` (`workday.go:141-143`) — `Harvest` extracts each page's records via
`RecordsAt` and hands them straight to that callback with no field-building anywhere in between.
Schema-mode projection would silently drop any real Workday tenant-API field beyond the handful
this bundle's schemas document (`id`/`name`/`updated_at` for `workers`, etc.); `passthrough`
reproduces legacy's verbatim emission exactly. This bundle's fixtures only exercise the
already-documented fields, so no schema widening was needed for fixture parity.

## Write actions & risks

None. Legacy `Write` always returns `connectors.ErrUnsupportedOperation`; `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`boundedInt`/`readMaxPages`, `workday.go:213-241`, bounded 1-500 for `page_size`). The
  engine's `page_number` paginator's `PageSize`/`MaxPages` fields are plain integers with no
  template/config-driven override mechanism, so neither can be wired to a `spec.json` config value;
  both are fixed in `streams.json`'s `base.pagination` block instead (`page_size: 100`, `max_pages:
  1`, both matching legacy's own defaults). Neither key is declared in `spec.json` (F6, REVIEW.md).
- **Legacy's fixture-mode-only fields (`connector`, `stream`, `fixture`) are not modeled.**
  Legacy's `readFixture` path (only reached when `config.mode == "fixture"`) stamps these 3 extra
  marker fields onto every fixture-mode record (`workday.go:151`). This bundle's schemas and
  fixtures target the live path only; the engine's own conformance/fixture-replay harness provides
  the credential-free test affordance this bundle needs.
- **Single-page fixtures only, matching `max_pages: 1`'s real, always-enforced behavior.** Since
  `page_size`/`max_pages` cannot be config-driven (previous bullet), `max_pages: 1` is a hard,
  unconfigurable cap in this bundle exactly as it is in legacy's own unset-`max_pages` default —
  every stream genuinely only ever fetches one page in practice, so a second fixture page would
  describe a request the connector never issues. This bundle ships single-page fixtures for every
  stream (the identical precedent set by `defs/searxng`'s own `max_pages: 1` streams), rather than
  a misleading 2-page fixture that `pagination_terminates` could never actually reach.
