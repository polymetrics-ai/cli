# Overview

Mux is a video infrastructure API. This bundle reads Mux Video assets, live streams, direct
uploads, and system signing keys through the Mux REST API using HTTP Basic authentication. It is
read-only, migrating `internal/connectors/mux` (the hand-written legacy connector, which stays
registered and unchanged until wave6's registry flip) at capability parity.

## Auth setup

Provide the Mux API access token id as the `username` config value and the token secret as the
`password` secret; both flow into HTTP Basic auth (`Authorization: Basic base64(username:password)`)
and the secret is never logged, matching legacy's `connsdk.Basic(username, secret)` wiring.

## Streams notes

All 4 streams (`assets`, `live_streams`, `uploads`, `signing_keys`) list against
`{"data":[...]}`-enveloped Mux endpoints (`video/v1/assets`, `video/v1/live-streams`,
`video/v1/uploads`, `system/v1/signing-keys`), each mapped field-for-field from legacy's own
`mapRecord` functions (`internal/connectors/mux/streams.go`). Pagination is `page_number`
(`page`/`limit` query params, `page_size: 25` matching legacy's `muxDefaultPageSize`) with a
short-page stop, identical to legacy's `harvest` loop. None of the 4 streams declare a
server-side `incremental` filter — legacy's `Read` never sends a lower-bound query parameter for
any Mux stream, even though `assets`/`live_streams`/`signing_keys` publish a `created_at` field as
their cursor for catalog purposes (`x-cursor-field` is set on those schemas to preserve that
catalog metadata) — only full-refresh reads are modeled, matching legacy exactly.

`created_at` is emitted as a **string** (not an integer), matching Mux's real wire shape and
legacy's own `Field{Name: "created_at", Type: "string"}` declaration — Mux's API returns this
field as a numeric-looking string, not a bare JSON number, and legacy never parses/re-types it.

## Write actions & risks

None. Mux is a read-only source connector in this bundle (`capabilities.write: false`), matching
legacy's `Write` stub that always returns `connectors.ErrUnsupportedOperation`.

## Known limits

- Full Mux API surface (asset creation/deletion, live stream creation, DRM, playback restrictions,
  webhooks, Data API video views/errors) is out of scope for this wave; see `api_surface.json`'s
  `api_surface.json` concrete exclusion entries
  entries. Only the 4 legacy-parity read streams are implemented.
- **`page_size` is not runtime-configurable.** Legacy exposes `page_size` as a config-driven
  override (`muxPageSize`, `mux.go:280-293`, default 25, max 100). The engine's `page_number`
  paginator's `PageSize` is a static bundle-authored int (not templated), so there is no way to
  expose it as a config override; `page_size` is therefore not declared in `spec.json` at all (F6,
  REVIEW.md: a declared-but-unwireable config key is worse than an absent one), matching miro's
  identical documented precedent. `streams.json`'s `base.pagination.page_size` is fixed at `25`,
  legacy's own default, so the real per-page record count is unchanged for any caller who never
  overrode it; a caller who relied on legacy's config override to request a different page size has
  no equivalent here, a documented scope narrowing, not a data difference.
- **`max_pages` is not runtime-configurable.** Legacy exposes `max_pages` as a config-driven hard
  request-count cap override (`muxMaxPages`, `mux.go:295-308`, accepting an integer, `all`, or
  `unlimited`, default `0`/unbounded). The engine's `PaginationSpec.MaxPages` is a static
  bundle-authored int (not templated) — there is no config-driven knob to wire it to, so it is left
  unset (unbounded) in `streams.json`, matching legacy's own default behavior; a caller who relied
  on legacy's config override to bound page count has no equivalent here, a documented scope
  narrowing, not a data difference.
