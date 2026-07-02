# Overview

TikTok Marketing is a wave2 fan-out declarative-HTTP migration. It reads TikTok Business
advertisers, campaigns, ad groups, and ads through the TikTok Marketing (Business) API
(`GET https://business-api.tiktok.com/open_api/v1.3/...`). This bundle targets capability parity
with `internal/connectors/tiktok-marketing` (the hand-written connector it migrates); the legacy
package stays registered and unchanged until wave6's registry flip. TikTok's Business API exposes
campaign/ad management mutations, but legacy's own package doc notes "there is no obviously safe
reverse-ETL write surface for a generic sync," so this bundle is read-only, matching legacy exactly.

## Auth setup

Provide a TikTok Business API access token via the `access_token` secret; it is sent as the
`Access-Token` header (`api_key_header` auth mode, not Bearer — TikTok's own convention), matching
legacy's custom `accessTokenHeader = "Access-Token"` authenticator (`tiktok_marketing.go:33-34,
258-263`), and is never logged. `base_url` defaults to
`https://business-api.tiktok.com/open_api/v1.3` (legacy's `tiktokDefaultBaseURL`).

## Streams notes

All four streams share the identical TikTok Business envelope:
`GET /<resource>/` returns `{"code":0,"message":"OK","data":{"list":[...],
"page_info":{"page":N,"page_size":N,"total_number":N,"total_page":N}}}`; records live at
`data.list` for every stream (legacy's `listPath: "data.list"`, uniform across all four
`tiktokStreamEndpoints`). Pagination is `page_number` (`page`/`page_size`, `start_page: 1`, static
`page_size: 100` matching legacy's `tiktokDefaultPageSize`).

The optional `advertiser_id` config filter is wired per-stream to match legacy's endpoint-specific
convention exactly (`harvest`, `tiktok_marketing.go:148-158`): `campaigns`/`adgroups`/`ads` send it
verbatim as the `advertiser_id` query parameter (`omit_when_absent: true` — left off entirely when
unset); the `advertisers` stream (backed by `advertiser/info/`) instead sends it wrapped as a
JSON-array-string `advertiser_ids` query parameter (`["<id>"]`, built via a mixed literal+reference
`computed` query template), matching legacy's special-cased `advertiser_ids` branch for that one
endpoint.

Every TikTok object publishes `modify_time` as `x-cursor-field` (matching legacy's own
`CursorFields: []string{"modify_time"}` on campaigns/adgroups/ads; `advertisers` has none, matching
legacy's `CursorFields: nil`), but TikTok's list endpoints expose no server-side incremental filter
parameter and legacy's own `harvest` never applies one — every read is a full paginated sweep
regardless of any prior sync's cursor. This bundle therefore declares `incremental.cursor_field`
with no `request_param`/`start_config_key`/`client_filtered` on the three streams that have one,
matching legacy's true read behavior exactly (cursor field published for downstream sync-mode
eligibility, never computed/sent as a filter).

## Write actions & risks

None. TikTok Marketing is read-only (`capabilities.write` is `false`); this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Legacy's `total_page`-based pagination stop is approximated by short-page stop only.** Legacy's
  `harvest` stops when `totalPage <= page || len(records) == 0`
  (`tiktok_marketing.go:191-196`), reading TikTok's own `data.page_info.total_page`. The engine's
  `page_number` paginator instead stops on a short page (`recordCount < page_size`). These two stop
  conditions are equivalent for every dataset except one whose total record count is an exact
  multiple of 100, where legacy stops immediately via the `total_page` comparison and the engine
  would issue one additional request returning an empty page before stopping — no different
  records are ever emitted either way (the same non-data-affecting divergence documented on this
  wave's aha/thinkific-courses/ticketmaster/tmdb bundles).
- **TikTok's HTTP-200-with-body-level-error-code envelope is not inspected.** Legacy's
  `tiktokAPIError` reads the response body's top-level `code` field on every request (including
  paginated reads) and surfaces a non-zero code as an explicit error with TikTok's own `message`
  text (`tiktok_marketing.go:330-342`, called from both `Check` and `harvest`). The declarative
  engine has no equivalent: `error_map` is keyed on HTTP status codes only (TikTok API-level errors
  are always returned with HTTP 200), and a body-code error response has no `data.list` array, so
  `records.path: "data.list"` simply resolves to nothing and the paginator's short-page/empty-page
  stop condition ends the read silently rather than surfacing TikTok's error message. This is a
  genuine gap in error-path fidelity (a TikTok-side API error becomes a silent empty/truncated
  read instead of a hard, explicit failure) — but it never changes the DATA emitted for any
  successful (code 0) response, which is the only case the §5 meta-rule (this convention doc)
  requires parity for; inspecting a 200-body error code is exactly the "GraphQL-errors-in-HTTP-200
  envelope" shape this migration wave's conventions name as canonical Tier-2 hook territory
  (`RecordHook`/`StreamHook`), out of scope for this JSON-only wave. Flagged here rather than
  silently worked around; a follow-up hooks wave can close this by adding a `StreamHook` or
  `CheckHook` that inspects `code` before falling back to the declarative path.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size`
  (default 100, 1-1000) and `max_pages` (0/absent/"all"/"unlimited" = unbounded, or a positive
  integer cap) as config-driven overrides (`tiktokPageSize`/`tiktokMaxPages`,
  `tiktok_marketing.go:300-328`). The engine's `page_number` paginator has no config-driven
  page-size or request-count-cap knob (mirrors the aha/thinkific-courses/ticketmaster/tmdb
  precedent from this same wave); `page_size`/`max_pages` are therefore not declared in
  `spec.json`, and this bundle sends TikTok's own default (`page_size=100`) as a static
  pagination-block value with no page cap.
- **The dotted `credentials.access_token` secret-key alias is narrowed to a single `access_token`
  key.** Legacy accepts either `credentials.access_token` (the catalog's dotted secret field name)
  or a bare `access_token` key as a convenience fallback (`tiktokAccessToken`,
  `tiktok_marketing.go:269-277`). This bundle declares only `access_token`; an operator previously
  using the dotted key name must rename it. This is a config-surface naming narrowing only — the
  resolved effective token value once configured is identical.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path stamps
  `connector`/`fixture`/`previous_cursor` marker fields with no live-path equivalent
  (`tiktok_marketing.go:233-238`). This bundle's schemas and fixtures target the live record shape
  only; the engine's own `internal/connectors/conformance` fixture-replay harness provides the
  credential-free test affordance this bundle needs.
