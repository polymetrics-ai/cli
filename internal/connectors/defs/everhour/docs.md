# Overview

Everhour is a time-tracking and project-management API (https://api.everhour.com). This bundle
migrates the read-only legacy `internal/connectors/everhour` package's declaratively-expressible
streams to a Tier-1 defs bundle: `projects`, `clients`, `users`, and `time`. It is full-refresh
only (no incremental cursor), matching legacy exactly. The legacy `tasks` substream is NOT ported
here — see Known limits.

## Auth setup

Provide an Everhour API key via the `api_key` secret; it is sent as the `X-Api-Key` header on
every request (never logged). No `Authorization` header is ever sent. `base_url` defaults to
`https://api.everhour.com` and only needs overriding for tests or proxies.

## Streams notes

All 4 migrated streams are single-request, full-array GET endpoints with no pagination and no
incremental cursor, exactly matching legacy's `readTopLevel` behavior:

- `projects` (`GET /projects`, `records.path: ""` — the response is a bare top-level JSON array,
  matching legacy's `connsdk.RecordsAt(resp.Body, "")` call) — primary key `id`.
- `clients` (`GET /clients`) — primary key `id`.
- `users` (`GET /team/users`) — primary key `id`. This is also the `check` request (mirrors
  legacy's `Check`, which lists team members to confirm auth/connectivity without mutating
  anything).
- `time` (`GET /team/time`) — primary key `id`.

Every `id` field is emitted as a string in both legacy (via `stringField`) and Everhour's real
wire shape (e.g. `"as:123"`), so no `computed_fields` coercion is needed — plain schema
projection by matching key name reproduces legacy's mapped record exactly for all 4 streams.

## Write actions & risks

None. Everhour is read-only in both legacy and this bundle (`capabilities.write: false`, no
`writes.json`).

## Known limits

- **`tasks` is not ported (blocked, ENGINE_GAP-adjacent).** Legacy's `tasks` stream is a
  sub-resource fan-out read: it first lists `/projects`, then issues one
  `GET /projects/<id>/tasks` request per project, stitching the parent `project_id` onto every
  child task record (`internal/connectors/everhour/everhour.go`'s `readSubstream`). The
  declarative dialect (`streams.json`'s `path`/`pagination`/`records` fields) has no mechanism to
  drive a per-parent-record child request loop within a single stream read — this is a named
  Tier-2 `StreamHook` trigger per `docs/migration/conventions.md` §1's Tier-2 table ("sub-resource
  fan-out reads"), and this wave is JSON-only (no `hooks/` packages permitted). A follow-up
  hooks-capable wave should add `internal/connectors/hooks/everhour/hooks.go` with a `StreamHook`
  reproducing `readSubstream` exactly, then add the `tasks` stream + schema here. Everhour's
  `metadata.json.description` and `api_surface.json` both call this out explicitly so it is not
  mistaken for silently-dropped scope.
- `rate_limit` is not declared on `streams.json`'s `base` block: legacy enforces no client-side
  rate limiting, so none is added here (matches legacy's actual behavior, not a new introduced
  throttle).
