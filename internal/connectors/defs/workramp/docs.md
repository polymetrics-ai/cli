# Overview

WorkRamp is a wave2 fan-out declarative-HTTP migration. It reads users, groups, and courses
through the WorkRamp API (`GET {{ config.base_url }}/v1/...`). This bundle is migrated from
`internal/connectors/workramp` (the hand-written connector it replaces); the legacy package
stays registered and unchanged until wave6's registry flip. Read-only (`capabilities.write` is
`false`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`).

## Auth setup

Provide a WorkRamp API key via the `api_key` secret; it is sent as a Bearer token on every
request (`mode: bearer`), matching legacy's `connsdk.Bearer(token)`. `base_url` defaults to
`https://api.workramp.com` (legacy's `defaultBaseURL`) and may be overridden for test proxies.

## Streams notes

All 3 streams (`users`, `groups`, `courses`) share the identical envelope (records at the
top-level `data` array) and `page_number` pagination (`page`/`limit` query params, matching
legacy's `PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage: 1}` — note WorkRamp
uses `limit` as its page-size parameter name, not `page_size`). `page_size` defaults to 100
(legacy's `defaultPageSize`); legacy bounds it to a max of 500 (`maxPageSize`) and `max_pages`
defaults to 1 (legacy's `readMaxPages` default) when unset.

`users` (`GET /v1/users`) emits `id`/`email`/`updated_at`, matching legacy's field set exactly.
`groups` (`GET /v1/groups`) emits `id`/`name`/`updated_at`. `courses` (`GET /v1/courses`) emits
`id`/`title`/`updated_at`. Primary key is `id` for every stream; `updated_at` is declared as the
incremental cursor field for manifest-surface parity, matching legacy's `cursorFields`, though
neither legacy nor this bundle actually issues a server-side incremental filter — legacy's `Read`
performs a full stream read every time regardless of any prior cursor.

## Write actions & risks

None. Legacy `workramp.go`'s `Write` returns `connectors.ErrUnsupportedOperation`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` config-driven overrides are not modeled.** Legacy reads
  `config["page_size"]` (bounded 1-500) and `config["max_pages"]` (default 1) at request time via
  `boundedInt`/`readMaxPages`. The engine's `page_number` paginator reads `PaginationSpec.PageSize`
  from the static `streams.json` `base.pagination` block only — there is no per-request
  config-driven override mechanism for either value in the current dialect. `page_size`/`max_pages`
  remain declared in `spec.json` as documentation of legacy's accepted config surface, but neither
  is wired into any template in this bundle.
- **No incremental filter is modeled**, matching legacy: `updated_at` is declared as
  `x-cursor-field` for manifest parity, but WorkRamp's `/v1/users`, `/v1/groups`, and
  `/v1/courses` endpoints (as legacy calls them) accept no time-range query parameter — both
  connectors always perform a full stream read on every sync.
- The full WorkRamp API surface (user/group/course mutation, enrollments, assessments) is out of
  scope for this wave; see `api_surface.json`'s `excluded` entries.
