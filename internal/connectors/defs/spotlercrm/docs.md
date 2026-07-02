# Overview

Spotler CRM is a wave2 fan-out migration from `internal/connectors/spotlercrm` (the legacy
hand-written connector this bundle replaces at capability parity). It reads Spotler CRM contacts,
accounts, opportunities, and tasks through the Spotler CRM API. Read-only; the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Spotler CRM API key via the `api_key` secret; it is sent as the `X-API-Key` header
value with no prefix (`mode: api_key_header`, empty `prefix`), matching legacy's
`connsdk.APIKeyHeader("X-API-Key", key, "")`.

## Streams notes

All 4 streams (`contacts`, `accounts`, `opportunities`, `tasks`) share the identical shape: `GET`,
records at `data`, `page_number` pagination (`page_param: page`, `size_param: limit`, `start_page:
1`, `page_size: 100`) — matches legacy's `connsdk.PageNumberPaginator{PageParam: "page",
SizeParam: "limit", StartPage: 1, PageSize: pageSize}` with `defaultPageSize = 100` (note: legacy
names its size query parameter `limit` even though the pagination style is page-number-based, not
offset/limit — this bundle reproduces that exact query-param name). Primary key is `id` for every
stream, matching legacy's `PrimaryKey: []string{"id"}`.

Legacy performs no incremental/state-cursor filtering during `Read` — no stream declares an
`incremental` block.

## Write actions & risks

None. Legacy `spotlercrm.go`'s `Write` returns `connectors.ErrUnsupportedOperation`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` are not exposed as config properties.** Legacy accepts
  `config.page_size` (bounded 1-500, default 100) and `config.max_pages` (default unbounded) at
  read time. The engine's `PaginationSpec.PageSize`/`MaxPages` fields are static values baked into
  `streams.json`'s pagination block, with no `{{ }}` templating support from `config.*` (matching
  the split-io/spotify-ads/searxng precedent, F6 REVIEW.md). This bundle hard-codes `page_size:
  100` (legacy's own default) and declares no `max_pages` (unbounded, matching legacy's own
  default). A caller that previously overrode either value per-run loses that capability; every
  default-config caller sees byte-identical behavior.
- Only the 4 legacy-parity streams are implemented; the wider Spotler CRM API (email campaigns,
  forms, custom fields, webhooks) is out of scope for this wave — see `api_surface.json`'s
  `excluded: {category: out_of_scope}` entries.
