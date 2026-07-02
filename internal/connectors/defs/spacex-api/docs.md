# Overview

SpaceX API is a wave2 fan-out migration from `internal/connectors/spacex-api` (the legacy
hand-written connector this bundle replaces at capability parity). It reads public SpaceX launch,
rocket, capsule, crew, payload, Starlink, launchpad, landpad, ship, roadster, and company data
through the public SpaceX API v4. Read-only, no-auth, matching the searxng golden's read-only
no-auth pattern.

## Auth setup

None. The SpaceX API is public and requires no credentials — `streams.json` `base.auth` declares
`[{"mode":"none"}]` unconditionally, matching legacy's `requester` (no `Auth` set on the
`connsdk.Requester`).

## Streams notes

All 11 streams (`launches`, `rockets`, `capsules`, `crew`, `payloads`, `starlink`, `launchpads`,
`landpads`, `ships`, `roadster`, `company`) are simple `GET` reads with no pagination (legacy's
`readRecords` issues exactly one request per stream, no page-advance loop) and no query
parameters at all — matching legacy's `readRecords(ctx, r, resource, endpoint.recordsPath, emit)`
call, which passes a nil `url.Values`. Every stream's `records.path` is `"."` (body root):
`launches` through `ships` return a top-level JSON array; `roadster` and `company` return a
top-level JSON object (`single_object: true`, both singleton resources with no list wrapper).

## Write actions & risks

None. SpaceX API is read-only in both legacy and this bundle (`capabilities.write: false`, no
`writes.json`).

## Known limits

- Legacy also supports two ad hoc request-shaping config knobs not modeled here: an `options`
  config value that redirects the `launches` stream to a `launches/<option>` sub-resource (e.g.
  `launches/upcoming`, `launches/past`), and an `id` config value that appends `/<id>` to any
  stream's resource path for a single-record-by-id lookup. Neither is expressible in this
  dialect: `streams.json`'s `path` templating has no absent-key-falsy tolerance for path segments
  (unlike `stream.Query`'s opt-in `omit_when_absent` object form) — conditionally appending a path
  segment only when a config value happens to be set would require the key to always resolve,
  which defeats the "optional" behavior legacy provides. Both knobs are intentionally not
  declared in `spec.json` (a declared-but-unwireable key is worse than an absent one, per the
  searxng `subreddit` precedent) and are out of scope for this wave (`api_surface.json`'s
  `duplicate_of`/`out_of_scope` entries for `/v4/launches/{id}`, `/v4/launches/latest`,
  `/v4/launches/next`) — the full-list `launches` stream (this bundle's actual implementation)
  already contains every record the `/v4/launches/{id}` and `options`-scoped sub-resource
  endpoints would return, so no read capability is lost, only the convenience of narrowing the
  request server-side.
- Only the 11 legacy-parity streams are implemented; the query-DSL endpoint (`POST
  /v4/<resource>/query`) is out of scope for this wave.
