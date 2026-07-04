# Overview

Fillout is a form builder; its REST API exposes the account's forms, question definitions,
submissions, and webhook management. This is a **partial** Tier-1 migration
(`docs/migration/quarantine.json`) for READS: only the `forms` stream is declarative here.
`questions` and the submissions LIST remain on the legacy `internal/connectors/fillout`
implementation because they require a per-form sub-resource fan-out whose id-resolution mode is
itself runtime-conditional — see Known limits. The legacy package stays registered and unchanged
until wave6's registry flip; this bundle exists so `forms` can be exercised through the declarative
engine in the meantime. **Pass B full-surface expansion** (this revision) adds the full practical
WRITE surface, which has no dependency on the blocked fan_out gap: `create_webhook`/`remove_webhook`
(flat-body POSTs) and `delete_submission_by_id` (a plain path-parameterized delete targeting one
already-known submission, independent of the still-blocked submissions LIST). `capabilities.write`
is now `true`.

## Auth setup

Provide a Fillout API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged. `base_url` defaults to
`https://api.fillout.com/v1/api` and may be overridden for tests or proxies.

## Streams notes

`forms` (`GET /forms`) returns a top-level JSON array (no envelope) — modeled with
`records.path: ""` (body root). Each raw item carries its id as `formId` (matching legacy's own
`firstString(item, "formId", "id")` normalization): `computed_fields` extracts
`"id": "{{ record.formId }}"` as a bare single-reference template, so the raw typed value is copied
straight through; on the rare raw shape that carries `id` directly instead of `formId`, schema
projection's ordinary exact-key-match copy of `id` acts as the fallback (the `computed_fields` entry
silently no-ops when `record.formId` is absent, per the engine's per-record-absent-path tolerance),
reproducing legacy's two-key fallback without any conditional templating. No pagination is declared
(legacy does not paginate `/forms` either — the endpoint returns every form in one response).

## Write actions & risks

Three write actions, all Pass B additions (legacy has no write path at all — this bundle is
strictly additive versus legacy's read-only implementation, not a parity port):

- **`create_webhook`** (`POST /webhook/create`, `body_type: json`) — registers a new outbound
  webhook: `formId` + `url`, both required. Response carries the new webhook's integer `id`
  (`fixtures/writes/create_webhook.json` declares the `response` block so a future `WriteHook`
  follow-up, if ever needed, could read it back — none is needed today, this is a single-request
  write). Risk: begins delivering live submission events to an external URL immediately.
- **`remove_webhook`** (`POST /webhook/delete`, `kind: custom` since Fillout's own API uses a POST
  with a body rather than a path-parameterized `DELETE` — `body_type: json` carrying `webhookId`).
  `confirm: destructive`: irreversibly stops event delivery for that webhook.
- **`delete_submission_by_id`** (`DELETE /forms/{form_id}/submissions/{submission_id}`, `kind:
  delete`, `path_fields: ["form_id", "submission_id"]`, `missing_ok_status: [404]` for idempotent
  re-delete). Permanently removes one form response; irreversible.

**Not modeled**: `POST /forms/{formId}/submissions` (`create_submissions`) — its request body is a
`submissions` array of NESTED objects (each carrying its own `questions[]`/`urlParameters[]`/
`scheduling[]`/`payments[]` sub-arrays), not the flat "every record field except `path_fields`"
shape the write dialect's default body construction expresses; see Known limits.

## Known limits

- **`questions` and the submissions LIST are NOT implemented here (quarantined, `ENGINE_GAP`).** Legacy's
  `resolveFormIDs` (`internal/connectors/fillout/fillout.go:259`) uses the configured
  comma-separated `form_id` when set, and otherwise discovers every form id via `GET /forms` — a
  **runtime conditional fallback between the two id sources**. The engine's `fan_out.ids_from`
  (`internal/connectors/engine/bundle.go`'s `FanOutSpec`) requires declaring **exactly one** of
  `config_key` or `request` at load time (`resolveFanOutIDs`, `internal/connectors/engine/read.go:122-138`,
  hard-errors if both or neither are set) — there is no combinator expressing "use config_key when
  present, else fall back to request". Statically picking `request` (always discover via `GET
  /forms`) would silently return every form's data for a caller who set `form_id` to narrow the
  scope — a legacy-accepted input (`form_id` set) that would now emit MORE records than legacy.
  Statically picking `config_key` would silently emit ZERO records for a caller who left `form_id`
  unset (legacy's other accepted, and most common, input) — a legacy-accepted input that would now
  emit FEWER records than legacy. Both choices change emitted-record data for at least one
  legacy-accepted input, which `docs/migration/conventions.md` section 5's meta-rule forbids as a
  silent deviation; this is filed as a distinct `ENGINE_GAP` (narrower than the original quarantine
  reason, which predates the `fan_out` dialect addition entirely) rather than worked around.
  `questions` (`GET /forms/{id}`, extracting the nested `questions` array) and `submissions`
  (`GET /forms/{id}/submissions`, offset/limit-paginated, incremental on `submissionTime` via an
  `afterDate` query param) are otherwise fully expressible with `fan_out`'s `path_var` `into` mode
  plus ordinary `offset_limit` pagination and `param_format: rfc3339` incremental — only the
  id-source fallback blocks them.
- `api_surface.json` marks `/forms/{id}` and `/forms/{id}/submissions` (LIST) `out_of_scope` pending
  the above gap; see `docs/migration/quarantine.json`'s `fillout` entry for the blocker record.
- **`GET /forms/{formId}/submissions/{submissionId}` (single-submission detail) is NOT
  implemented** — marked `duplicate_of` in `api_surface.json`: it is a narrower view of the same
  still-blocked submissions resource, and implementing only the single-id detail lookup without the
  list would not restore meaningful read parity with legacy's `submissions` stream.
- **`POST /forms/{formId}/submissions` (`create_submissions`) is NOT modeled as a write** — its
  request body is `{"submissions": [{...nested question/scheduling/payment arrays...}]}`, a
  NESTED-array shape; the write dialect's default body construction (`body_type: json` = every
  record field except `path_fields`) only expresses a flat top-level object, with no mechanism to
  declare a nested array-of-objects body from a flat input record. Expressing this correctly would
  need either a bespoke nested-body templating addition to the dialect, or a `WriteHook` (a new
  Tier-2 hook package this task's caps forbid creating for fillout, which has none today) — filed as
  `out_of_scope` in `api_surface.json` rather than silently approximated with a flattened body that
  would not match Fillout's real wire shape.
