# Overview

FullStory is a digital-experience analytics platform. This bundle reads FullStory segments, users,
events, and user-scoped sessions through the FullStory Server API (`https://api.fullstory.com`) and
can write server-side user attributes and custom events. It migrates
`internal/connectors/fullstory` (the hand-written legacy connector, which stays registered and
unchanged until wave6's registry flip) to a declarative bundle and expands the Pass B surface where
the API shape is ordinary JSON.

## Auth setup

Provide a FullStory API key via the `api_key` secret. It is sent as `Authorization: Basic
<api_key>` (`auth: [{"mode": "api_key_header", "header": "Authorization", "value": "{{
secrets.api_key }}", "prefix": "Basic "}]`) — matching legacy's `connsdk.APIKeyHeader("Authorization",
apiKey, "Basic ")`, i.e. the raw key with a static `"Basic "` prefix, NOT RFC 7617 base64-encoded
Basic auth. The `api_key` value is never logged. See Known limits for the `uid` secret (declared,
not wired).

## Streams notes

Four streams are declared:

- `segments` — saved session filters. The legacy connector uses `/segments/v2`; the current
  FullStory docs publish the same list/get segment family under `/segments/v1`, so
  `api_surface.json` records both the documented and legacy path forms.
- `users` — identified users, records at `results`.
- `events` — captured/custom events, records at `results`.
- `sessions` — recent session replay URLs for a supplied `session_uid` and/or `session_email`;
  FullStory documents that this endpoint is user-scoped and returns a single page.

All streams are `GET` and emit records from the top-level `results` array. `segments`/`users`
cursor field is `created`; `events`' cursor field is `event_time` — both informational only,
matching legacy (FullStory's analytics API supports only full-refresh sync; no `incremental` block
is declared, so no server-side filter is ever applied). The `sessions` stream has no cursor field
because it is a point lookup/list for a caller-supplied user identity.

The three legacy streams follow FullStory's `next_page_token`/`pageToken` convention
(`pagination.type: cursor` with `token_path: next_page_token`, `cursor_param: pageToken`):
the next page's `pageToken` is read from the current response body's `next_page_token`, and
pagination stops when that field is empty — identical to legacy's `harvest` loop. Every request
sends `limit=200` (matches legacy's default `page_size`).

The `sessions` stream overrides pagination to `none`, matching FullStory's own statement that the
session lookup endpoint does not paginate.

## Write actions & risks

Three write actions are exposed because they are documented as ordinary JSON server-side capture
operations:

- `create_user` — `POST /v2/users`; creates or upserts a FullStory user profile by `uid`.
- `update_user` — `POST /v2/users/{{ record.id }}`; updates display fields or custom properties
  for an existing user id.
- `create_event` — `POST /v2/events`; creates a custom event with optional `user`, `session`, and
  `properties` objects.

These writes enrich FullStory analytics dimensions and require the normal reverse-ETL plan,
preview, and approval flow. Destructive user deletion, async batch/stream imports, and generated
AI/session-summary endpoints are not exposed as writes; each is explicitly excluded in
`api_surface.json`.

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
- `sessions` is not a global account-wide session stream. FullStory requires a `uid` and/or
  `email` query to retrieve recent sessions for one user; this bundle wires optional
  `session_uid`/`session_email` config values for that documented lookup. Calling the stream
  without either value will send an unfiltered request and the live API may reject it.
- Async import/export workflows, raw file downloads, privacy/settings administration, beta
  element/extraction rule management, and AI-generated context/summary endpoints are enumerated in
  `api_surface.json` but excluded. They require job polling, binary/download handling, elevated
  administrative scope, or generated-output semantics beyond this connector's sync/write surface.
