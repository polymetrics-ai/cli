# Overview

DataScope is a field-data-collection platform (mobile forms, inspections, dispatch). This bundle
reads DataScope locations, form answers, lists, notifications, task assignments, tickets
(findings), and generated files, and writes location/list/task-assignment/form-answer mutations,
through the DataScope external REST API, migrating `internal/connectors/datascope` (the legacy
hand-written connector, which stays registered and unchanged until wave6's registry flip). As of
this Pass B full-surface expansion, every documented endpoint on the canonical API reference
(`https://dscope.github.io/docs/`) is covered — see `api_surface.json` for the endpoint-by-endpoint
accounting.

## Auth setup

Provide `api_key` as a secret; it is sent as the raw `Authorization` header value with no
Bearer/Basic prefix (`auth.mode: api_key_header`, `header: Authorization`, no `prefix`), matching
legacy's `connsdk.APIKeyHeader("Authorization", secret, "")` exactly. Never logged.

## Streams notes

The original 4 streams (`locations`, `answers`, `lists`, `notifications`) share the same shape:
`GET` against a DataScope endpoint, records at the response root (`records.path: "."` — every one
of these DataScope list endpoints returns a bare top-level JSON array), pagination `offset_limit`
(`limit_param: limit`, `offset_param: offset`), stopping on a short page.

New in this Pass B expansion:

- **`task_assigns`** (`GET /task_assigns`) — DataScope's dispatch/task-assignment records; unlike
  the 4 original streams, its envelope is a keyed object (`{"task_assigns": [...], "total", "limit",
  "offset"}`), so `records.path: "task_assigns"` selects the array rather than the response root.
  Shares the base `offset_limit` pagination.
- **`findings`** (`GET /findings/list`, DataScope's "Tickets" feature, formerly named "Issues" in
  their own docs) — records at the response root like the original 4 streams; the base
  `offset_limit` pagination applies unchanged (the real endpoint documents the same `limit`/
  `offset` query params as every other list endpoint here). Ticket ids are Firestore document ids
  (opaque strings, e.g. `"Krdz3aFoWZ4ZVgpuAart"`), so `x-primary-key` is typed `string`, not
  `integer`, unlike every other stream in this bundle. The raw API's `assignees`/`invitees` fields
  are object-keyed maps (`{"0": {...}, "1": {...}}`, an index-keyed dictionary, not a stable-shape
  object nor a JSON array) — this bundle's schema omits them (silently dropped by ordinary
  schema-mode projection) rather than fabricate a stable shape for a dictionary keyed by an
  unbounded, non-deterministic set of numeric-string keys; `assignees_concatenated`/
  `invitees_concatenated` (semicolon/ampersand-delimited flat strings carrying the same
  assignee/invitee data) are also omitted for the same reason — a bespoke delimited micro-format,
  not a JSON structure the dialect's `computed_fields` can decompose without a hook.
- **`files`** (`GET /files`) — DataScope's generated-file listing (PDF/report exports); the real
  endpoint documents only `start`/`end` query filters, no `limit`/`offset` pagination at all, so
  this stream overrides `pagination.type` to `none` (an honest per-stream override matching the
  files endpoint's actual, unpaginated shape — see `licenses` in the dbt bundle for the identical
  pattern applied to a different API).

No stream declares an `incremental` block for the reason already documented below (DataScope's
`start`/`end` window params use a non-standard datetime format and a "current wall-clock time"
value the engine cannot express) — this applies identically to `task_assigns`/`findings`/`files`,
all of which also expose `start`/`end`-shaped date-window filters in their real APIs.

No `computed_fields` are needed for any of the 7 streams: DataScope's own field names already match
this bundle's schema property names one-for-one (unlike, e.g., searxng's camelCase-vs-snake_case
rename case) — plain schema projection reproduces every field verbatim.

## Write actions & risks

9 write actions now cover every dialect-expressible DataScope mutation:

- **Locations**: `create_location`/`update_location` — field-data-collection location metadata
  (address, contact info, geocoordinates).
- **Task assignment**: `assign_task` (`POST /assign_task`) — creates a new field task/inspection
  assignment for a user on a scheduled date; low-risk (idempotent-adjacent — assigning a task has
  no destructive counterpart in this bundle).
- **Lists (metadata)**: `create_metadata_object`/`update_metadata_object` (individual list
  elements), `create_metadata_type`/`update_metadata_type` (the lists themselves), and
  `bulk_update_metadata_objects` (many elements of one list in a single call — a materially larger
  blast radius than a single-object update, so it carries its own elevated risk string and always
  requires approval).
- **Form answers**: `change_form_answer` (`POST /change_form_answer`) — overwrites a previously
  submitted form answer's value in place by `form_name`/`form_code`/`question_name` (optionally
  `subform_index` for repeating subform sections). This is a genuine after-the-fact rewrite of
  already-collected field data (not merely creating a new answer), so it always requires approval.

There is no `delete_*` action for any DataScope resource: the real API documents no delete endpoint
for locations, lists, task assignments, or tickets — only create/update. `capabilities.write` is
now `true` (previously `false`); `metadata.json`'s `risk.write`/`risk.approval` document per-action
risk tiers.

## Known limits

- **`answers`/`notifications`/`task_assigns`/`findings`/`files`' incremental date-window filtering
  is not modeled (documented scope narrowing, not a silent behavior change).** Two independent gaps
  compound here, identical to the original 4-stream analysis: (1) DataScope's window params use a
  non-standard `dd/mm/yyyy HH:MM` datetime format; the engine's `incremental.param_format` dialect
  supports only `rfc3339`/`unix_seconds`/`date` (`2006-01-02`)/`github_date_range` — none matches
  DataScope's layout, and there is no custom-format escape hatch. (2) Legacy's `end` parameter for
  the original windowed streams is always the CURRENT wall-clock time at request time
  (`time.Now().UTC()`), not a value derived from config/state/incremental lower bound — the
  engine's template `Vars` environment has no "current time" reference at all. Both gaps would need
  new engine dialect surface to close correctly; this bundle narrows scope to full-refresh-only
  reads instead of faking the window (see the original 4-stream analysis for the full reasoning,
  which applies identically to every newly-added windowed stream). `x-cursor-field` is still
  declared where legacy/DataScope's own catalog metadata implies a natural cursor (`created_at` on
  `task_assigns`), but no `incremental` block is declared, so only full-refresh sync modes apply.
- **`findings`' `assignees`/`invitees` object-keyed dictionary fields (and their
  `*_concatenated` delimited-string counterparts) are not migrated.** Neither is a JSON array (so
  `records.keyed_object`, which explodes a top-level records envelope, does not apply at the
  per-field level anyway) nor a fixed-shape object; they are index-keyed dictionaries
  (`{"0": {...}, "1": {...}}`) whose key set is unbounded and record-dependent. There is no
  declarative way to flatten an arbitrary-cardinality keyed dictionary FIELD (as opposed to a
  keyed-object records ENVELOPE, which the engine does support) into stable schema properties
  without either a `RecordHook` (a 3rd hook interface this Tier-1 bundle does not have, and does
  not need for anything else) or emitting a variable number of dynamically-named columns (which
  schema-mode projection cannot do at all — every column must be statically declared). Silently
  dropped, matching this bundle's existing schema-projection-drops-undeclared-fields behavior for
  every other stream.
- Webhook registration (`/webhooks/new`) has no programmatic API endpoint at all — DataScope's own
  docs describe it as a web-UI-only configuration flow (`https://www.mydatascope.com/webhooks`),
  so there is no request/response shape to declare a stream or write action against; this is the
  full extent of DataScope's "webhooks" surface, not a scoping choice.
- No DataScope resource in this bundle has a documented delete endpoint (locations, lists, task
  assignments, and tickets are all create/update-only in the real API); this bundle therefore
  declares no `delete_*` write action anywhere, matching the real surface exactly.
