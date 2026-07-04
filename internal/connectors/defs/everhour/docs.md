# Overview

Everhour is a time-tracking and project-management API (https://api.everhour.com). This bundle
originally migrated the read-only legacy `internal/connectors/everhour` package's streams
(`projects`, `clients`, `users`, `time`, `tasks`) to a Tier-1 defs bundle; it is full-refresh only
(no incremental cursor) for those 5, matching legacy exactly. `tasks` was previously blocked
(`docs/migration/status.json`'s `partial[]` ledger: "legacy 'tasks' substream is a sub-resource
fan-out read ... the declarative dialect's path/pagination/records fields have no mechanism to
drive a per-parent-record child-request loop") — this was closed by the engine's `fan_out` dialect
addition (S4 mini-wave item 2, `docs/migration/conventions.md` §3), which names everhour
explicitly as one of the connectors it retires the Tier-2 `StreamHook` requirement for. No Go
hook was needed; `tasks` is expressed entirely in `streams.json`.

This Pass B full-surface expansion (against the real Everhour API reference,
https://everhour.docs.apiary.io/) adds 6 more read streams (`sections`, `time_off_types`,
`allocations`, `expense_categories`, `expenses`, `invoices`) and 17 write actions covering the
client/project/task/section/time-record/expense lifecycle. `capabilities.write` is now `true`.

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

**Pass B additions:**

- `sections` (`GET /projects/{project_id}/sections`, Pass B addition) — the same per-project
  fan_out shape as `tasks`, reusing the identical `fan_out.ids_from.request` preliminary `GET
  /projects` listing. `project_id` is stamped onto every emitted section record. Primary key `id`
  (a genuine numeric section id, unlike the string-prefixed project/client ids — see the `id`
  type-coercion note below, which does NOT apply to `sections` since the real API's section id is
  confirmed numeric in the docs' own examples, not ambiguous).
- `time_off_types` (`GET /resource-planner/time-off-types`) and `allocations` (`GET /allocations`)
  — single-request, full-array GET endpoints with no pagination, same shape as `projects`/
  `clients`/`users`/`time`. Both have genuine numeric ids per the docs.
- `expense_categories` (`GET /expenses/categories`) and `expenses` (`GET /expenses`) — same
  single-request full-array shape; both have genuine numeric ids per the docs.
- `invoices` (`GET /invoices`) — same single-request full-array shape; genuine numeric id per the
  docs (`clientId`/`publicId` are separate string-shaped fields on the same record, not the
  primary key).

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

Everhour now supports 17 write actions, added in the Pass B full-surface expansion. All are
single-request JSON-body mutations with no compound follow-up requests, so all are fully Tier-1
declarative (no hook needed):

- **Clients**: `create_client` (POST /clients, low-risk), `update_client` (PUT
  /clients/{id}, low-risk), `delete_client` (DELETE /clients/{id}, irreversible, approval
  required).
- **Projects**: `create_project` (POST /projects, low-risk), `update_project` (PUT
  /projects/{id}, low-risk), `archive_project` (PATCH /projects/{id}/archive — hides the project
  from active lists and blocks new time entries while archived, approval required),
  `delete_project` (DELETE /projects/{id}, irreversible, approval required).
- **Sections**: `create_section` (POST /projects/{project_id}/sections, low-risk),
  `delete_section` (DELETE /sections/{id}, irreversible, approval required). No `update_section`
  — excluded (`api_surface.json`, `out_of_scope`) as a narrower follow-up mutation beyond the core
  create/delete pair.
- **Tasks**: `create_task` (POST /projects/{project_id}/tasks, low-risk), `update_task` (PUT
  /tasks/{id}, low-risk), `delete_task` (DELETE /tasks/{id}, irreversible, approval required).
- **Time records**: `create_time_record` (POST /time, low-risk), `update_time_record` (PUT
  /time/{id} — may mutate an already-invoiced/locked entry, approval required),
  `delete_time_record` (DELETE /time/{id}, irreversible and can affect billing history, approval
  required).
- **Expenses**: `create_expense` (POST /expenses, low-risk), `delete_expense` (DELETE
  /expenses/{id}, irreversible, approval required). No `update_expense` — excluded
  (`api_surface.json`, `out_of_scope`) this pass for breadth-vs-cost triage.

Excluded from writes this pass: client/project budget and billing sub-object configuration
(narrow admin actions distinct from core CRUD), task custom fields, task time estimates, timers/
timecards/timesheets (live/workflow-state actions, not batch data mutations), resource-planner
scheduling assignments and time-off allocations (HR/leave-balance consequences, no write extended
this pass), webhooks (no list/discovery endpoint exists — see Known limits), invoice status
transitions/line-item resets/exports/deletes (financial-document mutations with real accounting
consequences), and file attachments (binary payload, this dialect has no multipart/base64 body
construction mechanism). See `api_surface.json` for the full per-endpoint disposition.

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
  throttle). Everhour's own docs mention an approximate 20-requests-per-10-seconds limit, informational
  only (never enforced client-side by legacy), matching the informational-vs-enforced distinction
  conventions.md documents for `metadata.json.rate_limit`.
- **Webhooks are entirely excluded, both read and write (Pass B, `requires_elevated_scope`/
  `out_of_scope`).** Everhour's real API has no `GET /hooks` list endpoint — only
  `GET/PUT/DELETE /hooks/{hook_id}` by an id the caller must already have from elsewhere (identical
  shape to bitly's `custom_bitlinks` precedent in `docs/migration/conventions.md`'s worked
  examples). Without a discovery path there is no way to enumerate existing webhooks as a syncable
  stream, and registering a new outbound event-delivery URL sight-unseen (no way to read back and
  reconcile which webhooks already exist) was judged out of scope for this pass.
- **File attachments are excluded (Pass B, `binary_payload`).** Everhour's attachment
  create/download endpoints carry base64-encoded file bytes or return raw binary content; this
  dialect's write bodies are JSON/form field maps with no multipart/binary-payload construction
  mechanism, and the read path decodes JSON response bodies only.
- **`sections`' primary key `id` is a genuine numeric value, unlike `projects`/`clients`' string-
  prefixed ids — this is NOT the same ENGINE_GAP as the `id` type-coercion note above.** The
  Everhour docs' own example Section objects show a bare integer `id` (no `as:`/`cl:`-style
  platform prefix), so this bundle declares `sections.id` as `type: "integer"` directly rather than
  inheriting the `["string"]`-only workaround `tasks`/`users`/`time` carry forward from legacy's
  uniform `stringField` coercion — there is no legacy Go mapper for `sections` to be parity-bound
  to in the first place (it is a Pass B addition with no prior Go implementation), so this bundle's
  schema is free to reflect the real wire type directly. Same reasoning applies to
  `time_off_types`/`allocations`/`expense_categories`/`expenses`/`invoices` — all newly-added
  streams with confirmed-numeric ids and no legacy string-coercion precedent to preserve.
