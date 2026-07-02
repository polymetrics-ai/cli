# Overview

PersistIQ is a wave2 fan-out declarative-HTTP migration. It reads PersistIQ leads, users,
campaigns, mailboxes, activities, and accounts through v1 REST list endpoints
(`GET https://api.persistiq.com/v1/...`). This bundle is engine-vs-legacy parity-tested against
`internal/connectors/persistiq` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a PersistIQ API key via the `api_key` secret; it is sent as the raw `X-API-KEY` header
(`auth.mode: api_key_header`, `header: X-API-KEY`, no prefix) and is never logged, matching
legacy's `connsdk.APIKeyHeader("X-API-KEY", key, "")` (`persistiq.go:127`). `base_url` defaults to
`https://api.persistiq.com` and may be overridden for tests/proxies.

## Streams notes

All 6 streams (`leads`, `users`, `campaigns`, `mailboxes`, `activities`, `accounts`) share the same
shape: `GET` against the PersistIQ v1 list endpoint, records at a top-level key matching the stream
name, primary key `["id"]`. Pagination is page-number based (`pagination.type: page_number`,
`page_param: page`, `size_param: per_page`, `start_page: 1`, `page_size: 100`), stopping on a short
page — identical to legacy's `connsdk.PageNumberPaginator{PageParam: "page", SizeParam:
"per_page", StartPage: 1, PageSize: size}`. None of legacy's 6 streams declare an incremental
cursor field (legacy `streams()` sets no `CursorFields` for any of them), so this bundle declares
no `incremental` block for any stream, matching legacy exactly (full-refresh reads only).

## Write actions & risks

None. Legacy `persistiq.Write` always returns `connectors.ErrUnsupportedOperation`;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- Full PersistIQ API surface (lead creation, campaign membership mutation) is out of scope; see
  `api_surface.json`'s `excluded: {category: out_of_scope}` entries — legacy itself never
  implemented these.
- Legacy's schema declares only 4 common fields (`id`, `name`, `email`, `status`) identically
  across all 6 streams (`commonFields()`, `persistiq.go:104-106`); this bundle's per-stream schemas
  mirror that exactly rather than guessing a richer per-resource shape from the raw API.
