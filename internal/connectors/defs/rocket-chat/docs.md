# Overview

Rocket.Chat is a wave2 fan-out declarative-HTTP migration. It reads Rocket.Chat users, public
channels, private groups, direct-message rooms, and the rooms "changed" feed through the
Rocket.Chat REST API v1 (`GET https://<server>/api/v1/...`). This bundle is migrated from
`internal/connectors/rocket-chat` (the hand-written connector it replaces at capability parity);
the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Rocket.Chat personal access token pair via the `auth_token` and `user_id` secrets; both
are sent as headers (`X-Auth-Token: <auth_token>`, `X-User-Id: <user_id>`) on every request and are
never logged, matching legacy's `requester` (`rocket_chat.go:172-176`). Both secrets are required —
an absent value on either header is a hard error (never silently omitted), matching legacy's own
`"rocket-chat connector requires secrets auth_token and user_id"` check.

`base_url` is required (legacy has no default) and must be the full server URL including the
`/api/v1` path segment (e.g. `https://chat.example.com/api/v1`). Legacy derives this by appending
`/api/v1` in code when the caller's URL doesn't already end with it
(`rocket_chat.go:251-254`); the engine's declarative `base.url` template has no equivalent
suffix-append mechanism, so this bundle narrows the config surface to require the full path
up front — see Known limits.

## Streams notes

All 5 streams are simple `GET` list/feed endpoints:

- `users` (`GET /users.list`, records at `users`)
- `channels` (`GET /channels.list`, records at `channels`)
- `groups` (`GET /groups.list`, records at `groups`)
- `direct_messages` (`GET /im.list`, records at `ims`)
- `rooms` (`GET /rooms.get`, records at `update` — Rocket.Chat's "changed rooms" feed shape, which
  also returns a `remove` array this bundle does not surface, matching legacy's exact `recordsAt`
  lookup of only the `update` path)

Every stream's raw wire records key their identifier as `_id` and their last-modified timestamp as
`_updatedAt` (Rocket.Chat's real REST API shape); `computed_fields` renames both to this bundle's
schema-declared `id`/`updated_at` (matching legacy's own `mapRecord`'s `first(out, "id", "_id",
...)` fallback and its `updated_at` cursor field declaration). The `rooms` stream additionally
renames the raw `t` (room type) field to `type`. Every stream stamps a static-literal `stream`
marker field naming which stream a record came from, matching legacy's own `out["stream"] = stream`
(`rocket_chat.go:204`).

Every stream declares `"projection": "passthrough"` — legacy's `mapRecord` (`rocket_chat.go:196-206`)
copies every raw API field into the emitted record unfiltered (`for k, v := range rec { out[k] = v
}`) before adding `id`/`stream`, so this bundle matches that exactly rather than silently dropping
any raw field the schema doesn't declare (the default `"schema"` projection mode would do exactly
that). `schemas/*.json` document the fields this bundle actively derives/relies on; the full raw
Rocket.Chat wire shape (additional per-object fields not listed here) also survives on every emitted
record, matching legacy's real behavior.

`users`/`channels`/`groups`/`direct_messages` accept two optional passthrough filters — `query`
(Rocket.Chat's own MongoDB-style query-string filter) and `fields` (field-projection filter) — sent
only when configured (`omit_when_absent`), matching legacy's config-key passthrough
(`rocket_chat.go:87-91`). `rooms` accepts `room_id` (sent as `roomId`) and `updated_since` (sent as
`updatedSince`) the same way.

Pagination is offset+limit (`pagination.type: offset_limit`, `limit_param: count`,
`offset_param: offset`), matching legacy's `connsdk.OffsetPaginator{LimitParam: "count",
OffsetParam: "offset"}` short-page-stop semantics (`recordCount < page_size` stops pagination)
exactly. `streams.json`'s `pagination.page_size: 2` (vs. legacy's real default of 100) exists purely
to keep this bundle's committed 2-page conformance fixture small and reviewable (jira's identical
precedent, `docs/migration/conventions.md`) — `page_size` is a static bundle-authored JSON int with
no config-driven override on either side (see Known limits), so this is not a live-vs-fixture
behavior divergence, only a fixture-authoring convenience; a real Rocket.Chat server is paged
through 2 records at a time by this bundle rather than legacy's 100, functionally identical
(both fully exhaust every page) but chattier. None of Rocket.Chat's
list endpoints expose a server-side incremental filter parameter (legacy's own streams declare
`CursorFields` for catalog purposes only, with no `incremental` read-path wiring at all) — this
bundle likewise declares no `incremental` block for any stream; every read is full refresh, matching
legacy's real behavior exactly.

## Write actions & risks

None. Rocket.Chat's list endpoints are read-only in this bundle (legacy: `Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`base_url` requires the full `/api/v1` path segment; legacy's suffix-auto-append is not
  modeled.** Legacy accepts a bare server URL (e.g. `https://chat.example.com`) and appends
  `/api/v1` in code if the caller's URL doesn't already end with it (`rocket_chat.go:251-254`). The
  engine's `base.url` template resolves `config.base_url` verbatim with no derived-suffix
  mechanism (conventions.md's `spec.json default` materialization note: a DERIVED default — here, a
  conditional path suffix rather than a fixed literal — is not expressible via `"default"` alone).
  This bundle therefore requires the caller to supply the full URL including `/api/v1` up front — a
  documented config-surface narrowing, not a silent behavior change (an operator who previously
  supplied a bare host must now supply the full path).
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size` (1-1000,
  default 100) and `max_pages` (0/all/unlimited = unbounded) as config-driven overrides
  (`rocket_chat.go:258-264`). The engine's `offset_limit` paginator's `PaginationSpec.PageSize` is a
  static JSON number fixed at bundle-author time, not a `config.*`-templated value, and there is no
  `MaxPages`-equivalent config knob wired to a per-stream override either (only a bundle-wide
  `pagination.MaxPages` static cap, unused here to match legacy's default-unbounded behavior). This
  bundle sends a fixed page size (`count=2`, chosen for fixture-authoring convenience — see Streams
  notes; unbounded pages) rather than legacy's real default of 100; `page_size`/`max_pages` are not
  declared in `spec.json` at all (F6, REVIEW.md: a declared-but-unwireable config key is worse than
  an absent one).
- **`rooms`'s `remove` array is not surfaced.** Rocket.Chat's `rooms.get` "changed rooms" feed
  returns both an `update` array (upserted rooms) and a `remove` array (deleted room ids); legacy's
  own `recordsAt` only ever looks up the `update` path (`rocket_chat.go:110`, `endpoints["rooms"]`),
  so this bundle matches that exactly — deleted-room tombstones are out of scope on both sides.
