# Overview

Smartwaiver is a wave2 fan-out declarative-HTTP migration. It reads Smartwaiver waivers,
checkins, templates, published keys, and user info through the Smartwaiver v4 API
(`GET https://api.smartwaiver.com/v4/...`). This bundle migrates
`internal/connectors/smartwaiver` (the hand-written connector); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Smartwaiver API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`), matching legacy's `connsdk.Bearer(key)`
(`smartwaiver.go:150`). `base_url` defaults to `https://api.smartwaiver.com` and may be overridden
for tests/proxies, matching legacy's own `validatedBaseURL` default.

## Streams notes

All five streams hit their own `GET` endpoint: `waivers` (`/v4/waivers`, records at
`waivers.waivers`), `checkins` (`/v4/checkins`, records at `checkins.checkins`), `templates`
(`/v4/templates`, records at `templates.templates`), `published_keys` (`/v4/keys/published`,
records at `published_keys.keys`), and `user_info` (`/v4/info`, a single JSON object with no
records-array wrapper — `records.path: "."` returns the whole body as one record), matching
legacy's `streamEndpoints` map's nested/flat `recordsPath` shapes exactly.

None of the streams paginate in legacy (a single `r.Do` call per read, no loop) —
`pagination.type: none` is declared, one request per read — despite every request sending
`limit`/`offset=0` query params, matching legacy's `queryParams` (`smartwaiver.go:153-163`).
`limit` defaults to 100 (legacy's `defaultPageSize`) and is configurable via `page_size`.
`fromDts`/`toDts` are optional passthrough date filters, omitted entirely when unset, matching
legacy's `copyConfig` (only set when non-empty).

`Check` hits `/v4/me` (an account-identity probe distinct from the `user_info` stream's `/v4/info`
endpoint), matching legacy's `Check` (`smartwaiver.go:47`) exactly.

## Write actions & risks

None. Smartwaiver's legacy connector is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`page_size`'s upper bound (100) is not enforced by this bundle.** Legacy validates
  `page_size` is an integer between 1 and 100 in Go code (`smartwaiver.go:169-179`) before sending
  it as `limit`. The engine's declarative query dialect has no numeric-range validation
  primitive; an out-of-range `page_size` is sent to the API as-is rather than rejected
  client-side. This is a scope narrowing (client-side validation removed, not a data-shape
  change): the API itself is the ultimate arbiter of an invalid page size on both sides.
- **`start_date_2` (a secondary `fromDts` override config key) is not modeled.** Legacy accepts
  BOTH `start_date` and `start_date_2` as `fromDts` sources, applying `copyConfig` for
  `start_date` first and then `start_date_2` second (`smartwaiver.go:159-160`) — since both calls
  target the same `url.Values` key via `Set` (not `Add`), `start_date_2`, when present,
  unconditionally overwrites whatever `start_date` set, a two-config-key coalesce-with-precedence
  rule. The engine's `stream.Query` dialect has no mechanism to express "one of two optional
  config keys, second takes priority when both are set" as a single query entry — only one
  `template`/`default`/`omit_when_absent` triple can target one param name. This bundle therefore
  models only the primary `start_date` key (the documented, non-suffixed name); `start_date_2` is
  not declared in `spec.json` at all (a declared-but-unwireable key is worse than an absent one).
  Out-of-scope, not silently wrong: any sync relying solely on `start_date` is unaffected.
- **Response field sets in `schemas/*.json` are conservative.** Smartwaiver's public API
  reference was consulted for real wire-shape field names (`waiverId`, `templateId`, `createdOn`,
  `checkinId`, `key`, `label`, `username`, etc.), but legacy's own `streams()` catalog only ever
  declared a generic 3-field shape (`waiverId`/`templateId`/`createdAt`) shared identically across
  all 5 streams regardless of each stream's real shape (a legacy catalog-authoring shortcut, not a
  real per-stream contract). This bundle's schemas instead project each stream's own real,
  distinct identity/timestamp fields — a strictly more accurate, non-regressive superset of what
  legacy's shared catalog described, verified against Smartwaiver's public API documentation.
