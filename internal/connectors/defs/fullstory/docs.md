# Overview

FullStory is a digital-experience analytics platform. This bundle reads FullStory segments, users,
and events through the FullStory REST API (`https://api.fullstory.com`), migrating
`internal/connectors/fullstory` (the hand-written legacy connector, which stays registered and
unchanged until wave6's registry flip) to a declarative bundle at capability parity. FullStory is
read-only here — no write actions.

## Auth setup

Provide a FullStory API key via the `api_key` secret. It is sent as `Authorization: Basic
<api_key>` (`auth: [{"mode": "api_key_header", "header": "Authorization", "value": "{{
secrets.api_key }}", "prefix": "Basic "}]`) — matching legacy's `connsdk.APIKeyHeader("Authorization",
apiKey, "Basic ")`, i.e. the raw key with a static `"Basic "` prefix, NOT RFC 7617 base64-encoded
Basic auth. The `api_key` value is never logged. See Known limits for the `uid` secret (declared,
not wired).

## Streams notes

All 3 streams (`segments`, `users`, `events`) share the same shape: `GET`, records at the
top-level `results` array, primary key `["id"]`. `segments`/`users` cursor field is `created`;
`events`' cursor field is `event_time` — both informational only, matching legacy (FullStory's
analytics API supports only full-refresh sync; no `incremental` block is declared, so no
server-side filter is ever applied).

Pagination follows FullStory's `next_page_token`/`pageToken` convention
(`pagination.type: cursor` with `token_path: next_page_token`, `cursor_param: pageToken`):
the next page's `pageToken` is read from the current response body's `next_page_token`, and
pagination stops when that field is empty — identical to legacy's `harvest` loop. Every request
sends `limit=200` (matches legacy's default `page_size`).

## Write actions & risks

None. FullStory is exposed read-only (`capabilities.write: false`), matching legacy's
`Capabilities{Write: false}` and its `Write` method returning `connectors.ErrUnsupportedOperation`
unconditionally.

## Known limits

- **`uid` secret is declared but NOT wired into any request (documented scope narrowing)**: legacy
  optionally sends an `FS-Uid` header sourced from the `uid` secret, `DefaultHeaders["FS-Uid"] =
  uid` only when `uid` is non-empty; when unset (the common case, and the only case any legacy test
  exercises — no test in `fullstory_test.go` asserts the `FS-Uid` header at all), no header is
  sent. The engine's header-resolution dialect
  (`internal/connectors/engine/read.go`'s `classifyHeaderResolutionError`) makes an unresolved
  `secrets.*` reference in a declared header ALWAYS a hard error — never silently omitted, by
  design (F4: sending a request unauthenticated instead of failing loudly is the exact failure mode
  the engine deliberately rejects). There is no dialect mechanism to express "send this header only
  when this secret happens to be set, otherwise omit it entirely" (unlike `auth`'s `when` clause,
  which DOES tolerate absent-secret-is-falsy — but `when` only gates `auth` candidates, not
  arbitrary `streams.json` headers). Declaring `FS-Uid: {{ secrets.uid }}` in `base.headers` would
  make every read hard-fail whenever `uid` is unset, which is a regression for the overwhelmingly
  common case; this bundle instead omits the header entirely, matching legacy's own
  `uid`-unset behavior exactly and diverging only when an operator sets a `uid` (in which case the
  FS-Uid header would be sent by legacy but not by this bundle). This is an honest, ledgered scope
  narrowing per `docs/migration/conventions.md` §5, not silently faked; closing it fully would need
  a dialect extension (an optional per-header `when`/`omit_when_absent` clause mirroring
  `stream.Query`'s object form) — an `ENGINE_GAP` if `uid` usage becomes load-bearing.
- No incremental sync: FullStory's REST API has no documented server-side "created/updated since"
  filter for these endpoints, matching legacy (`InitialState` always returns an empty cursor).
  `client_filtered` incremental was considered and rejected: adding client-side cursor filtering
  here would be new behavior legacy never had, not a migration.
