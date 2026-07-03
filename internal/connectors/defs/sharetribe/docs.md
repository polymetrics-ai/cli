# Overview

Sharetribe is a wave2 fan-out declarative-HTTP migration. It reads Sharetribe listings, users,
transactions, and events through the Sharetribe Integration API
(`GET https://flex-api.sharetribe.com/v1/integration_api/<resource>/query`). This bundle is
capability-parity migrated from `internal/connectors/sharetribe` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Sharetribe Integration API OAuth2 access token via the `oauth_access_token` secret; it is
sent as a Bearer token (`Authorization: Bearer <oauth_access_token>`), matching legacy's
`connsdk.Bearer(token)` (`sharetribe.go:102`) exactly, and is never logged. `base_url` defaults to
`https://flex-api.sharetribe.com/v1` (legacy's `defaultBaseURL`) and may be overridden for
tests/proxies.

## Streams notes

All 4 streams (`listings`, `users`, `transactions`, `events`) share the same shape: `GET
/integration_api/<resource>/query`, records at the response body's `data` array, and `page_number`
pagination (`page`/`per_page` query params, 1-based start page) — an exact port of legacy's
`connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "per_page", StartPage: 1, PageSize:
pageSize}` (`sharetribe.go:87`). A page returning fewer records than `page_size` signals the last
page. Every stream shares an identical field set (`id`, `type`, `attributes`, `updated_at`),
matching legacy's own single shared `fields` slice (`sharetribe.go:115`).

No stream is incremental; legacy's `streams()` declares no `CursorFields` for any of the 4 streams
and never filters by time (matching this bundle's omission of any `incremental` block).

## Write actions & risks

None. Sharetribe is read-only (`capabilities.write: false`, no `writes.json`), matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable per the engine dialect.** Legacy exposes
  both as config-driven overrides (`sharetribe.go:79-86`, `positiveInt`/`parseMaxPages`, `page_size`
  clamped 1-1000, `max_pages` defaulting to 1). The `page_number` paginator's `page_size` is a fixed
  value baked into `streams.json`'s `base.pagination` block, and there is no per-request `max_pages`
  override mechanism at all (conventions.md §3). Neither key is declared in `spec.json` (a
  declared-but-unwireable key is worse than an absent one — searxng precedent).
- `page_size` is baked at `100`, matching legacy's own default (`sharetribe.go:19`,
  `defaultPageSize`); fixtures use a full 100-record page 1 for `listings` (to exercise the
  page_number continuation into page 2) and single, sub-`page_size` pages for the other three
  streams (to exercise the short-page termination signal).
- Legacy's own default `max_pages` is `1` (`sharetribe.go:19`, `defaultMaxPages`); this bundle's
  unset `MaxPages` (unbounded, stopped only by the short-page signal) is a strict superset of
  legacy's default single-page behavior — no caller-visible behavior regresses for the common case
  (fewer records than one page).
