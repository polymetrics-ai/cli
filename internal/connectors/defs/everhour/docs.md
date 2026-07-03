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

**`id` type-coercion gap (ENGINE_GAP, partial).** Legacy's `mapRecord` functions
(`internal/connectors/everhour/streams.go`) all route `id` through a shared `stringField` helper
applied *unconditionally* across all 5 record mappers — a pure type coercion (pass a string
through byte-for-byte; stringify anything else via `fmt.Sprintf("%v", v)`; empty string for nil)
that never inspects the value's content. That helper only exists because Everhour's real wire
shape is not uniform across endpoints: `/projects` and `/clients` return prefixed string ids
(`"as:123"`, `"cl:456"` — confirmed by legacy's own recorded test fixtures,
`internal/connectors/everhour/everhour_test.go`), while `/team/users` and `/team/time` return
plain numeric ids on the real API. Legacy's helper guarantees every stream emits a string `id`
regardless of this split.

This bundle CANNOT safely reproduce that coercion with the current engine dialect and is
therefore a genuine `ENGINE_GAP`, not a silent workaround:
- A bare `computed_fields` reference (`"id": "{{ record.id }}"`) now performs *typed* extraction
  (conventions.md's typed-extraction rule) — it copies the raw JSON value verbatim, so a numeric
  `users`/`time` id would pass through as a native number, not a coerced string, silently
  diverging from legacy.
- Every filter that WOULD force stringification has a real, corrupting side effect on at least
  one of Everhour's actual id shapes: `urlencode` percent-encodes the `:` in `projects`/`clients`
  ids (`"as:123"` -> `"as%3A123"`); `last_path_segment` truncates on `/` and was already flagged
  as a blocker-severity misuse for this exact substitution pattern on a non-URI id field (see
  `internal/connectors/defs/hibob`'s review finding) — reusing it here would repeat the identical
  defect class; `join:<sep>` hard-errors on a non-array value; `base64`/`unix_seconds`/`const:`
  are unrelated to this shape.
- No dialect mechanism lets one `computed_fields` entry read another's already-coerced output
  (each template resolves against the raw pre-projection record only), so there is no way to
  stage a stringify step through an intermediate field either.

Given this, `id` stays declared `type: "string"` in every schema (matching legacy's guaranteed
emitted contract) and fixtures keep string `id` values for all 4 streams — `projects`/`clients`
fixtures reflect Everhour's real string-prefixed wire ids directly (no coercion needed, verified
against legacy's own recorded test data); `users`/`time` fixtures are schema-conforming
placeholders, not a recorded proof of those two endpoints' real wire shape, because this bundle
cannot yet prove or safely reproduce whatever coercion the real API's numeric id would need. If
`/team/users` or `/team/time` genuinely returns a bare JSON integer for `id` in production, a live
sync of this bundle (unlike legacy) would emit that field as a native number instead of a string —
a real, open parity gap, not a cosmetic one, until the engine gains a pure stringify-coercion
filter (or an equivalent mechanism) with no side effects on non-numeric input. Tracked here rather
than left undocumented; see `blockers[]` in this migration's result record.

## Write actions & risks

None. Everhour is read-only in both legacy and this bundle (`capabilities.write: false`, no
`writes.json`).

## Known limits

- **`users`/`time` `id` type-coercion (ENGINE_GAP, partial) — see Streams notes above** for the
  full analysis. Summary: legacy uniformly coerces every stream's `id` to a string via
  `stringField`; this bundle cannot reproduce that coercion without a filter that corrupts a
  different Everhour id shape (`urlencode` mangles `projects`/`clients`' `:`; `last_path_segment`
  truncates on `/`, already a blocker-severity misuse elsewhere), so if `/team/users`/`/team/time`
  genuinely returns numeric wire ids in production, a live sync would emit a native number instead
  of legacy's guaranteed string. `id` schemas stay `type: "string"` and fixtures are
  schema-conforming placeholders for these two streams, not recorded proof of the real wire shape.
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
