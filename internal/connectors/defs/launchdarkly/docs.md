# Overview

LaunchDarkly is a wave2 fan-out declarative-HTTP migration. It reads LaunchDarkly projects, members,
audit log entries, feature flags, and environments through the LaunchDarkly REST API v2 (default
`https://app.launchdarkly.com/api/v2`). This bundle migrates `internal/connectors/launchdarkly` (the
hand-written connector) at capability parity; the legacy package stays registered and unchanged until
wave6's registry flip. LaunchDarkly is read-only here — legacy has no reverse-ETL write set — so
`capabilities.write` is `false` and no `writes.json` is shipped.

## Auth setup

Provide a LaunchDarkly access token via the `access_token` secret. It is sent VERBATIM (no `Bearer`
prefix) as the `Authorization` header (`auth: [{"mode": "api_key_header", "header": "Authorization",
"value": "{{ secrets.access_token }}"}]`, `prefix` intentionally omitted/empty), matching legacy's
`connsdk.APIKeyHeader("Authorization", secret, "")` (`launchdarkly.go:200`) exactly — never logged.
`project_key` is an optional config value required only by the `flags` and `environments` streams
(which are scoped to a single LaunchDarkly project); those two streams' `path` templates
`{{ config.project_key }}` and hard-error at read time if it is unset, matching legacy's
`resolveResource` (`launchdarkly.go:206-215`) exactly — `projects`/`members`/`auditlog` never reference
it and work with no `project_key` configured. `base_url` defaults to
`https://app.launchdarkly.com/api/v2` and may be overridden for tests/proxies.

## Streams notes

All 5 streams share LaunchDarkly's uniform envelope: records live under the top-level `items` array
(`records.path: "items"`), and pagination is `offset_limit` (`limit`/`offset` query params) — matching
legacy's `connsdk.OffsetPaginator{LimitParam: "limit", OffsetParam: "offset", PageSize: pageSize}`
(`launchdarkly.go:134-141`) exactly, with a short-page stop (fewer than `page_size` records on a page
stops pagination). `flags` and `environments` template `project_key` into their path
(`/flags/{{ config.project_key }}`, `/projects/{{ config.project_key }}/environments`), matching
legacy's `{project_key}` substitution.

None of the 5 streams declare an `incremental` block: legacy's `Read` never applies a server-side or
client-side cursor filter for any stream, including `auditlog` — `CursorFields: []string{"date"}` is
declared on the legacy `auditlog` `Stream` catalog entry purely for downstream state-cursor bookkeeping,
never used to filter a request (there is no `after`/`before`-style query param sent, ever). This
bundle's `auditlog` schema mirrors that exactly: `x-cursor-field: "date"` is declared (so
`incremental_append`-family sync modes stay selectable downstream, per design §B.6) with no
corresponding `streams.json` `incremental` block — the identical "declared for bookkeeping, not
enforced" shape as legacy, klaviyo, and high-level's own migrated bundles.

Every stream's record mapper is a flat field-for-field passthrough from the raw LaunchDarkly object
(no `attributes` nesting, unlike Klaviyo/Leadfeeder) — schema projection alone reproduces legacy's
mappers exactly; no `computed_fields` renames are needed.

## Write actions & risks

None. LaunchDarkly is `capabilities.write: false`; no `writes.json` is shipped, matching legacy's
`Write` always returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size`/`max_pages` are no longer runtime-configurable, and `page_size` is fixed small (2)
  rather than at legacy's default (20).** Legacy exposes both as config overrides
  (`launchdarklyPageSize`/`launchdarklyMaxPages`, `launchdarkly.go:245-273`, default 20, max 100). The
  engine's `offset_limit` paginator's `PageSize` is a static bundle-authored int (not templated), with
  no `MaxPages`-equivalent config-driven knob either; `max_pages` is left unbounded (matching legacy's
  own default-unlimited behavior). `page_size` is set to `2` (not legacy's default of 20) specifically
  so the mandatory 2-page conformance fixture (`fixtures/streams/projects/{page_1,page_2}.json`) is
  realistic to author and honestly exercises the short-page stop rule (`conformance`'s
  `pagination_terminates` check requires the replay server to serve exactly one request per fixture
  page — a `page_size` of 20 against a small hand-authored fixture would short-circuit after page 1 and
  never touch page 2 at all), matching cal-com's and bamboo-hr's identical documented precedent
  (`docs/migration/conventions.md`). This changes the real per-page record count from legacy's 20 to
  2 — a REST-shape difference (more, smaller requests), never a data-emission difference (every
  project/member/audit-entry/flag/environment is still read exactly once, across more pages).
- **`auditlog`'s `date` cursor field is catalog-only, matching legacy exactly (not a deviation).** See
  "Streams notes" above — this is legacy's own behavior, ported faithfully, not a scope narrowing
  introduced by this migration.
