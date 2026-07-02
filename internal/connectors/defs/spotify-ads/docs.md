# Overview

Spotify Ads is a wave2 fan-out migration from `internal/connectors/spotify-ads` (the legacy
hand-written connector this bundle replaces at capability parity). It reads Spotify Ads ad
accounts, campaigns, ad sets, and ads through the Spotify Ads API. Read-only; the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Spotify Ads OAuth access token via the `access_token` secret; it is used only for
Bearer auth (`Authorization: Bearer <access_token>`) and is never logged, matching legacy's
`connsdk.Bearer(token)`.

## Streams notes

`ad_accounts` needs no path parameter. `campaigns`, `ad_sets`, and `ads` each require
`config.ad_account_id` to substitute the `{ad_account_id}` path segment (`path:
"/ad_accounts/{{ config.ad_account_id }}/campaigns"`, etc.) — legacy's `resolveResource`
hard-errors with `"stream requires config ad_account_id"` when unset; an absent
`config.ad_account_id` surfaces as a hard path-interpolation error on those three streams' reads,
reproducing the same requirement through the engine's own error path.

All 4 streams share the identical pagination shape: `offset_limit` (`limit_param: limit`,
`offset_param: offset`, `page_size: 100`) — matches legacy's `connsdk.OffsetPaginator{LimitParam:
"limit", OffsetParam: "offset", PageSize: pageSize}` with `defaultPageSize = 100`. Each stream's
`records.path` matches its own resource name (`ad_accounts`, `campaigns`, `ad_sets`, `ads`),
identical to legacy's per-endpoint `recordsPath`.

Legacy performs no incremental/state-cursor filtering during `Read` — no stream declares an
`incremental` block.

## Write actions & risks

None. Legacy `spotify_ads.go`'s `Write` returns `connectors.ErrUnsupportedOperation`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` are not exposed as config properties.** Legacy accepts
  `config.page_size` (bounded 1-500, default 100) and `config.max_pages` (default unbounded) at
  read time. The engine's `PaginationSpec.PageSize`/`MaxPages` fields are static values baked into
  `streams.json`'s pagination block, with no `{{ }}` templating support from `config.*` (matching
  the split-io/searxng precedent, F6 REVIEW.md). This bundle hard-codes `page_size: 100` (legacy's
  own default) and declares no `max_pages` (unbounded, matching legacy's own default). A caller
  that previously overrode either value per-run loses that capability; every default-config
  caller sees byte-identical behavior.
- Only the 4 legacy-parity streams are implemented; the wider Spotify Ads API (creatives,
  targeting, insights/reporting, budget management) is out of scope for this wave — see
  `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
