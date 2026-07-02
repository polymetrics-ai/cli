# Overview

Rocketlane is a wave2 fan-out declarative-HTTP migration. It reads Rocketlane projects, tasks,
customers, users, and time entries through the Rocketlane REST API
(`GET https://api.rocketlane.com/api/1.0/...`). This bundle is migrated from
`internal/connectors/rocketlane` (the hand-written connector it replaces at capability parity); the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Rocketlane API key via the `api_key` secret; it is sent as the `api-key` header
(`api_key_header`, no prefix) and is never logged, matching legacy's
`connsdk.APIKeyHeader("api-key", key, "")` (`rocketlane.go:172`). `base_url` defaults to
`https://api.rocketlane.com/api/1.0` and may be overridden for tests/proxies, matching legacy's own
`defaultBaseURL` fallback.

## Streams notes

All 5 streams share the same shape: `GET` against a Rocketlane list endpoint, records at `data`
(`projects`, `tasks`, `customers`, `users`, `time-entries` — the last with a hyphen in its path,
matching legacy's `endpoints["time_entries"].path = "time-entries"`); every stream declares
`"projection": "passthrough"` since legacy's `mapRecord` (`rocketlane.go:191-201`) copies every raw
API field into the emitted record unfiltered before adding `id`/`stream`, so this bundle matches
that rather than silently dropping any raw field the schema doesn't declare. Pagination is 1-based
page-number (`pagination.type: page_number`, `page_param: page`, `size_param: pageSize`,
`start_page: 1`), matching legacy's `connsdk.PageNumberPaginator{PageParam: "page", SizeParam:
"pageSize", StartPage: 1}` short-page-stop semantics exactly. `streams.json`'s `pagination.page_size:
2` (vs. legacy's real default of 100) exists purely to keep this bundle's committed 2-page
conformance fixture (`projects`) small and reviewable (jira's identical precedent,
`docs/migration/conventions.md`) — `page_size` is a static bundle-authored JSON int with no
config-driven override on either side (see Known limits), so this is a fixture-authoring
convenience only, not a live-vs-fixture behavior divergence.

Legacy's four optional config passthrough filters (`updated_after`, `created_after`, `projectId`,
`status`, `rocketlane.go:88-91`) are wired per-stream to the endpoints legacy sent them on:
`updated_after`/`created_after` on every stream, `status` on `projects`/`tasks`/`users`, and
`project_id` (sent as `projectId`) on `tasks`/`time_entries` — each `omit_when_absent` so an
unconfigured filter is left off entirely rather than sent empty, matching legacy's own
`strings.TrimSpace(...) != ""` guard.

None of Rocketlane's list endpoints expose a server-side incremental filter parameter that legacy's
read path actually wires up — legacy's own streams declare `CursorFields: []string{"updated_at"}`
for catalog-metadata purposes only, with no `incremental` read-path implementation at all (the
`updated_after`/`created_after` passthrough filters are caller-supplied one-shot query params, not
an engine-driven persisted-cursor incremental loop). This bundle matches that exactly: every
stream's schema declares `x-cursor-field: updated_at` (metadata parity with legacy's catalog
declaration) but no stream declares an `incremental` block — every managed read is full refresh,
with `updated_after`/`created_after` available as manual one-shot filters via `spec.json` config.

## Write actions & risks

None. Rocketlane's endpoints are read-only in this bundle (legacy: `Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size` (1-200,
  default 100) and `max_pages` (0/all/unlimited = unbounded) as config-driven overrides
  (`rocketlane.go:249-255`). The engine's `page_number` paginator's `PaginationSpec.PageSize` is a
  static JSON number fixed at bundle-author time, not a `config.*`-templated value, and there is no
  `MaxPages`-equivalent config knob wired to a per-stream override either. This bundle sends a fixed
  page size (`pageSize=2`, chosen for fixture-authoring convenience — see Streams notes; unbounded
  pages) rather than legacy's real default of 100; `page_size`/`max_pages` are not declared in
  `spec.json` at all (F6, REVIEW.md: a declared-but-unwireable config key is worse than an absent
  one).
