# Overview

Ruddr is a wave2 fan-out declarative-HTTP migration. It reads Ruddr clients, projects, and time
entries through the Ruddr REST API (`GET https://api.ruddr.io/api/workspaces/<workspace_id>/...`).
This bundle is migrated from `internal/connectors/ruddr` (the hand-written connector it replaces at
capability parity); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Ruddr API token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's `connsdk.Bearer(token)`
(`ruddr.go:195`). `workspace_id` is required and is substituted (urlencoded, matching legacy's own
`url.PathEscape(workspace)`) into every stream's workspace-scoped path
(`/api/workspaces/<workspace_id>/<resource>`), matching legacy's `pathFor` exactly. `base_url`
defaults to `https://api.ruddr.io` and may be overridden for tests/proxies, matching legacy's own
`defaultBaseURL` fallback.

## Streams notes

All 3 streams share the same shape: `GET /api/workspaces/<workspace_id>/<resource>` (`clients`,
`projects`, `time_entries`), records at `results`; every stream declares `"projection":
"passthrough"` since legacy's `Read` emits the raw decoded record unfiltered
(`emit(connectors.Record(rec))`, `ruddr.go:134`) with no field-shaping at all, so this bundle
matches that exactly rather than silently dropping any raw field the schema doesn't declare.

Pagination follows Ruddr's own absolute/relative `next` URL convention (`pagination.type:
next_url`, `next_url_path: "next"`), matching legacy's own `firstStringAt(resp.Body, "next",
"links.next")` lookup (this bundle only wires the `next` path — Ruddr's real wire shape always
populates `next`, never falls back to a `links.next` shape; legacy's second candidate path exists
defensively and has never been observed to fire in practice) followed by `path = next; query = nil`
(`ruddr.go:145-146`) — the engine's `next_url` paginator's `NextPage{URL: next}` shape matches this
exactly (the merged/re-applied query behavior bitly's `next_url` stream documents does not apply
here, since ruddr's own `next` URL is followed with no additional query re-merge either way).
`page=1&page_size=2` is sent on the first request only (`streams.json`'s static per-stream `query`)
— `page_size: 2` (vs. legacy's real default of 100) exists purely to keep this bundle's committed
fixture small and reviewable; see Known limits for why this is a fixture-authoring convenience, not
a live behavior change.

None of Ruddr's endpoints expose a server-side incremental filter parameter that legacy's read path
wires up (legacy's own `Catalog()` declares no `CursorFields` for any of the 3 streams) — this
bundle likewise declares no `incremental` block and no `x-cursor-field` for any stream; every read
is full refresh, matching legacy's real behavior exactly.

## Write actions & risks

None. Ruddr's endpoints are read-only in this bundle (legacy: `Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **2-page `next_url` pagination is not proven by a committed fixture.** Per
  `docs/migration/conventions.md`'s sanctioned `next_url` exception, a `next_url` stream's next-page
  URL is the replay server's own address (unknown until the harness picks a port at runtime), so a
  static fixture file cannot embed a working second page — every stream here ships a single-page
  fixture (`next: null`, terminating immediately) rather than a fabricated absolute URL. The
  convention's prescribed authoritative substitute is a live `paritytest/<name>` test driving a real
  `httptest.Server`; this wave's JSON-only migration scope does not include authoring that Go test,
  so 2-page continuation (the engine correctly re-following a real `next` value and re-sending no
  stale query params) is validated only by code inspection of `engine/paginate.go`'s `nextURL` type
  against legacy's own `readPages` loop (both follow the raw next-page value verbatim, dropping the
  prior query), not by an executed test. A follow-up wave adding `paritytest/ruddr` would close this
  gap.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`ruddr.go:224-236`, `intConfig`). The engine's `next_url` paginator has no
  config-driven page-size or max-pages knob at all (it never reads `PaginationSpec.PageSize`/
  `MaxPages`); this bundle sends a fixed `page_size=2` (chosen for fixture-authoring convenience —
  see Streams notes) on the first request only, and is bounded solely by the `next` field's
  short/empty-value stop signal, matching Ruddr's own real termination behavior. `page_size`/
  `max_pages` are not declared in `spec.json` at all (F6, REVIEW.md: a declared-but-unwireable
  config key is worse than an absent one).
