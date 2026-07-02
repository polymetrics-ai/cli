# Overview

Segment is a wave2 fan-out declarative-HTTP migration. It reads Segment workspace, source, and
destination metadata through the Segment Public API (`GET https://api.segmentapis.com/...`). This
bundle is engine-vs-legacy parity-tested against `internal/connectors/segment` (the hand-written
connector it migrates); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide a Segment Public API access token via the `api_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_token>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)` (`segment.go:106`). `base_url` defaults to `https://api.segmentapis.com`
and may be overridden for tests, proxies, or a region-specific endpoint.

## Streams notes

`workspaces`, `sources`, and `destinations` are simple list endpoints (`GET /workspaces`,
`/sources`, `/destinations`); records live at the top-level key matching the stream name, exactly
matching legacy's `streamEndpoints` records-path map (`segment.go:111-115`). None of the three
streams expose an incremental cursor field that legacy actually filters on — legacy declares no
`CursorFields`/incremental request param anywhere; this bundle likewise declares no `incremental`
block for any stream, matching legacy exactly (full refresh only). Pagination is `page_number`
(`page`/`page_size` query params, `start_page: 1`, `page_size: 100`), matching legacy's
`connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "page_size", StartPage: 1, PageSize:
pageSize}` (`segment.go:87`) with legacy's own default page size of 100 (`defaultPageSize`,
`segment.go:18`).

## Write actions & risks

None. Legacy's package declares `Capabilities.Write: false` and its `Write` method always returns
`connectors.ErrUnsupportedOperation`; `capabilities.write` is `false` here and this bundle ships no
`writes.json`.

## Known limits

- **Region-based base-URL derivation is not modeled.** Legacy accepts an optional `region` config
  value that, when `base_url` is unset, derives `https://<region>.segmentapis.com`
  (`segment.go:102-105`). This is a DERIVED default (a function of another config value, not a
  fixed literal) — conventions.md §3's `spec.json` `"default"` materialization mechanism only
  fills in a fixed literal for a genuinely-absent key, and there is no computed-base-URL template
  primitive in this dialect (the same class of gap documented for sentry/chargebee). This bundle
  therefore narrows the config surface to `base_url` alone: an operator on a non-default Segment
  region must set `base_url` directly (e.g. `https://api.eu1.segmentapis.com`) rather than a bare
  `region` value. `region` is not declared in `spec.json` at all (a declared-but-unwireable key is
  worse than an absent one, F6, REVIEW.md). This is a documented, accepted config-surface
  narrowing, not a silent behavior change — every request `base_url` legacy would actually reach
  (default global endpoint, or an explicit full `base_url` override) is reproduced identically;
  only the `region`-shorthand convenience is dropped.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) stamps a `previous_cursor` field (echoing
  `req.State["cursor"]` when a prior cursor happens to be set) onto fixture-mode records
  (`segment.go:126-140`). This is not part of the LIVE record shape; this bundle's schemas and
  fixtures target the live path only. The engine's own conformance/fixture-replay harness provides
  the credential-free test affordance this bundle needs.
- **`max_pages` is not runtime-configurable.** Legacy exposes a `max_pages` config override
  (`segment.go:83`, defaulting to 1) as a hard request-count cap. The engine's `page_number`
  paginator has no equivalent `MaxPages` wired from a `spec.json` config key today (`MaxPages` is
  set on `PaginationSpec` at bundle-author time, not resolved from a config template); this bundle
  declares no runtime cap, relying instead on the paginator's short-page stop signal (a page
  returning fewer than `page_size` records stops pagination) for termination, which is the same
  practical stop condition legacy's own default (`max_pages: 1`) plus short-page detection would
  reach for any fixture/small dataset exercised by this bundle's tests.
