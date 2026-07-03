# Overview

Fillout is a form builder; its REST API exposes the account's forms, question definitions, and
submissions. This is a **partial** Tier-1 migration (`docs/migration/quarantine.json`): only the
`forms` stream is declarative here. `questions` and `submissions` remain on the legacy
`internal/connectors/fillout` implementation because they require a per-form sub-resource fan-out
whose id-resolution mode is itself runtime-conditional — see Known limits. The legacy package stays
registered and unchanged until wave6's registry flip; this bundle exists so `forms` can be exercised
through the declarative engine in the meantime.

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

None. Fillout submissions are inbound user data, not a safe reverse-ETL write target;
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy exactly.

## Known limits

- **`questions` and `submissions` are NOT implemented here (quarantined, `ENGINE_GAP`).** Legacy's
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
- `api_surface.json` marks `/forms/{id}` and `/forms/{id}/submissions` `out_of_scope` pending the
  above gap; see `docs/migration/quarantine.json`'s `fillout` entry for the blocker record.
