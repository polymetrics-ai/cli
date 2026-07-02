# Overview

AppFollow is a wave2 fan-out declarative-HTTP migration. It reads AppFollow account users and app
collections through the read-only AppFollow REST API v2
(`GET https://api.appfollow.io/api/v2/...`). This bundle targets capability parity with
`internal/connectors/appfollow` (the hand-written connector it migrates) for the `users` and
`app_collections` streams only; the legacy package stays registered and unchanged until wave6's
registry flip, and remains authoritative for the `app_lists`/`ratings` streams this bundle does
not cover (see Known limits).

## Auth setup

Provide an AppFollow API token via the `api_secret` secret; it is sent as the
`X-AppFollow-API-Token` header and is never logged, matching legacy's
`connsdk.APIKeyHeader(appfollowTokenHeader, secret, "")` (`appfollow.go:322`). `base_url` defaults
to `https://api.appfollow.io/api/v2` and may be overridden for tests/proxies.

## Streams notes

Both streams are single-request, unpaginated `GET` endpoints (`pagination` omitted; legacy's
reporting endpoints return their full result in one body, matching `readSimple`):

- `users` — `GET /account/users`, records at the response root (`records.path: ""`), primary key
  `["id"]`.
- `app_collections` — `GET /account/apps`, records at the `apps` array, primary key `["id"]`.

## Write actions & risks

None. Legacy's `Write` unconditionally returns `connectors.ErrUnsupportedOperation`;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`app_lists` and `ratings` are not modeled as streams in this bundle (ENGINE_GAP).** Both are
  sub-resource fan-out reads: legacy's `readAppLists` (`appfollow.go:157`) issues one
  `GET /account/apps/app?apps_id=<id>` request per configured (or auto-discovered)
  `app_collection_ids` entry, and `readRatings` (`appfollow.go:205`) issues one
  `GET /meta/ratings?ext_id=<id>` request per configured `ext_ids` entry, flattening each
  response's nested rows and stamping the enclosing id onto every emitted record. The engine's
  declarative read path (`internal/connectors/engine/read.go`) drives exactly one paginated
  request sequence per stream against a single templated `path` — there is no primitive for
  "issue N independent request sequences, one per value in a config-supplied list, and merge their
  results." This is one of conventions.md's named Tier-2 `StreamHook` triggers ("sub-resource
  fan-out reads", e.g. issue -> comments per issue), and Tier-2/Tier-3 escape hatches are out of
  scope for this wave's fan-out task — legacy stays authoritative for these two streams.

- **Config-driven collection-id discovery is not modeled.** Legacy's `app_lists` stream falls back
  to auto-discovering collection ids from `/account/apps` when `app_collection_ids` is unset
  (`discoverCollectionIDs`, `appfollow.go:185`) — moot here since the stream itself is not
  implemented (see above), but recorded so a future capability-expansion agent implementing
  `app_lists` via a `StreamHook` knows to port this fallback too, not just the happy path.
