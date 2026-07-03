# Overview

OneSignal was quarantined during wave1 for an `ENGINE_GAP`: it legitimately needs two DIFFERENT
credential values used simultaneously across streams in the same bundle — the account-level
`apps` stream must always authenticate with `user_auth_key`, while `devices`/`notifications`/
`outcomes` must always authenticate with the per-app `app_api_key` (falling back to
`rest_api_key`). Every S3/S4 engine mini-wave increment since (typed `computed_fields`, config
vars in `computed_fields`, the optional-query dialect, `last_path_segment`, `stop_path` cursor
guard, spec-default materialization, 0-indexed `start_page`, `fan_out`, `keyed_object`, oauth2
`extra_params`, date-only incremental bounds) addresses pagination/incremental/query/auth-mode
shapes, but **none of them add a per-stream `auth`/`headers` override** — `engine/bundle.go`'s
`StreamSpec` still has no `Auth` or `Headers` field, and `read.go:474-475` resolves auth exactly
once per `Read`/`Check` call from the bundle-wide `b.HTTP.Auth` candidate list (`selectAuth`),
applied identically regardless of which stream is being read. This bundle is therefore a
**partial** unblock: only the `apps` stream (the account-level stream, `user_auth_key`-only) is
expressible in a single-auth bundle and is ported here. `devices`/`notifications`/`outcomes`
remain quarantined — porting them into THIS bundle would either silently send the wrong
credential to the wrong endpoint (a correctness bug, not a migration) or require a `AuthHook`/
`StreamHook` (Tier-2 Go), which is out of scope for this JSON+docs.md-only unblock task. The
legacy `internal/connectors/onesignal` package stays registered and authoritative for
`devices`/`notifications`/`outcomes` until a future increment adds a per-stream auth override (or
a Tier-2/3 escalation is approved).

## Auth setup

Provide a OneSignal user/organization auth key as the `user_auth_key` secret; it is sent as
`Authorization: Basic <user_auth_key>` (`api_key_header` auth mode with `header: "Authorization"`,
`prefix: "Basic "`) — the literal key is used directly, not base64 `user:pass` encoded, matching
legacy's `connsdk.APIKeyHeader("Authorization", key, "Basic ")` exactly. Never logged.

## Streams notes

`apps` (`GET /apps`) is the only ported stream. The response is a bare top-level JSON array
(`records.path: ""`), matching legacy's `onesignalStreamEndpoints["apps"].recordsPath == ""`
branch. Primary key is `id`. No `incremental` block is declared: legacy's `apps` endpoint has no
`CursorFields` published and `Read`'s `readSingle` path (used for `apps`, since it is not
`paginated`) sends no cursor/filter parameter at all — matching conventions.md §8 rule 2.

## Write actions & risks

None. OneSignal is exposed read-only in both legacy and this bundle (`capabilities.write:
false`, no `writes.json`).

## Known limits

- **`devices` (legacy: `players` resource) is not ported (blocked, `ENGINE_GAP`).** Legacy
  authenticates this endpoint with the per-app REST key (`app_api_key`, falling back to
  `rest_api_key`) and requires `app_id` — a DIFFERENT credential than `apps`' `user_auth_key`. The
  engine's `base.auth` is a single bundle-wide candidate list (`selectAuth`, first-match-wins by
  `when`), applied identically to every stream in the bundle; `StreamSpec` has no per-stream
  `auth`/`headers` field to override it. Declaring both credentials as `when`-gated candidates in
  the SAME `base.auth` list would not reproduce legacy's behavior either: legacy's `apps` stream
  MUST use `user_auth_key` even when `app_api_key` happens to also be configured (and vice versa
  for `devices`) — a `when`-ordered candidate list picks one winner per `Read` call, not one winner
  PER STREAM, so it cannot pin two streams in the same bundle to two independently-mandatory
  credentials simultaneously.
- **`notifications` is not ported (blocked, `ENGINE_GAP`).** Identical per-stream-credential gap
  as `devices` (also requires `app_api_key`/`rest_api_key` + `app_id`).
- **`outcomes` (legacy path `apps/{app_id}/outcomes`) is not ported (blocked, `ENGINE_GAP`).**
  Identical per-stream-credential gap as `devices`/`notifications`.
- `app_api_key`, `rest_api_key`, and `app_id` are deliberately NOT declared in `spec.json`: none of
  them is wired into any template in this `apps`-only bundle, and conventions.md F6 treats a
  declared-but-unwireable config/secret key as worse than an absent one. They will be reintroduced
  if/when `devices`/`notifications`/`outcomes` become expressible.
- Full OneSignal API surface (segments, templates, and any write/mutation endpoints) is otherwise
  out of scope; see `api_surface.json`'s `excluded` entries.
