# Overview

Everhour is a time-tracking and project-management API (https://api.everhour.com). This bundle
migrates the read-only legacy `internal/connectors/everhour` package's streams to a Tier-1 defs
bundle: `projects`, `clients`, `users`, `time`, and `tasks`. It is full-refresh only (no
incremental cursor), matching legacy exactly. `tasks` was previously blocked (`docs/migration/
status.json`'s `partial[]` ledger: "legacy 'tasks' substream is a sub-resource fan-out read ...
the declarative dialect's path/pagination/records fields have no mechanism to drive a
per-parent-record child-request loop") — this is now closed by the engine's `fan_out` dialect
addition (S4 mini-wave item 2, `docs/migration/conventions.md` §3), which names everhour
explicitly as one of the connectors it retires the Tier-2 `StreamHook` requirement for. No Go
hook was needed; `tasks` is expressed entirely in `streams.json`.

## Auth setup

Provide an Everhour API key via the `api_key` secret; it is sent as the `X-Api-Key` header on
every request (never logged). No `Authorization` header is ever sent. `base_url` defaults to
`https://api.everhour.com` and only needs overriding for tests or proxies.

## Streams notes

`projects`, `clients`, `users`, and `time` are single-request, full-array GET endpoints with no
pagination and no incremental cursor, exactly matching legacy's `readTopLevel` behavior:

- `projects` (`GET /projects`, `records.path: ""` — the response is a bare top-level JSON array,
  matching legacy's `connsdk.RecordsAt(resp.Body, "")` call) — primary key `id`.
- `clients` (`GET /clients`) — primary key `id`.
- `users` (`GET /team/users`) — primary key `id`. This is also the `check` request (mirrors
  legacy's `Check`, which lists team members to confirm auth/connectivity without mutating
  anything).
- `time` (`GET /team/time`) — primary key `id`.

`tasks` reproduces legacy's `readSubstream` (`internal/connectors/everhour/everhour.go`) via
`streams.json`'s `fan_out`: a preliminary `GET /projects` request (`fan_out.ids_from.request`,
`records_path: ""` — the same bare top-level array `projects` itself reads) extracts every
project's `id`; the engine then repeats `GET /projects/{{ fanout.id }}/tasks` once per resolved
id (`fan_out.into.path_var`), stamping the parent id onto every emitted task record via
`fan_out.stamp_field: "project_id"` — the identical field name legacy's `readSubstream` stitches
on (`endpoint.parentIDField = "project_id"`). Each project's task sub-sequence is independent
(fresh pagination/incremental state per id, per the engine's fan-out contract), matching legacy's
per-project HTTP loop exactly. Primary key `id`.

**Path-encoding note (accepted, not a deviation).** `stream.Path`'s `{{ fanout.id }}` reference
resolves through `InterpolatePath`, which urlencode-encodes every path segment by default
(`docs/migration/conventions.md` §3) — this is the SAME default every other templated path
segment in this dialect gets, not something fan_out-specific. A real Everhour project id
containing a literal `:` (legacy's own recorded shape, e.g. `"as:123"` —
`internal/connectors/everhour/everhour_test.go`) is therefore requested as
`/projects/as%3A123/tasks`, not legacy's unencoded `/projects/as:123/tasks`. Percent-encoding a
reserved path-segment character (RFC 3986 `pchar` includes `:`) is standards-correct and virtually
every HTTP server/router decodes it back to `:` before route matching — this is expected to be
functionally identical on the wire, not a data-changing deviation, but is called out here since it
could not be verified against a live Everhour endpoint in this migration.

**`id` type-coercion gap (ENGINE_GAP, partial — carried from the prior migration attempt, still
open).** Legacy's `mapRecord` functions (`internal/connectors/everhour/streams.go`) all route `id`
through a shared `stringField` helper applied *unconditionally* across all 5 record mappers — a
pure type coercion (pass a string through byte-for-byte; stringify anything else via
`fmt.Sprintf("%v", v)`; empty string for nil) that never inspects the value's content. That helper
only exists because Everhour's real wire shape is not uniform across endpoints: `/projects` and
`/clients` return prefixed string ids (`"as:123"`, `"cl:456"` — confirmed by legacy's own recorded
test fixtures, `internal/connectors/everhour/everhour_test.go`), while `/team/users`, `/team/time`,
and `/projects/<id>/tasks` are not confirmed either way by any recorded legacy test (legacy's own
tests only show alphanumeric task ids like `"t1"`) — so a numeric wire id on any of these three
streams remains a plausible, unverified real-world shape. Legacy's helper guarantees every stream
emits a string `id` regardless.

This bundle still cannot safely reproduce that coercion with the current engine dialect:
- A bare `computed_fields` reference (`"id": "{{ record.id }}"`) performs *typed* extraction
  (conventions.md's typed-extraction rule) — it copies the raw JSON value verbatim, so a numeric
  id would pass through as a native number, not a coerced string, silently diverging from legacy.
- Every filter that WOULD force stringification has a real, corrupting side effect on at least one
  of Everhour's actual id shapes: `urlencode` percent-encodes the `:` in `projects`/`clients` ids
  (`"as:123"` -> `"as%3A123"`); `last_path_segment` truncates on `/` and was already flagged as a
  blocker-severity misuse for this exact substitution pattern on a non-URI id field (see
  `internal/connectors/defs/hibob`'s review finding, `docs/migration/wave2-review-raw.json`) —
  reusing it here would repeat the identical defect class; `join:<sep>` hard-errors on a
  non-array value; `base64`/`unix_seconds`/`const:` are unrelated to this shape.
- No dialect mechanism lets one `computed_fields` entry read another's already-coerced output
  (each template resolves against the raw pre-projection record only), so there is no way to stage
  a stringify step through an intermediate field either.

Given this, `id` stays declared `type: "string"` in every schema (matching legacy's guaranteed
emitted contract) and fixtures keep string `id` values for all 5 streams — `projects`/`clients`
fixtures reflect Everhour's real string-prefixed wire ids directly (no coercion needed, verified
against legacy's own recorded test data); `users`/`time`/`tasks` fixtures are schema-conforming
placeholders, not recorded proof of those endpoints' real wire shape, because this bundle cannot
yet prove or safely reproduce whatever coercion a numeric real-world id would need. If any of
those three streams genuinely returns a bare JSON integer for `id` in production, a live sync of
this bundle (unlike legacy) would emit that field as a native number instead of a string — a real,
open parity gap, not a cosmetic one, until the engine gains a pure stringify-coercion filter (or
an equivalent mechanism) with no side effects on non-numeric input. This is unchanged from the
prior migration attempt's analysis and is carried forward, not newly introduced by the `tasks`
fan-out addition.

## Write actions & risks

None. Everhour is read-only in both legacy and this bundle (`capabilities.write: false`, no
`writes.json`).

## Known limits

- **`users`/`time`/`tasks` `id` type-coercion (ENGINE_GAP, partial) — see Streams notes above**
  for the full analysis. Summary: legacy uniformly coerces every stream's `id` to a string via
  `stringField`; this bundle cannot reproduce that coercion without a filter that corrupts a
  different Everhour id shape (`urlencode` mangles `projects`/`clients`' `:`; `last_path_segment`
  truncates on `/`, already a blocker-severity misuse elsewhere), so if `/team/users`/`/team/time`/
  `/projects/<id>/tasks` genuinely returns numeric wire ids in production, a live sync would emit
  a native number instead of legacy's guaranteed string. `id` schemas stay `type: "string"` and
  fixtures are schema-conforming placeholders for these streams, not recorded proof of the real
  wire shape.
- **`tasks`' fan-out path segment is urlencoded by default (accepted) — see Streams notes above.**
  A project id containing `:` is requested percent-encoded (`as%3A123`), not legacy's unencoded
  form; standards-correct and expected to be functionally identical, not verified live.
- `rate_limit` is not declared on `streams.json`'s `base` block: legacy enforces no client-side
  rate limiting, so none is added here (matches legacy's actual behavior, not a new introduced
  throttle).
