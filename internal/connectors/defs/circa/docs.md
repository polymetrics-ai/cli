# Overview

Circa is a wave2 fan-out declarative-HTTP migration. It reads Circa events, contacts, companies,
and teams through the Circa REST API (`GET https://app.circa.co/api/v1/...`). This bundle targets
capability parity with `internal/connectors/circa` (the hand-written connector it migrates); the
legacy package stays registered and unchanged until wave6's registry flip. The connector is
read-only.

## Auth setup

Provide a Circa API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged. `base_url` defaults to
`https://app.circa.co/api/v1` and may be overridden for tests/proxies.

## Streams notes

All four streams read Circa's `{data:[...]}` list envelope. Pagination is `page_number`
(`page_param: page`, `size_param: ""` — Circa's API never accepts/needs a page-size query
parameter; legacy's own `harvest` only ever sends `page`, matching the engine's `page_number`
paginator with an empty `size_param`, `start_page: 1`, static `page_size: 25` matching legacy's
`circaDefaultPageSize`). The engine's short-page stop (`recordCount < page_size`) is identical to
legacy's own stop condition (`len(records) < pageSize`).

`events`/`contacts`/`companies` are incremental (`x-cursor-field: updated_at`), sent as
`updated_at[min]` (`incremental.request_param`) computed from the sync's persisted cursor or, on a
fresh sync, from the RFC3339 `start_date` config value (`incremental.start_config_key`) — identical
to legacy's `incrementalLowerBound`. `teams` is a full-refresh-only stream (no `incremental` block),
matching legacy's `circaStreamEndpoints["teams"].incremental == false`.

## Write actions & risks

None. Circa is read-only in this bundle (`capabilities.write: false`, no `writes.json`), matching
legacy's `Write` unconditionally returning `connectors.ErrUnsupportedOperation`.

## Known limits

- `page_size` (legacy's config-driven page-size override, 1-100, default 25) and `max_pages`
  (legacy's 0/all/unlimited-or-positive-integer request-count cap) are not runtime-configurable
  here: the engine's `page_number` paginator's `PageSize` is a static value set once in
  `streams.json` (25, matching legacy's `circaDefaultPageSize`), and `PaginationSpec` has no
  `MaxPages` field this paginator type reads. `spec.json` intentionally omits both `page_size` and
  `max_pages` — a declared-but-unwireable key is worse than an absent one (conventions.md F6),
  matching the aha/appfigures/cin7 wave2 precedent for the same paginator-type limitation.
- Legacy's `base_url` SSRF-guard scheme/host validation (https/http only, host required) is
  reproduced by the engine's own base-URL handling; no bundle-level behavior change.
