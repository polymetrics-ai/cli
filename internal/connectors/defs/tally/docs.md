# Overview

Tally (`https://tally.so`) is a form builder — this connector talks to the **Tally Forms API**
(`https://api.tally.so`), not any Tally accounting/finance product of the same name. It is a
greenfield build (no legacy Go connector exists for this name) authored directly from Tally's
published OpenAPI spec (`https://developers.tally.so/api-reference/openapi.json`) and its
human-readable reference (`https://developers.tally.so/api-reference/introduction`). It reads
forms, form-scoped submissions (fanned out over every form id), webhooks, and workspaces, and
writes the documented form/webhook/workspace mutations and submission deletion.

## Auth setup

Provide a Tally personal access token via the `api_key` secret; it is sent as
`Authorization: Bearer <api_key>` (`streams.json` `base.auth`'s `bearer` mode). Never logged.
Generate a token from Tally's account settings. `base_url` defaults to `https://api.tally.so` and
may be overridden for tests/proxies.

## Streams notes

- **`forms`** (`GET /forms`) — top-level paginated collection; records live at the `items` key.
  Not incremental: Tally's list-forms endpoint has no time-filter query parameter, only
  `page`/`limit`/`workspaceIds`.
- **`workspaces`** (`GET /workspaces`) — top-level paginated collection; records live at the
  `items` key. Not incremental (no documented time-filter parameter).
- **`webhooks`** (`GET /webhooks`) — account-wide paginated collection (spans every accessible
  form/workspace per Tally's own docs, not form-scoped); records live at the `webhooks` key. Not
  incremental (no documented time-filter parameter).
- **`submissions`** (`GET /forms/{formId}/submissions`) — form-scoped: this bundle uses the
  engine's `fan_out` dialect (`conventions.md` §3) to first list every form id (`ids_from.request`
  paginates `GET /forms` to exhaustion via the SAME base pagination spec every other stream uses,
  extracting `id` off each record at `items`), then repeats the full paginated submissions read
  once per form id, threading the id into the path as `{{ fanout.id }}` and stamping it onto every
  emitted record's `form_id` field. Records live at the `submissions` key of each page. Incremental
  via `submittedAt`: `incremental.request_param: startDate` sends the resolved lower bound
  (state cursor, falling back to the `start_date` config key) as the documented `startDate` query
  filter (ISO 8601/RFC3339, per Tally's docs) on every sub-sequence's requests; a fresh full sync
  (no state, no `start_date`) omits the parameter entirely. A `computed_fields` rename copies the
  raw `submittedAt` value into `submitted_at` (the schema's declared `x-cursor-field`) so the
  cursor-field name follows this bundle's snake_case convention while the raw camelCase field is
  also preserved verbatim for parity with the documented wire shape. **`conformance`'s dynamic
  `cursor_advances` check carries a `skip_dynamic` marker on this stream**: that check's generic
  re-read harness has no fan_out awareness (it cannot distinguish the preliminary `/forms`
  id-listing request from the per-form submissions sub-sequence against a single always-empty
  capture server, so the sub-sequence request that would carry `startDate` is never issued in that
  harness). The real behavior is proven instead by
  `internal/connectors/paritytest/tally/parity_test.go`
  (`TestParityTally_SubmissionsFanOutSendsStartDateOnResumedSync` and
  `TestParityTally_SubmissionsFreshSyncOmitsStartDate`), which drives a real two-request server and
  asserts both the resumed-sync and fresh-sync cases directly.

Pagination is uniform `page_number` (`page_param: page`, `size_param: limit`, `start_page: 1`,
`page_size: 100`) across every stream, exactly matching Tally's documented `page`/`limit` query
parameters and 1-based page numbering (`default: 1` in the OpenAPI spec) — the engine's short-page
stop rule (a page returning fewer than `page_size` records ends pagination) is the correct
termination signal for this shape; Tally's own `hasMore` response field is a redundant, not
consulted, restatement of the same fact for API consumers who prefer not to count. `limit`'s
documented max is 500 for forms/submissions and 100 for webhooks; `page_size: 100` (this bundle's
uniform value) is within both documented maxima. `page_size` is also exposed as a `spec.json`
config override (default `"100"`) wired into each stream's `query.limit` — the base pagination
block's own `page_size: 100` still governs the client-side short-page stop threshold and is
independent of the config override, matching the stripe/akeneo precedent (`conventions.md` §3).

## Write actions & risks

- **`create_webhook`** / **`update_webhook`** / **`delete_webhook`** (`POST`/`PATCH`/`DELETE
  /webhooks[/{id}]`) — the documented webhook mutation surface this build's mandate calls out by
  name. `eventTypes` currently has exactly one documented value (`FORM_RESPONSE`). Deleting a
  form's last webhook also marks its webhook integration deleted (Tally's own documented behavior,
  not a side effect this bundle introduces). `delete_webhook` is `confirm: destructive`.
- **`create_form`** / **`update_form`** / **`delete_form`** — full form lifecycle mutations.
  `delete_form` moves a form to the trash (Tally's documented soft-delete semantics, not a
  permanent purge) and is `confirm: destructive`. `update_form` requires at least the form `id`
  plus one other field (`minProperties: 2` on `record_schema`, since a body with zero changed
  fields is a meaningless PATCH).
- **`delete_submission`** (`DELETE /forms/{formId}/submissions/{submissionId}`) — permanently
  removes one respondent's submission and its answers; `confirm: destructive`. Requires both
  `form_id` and `id` since the endpoint path is form-scoped.
- **`create_workspace`** (`POST /workspaces`) — Tally's docs state this requires a Pro
  subscription on the account; a Free-tier account will see this fail with a real API 403, not a
  bundle-side check (the dialect has no plan-tier precondition mechanism, nor should it fake one).

## Known limits

- **Webhook event delivery-log and retry endpoints are not modeled**
  (`GET /webhooks/{webhookId}/events`, `POST /webhooks/{webhookId}/events/{eventId}`) — these are
  delivery-log inspection/retry actions, not syncable data objects or reverse-ETL mutations; see
  `api_surface.json`.
- **Analytics endpoints are not modeled** (`metrics`/`visits`/`submissions`/`dimensions`/
  `drop-off` under `/forms/{formId}/analytics/*`) — each returns an aggregate snapshot for a
  parameterized time window, not a list of syncable records with a stable primary key.
- **Organization/user/invite administration is not modeled** — out of scope for a forms/
  submissions/webhooks/workspaces-focused connector; these mutate account membership, not form
  data.
- **Workspace update/delete and folder management are not modeled** — `PATCH`/`DELETE
  /workspaces/{workspaceId}` and the entire `/workspaces/{workspaceId}/folders` sub-resource are
  excluded (`requires_elevated_scope`/`destructive_admin`/`out_of_scope` respectively in
  `api_surface.json`); `create_workspace` is the only workspace write this build ships.
- **`questions`/`blocks` form-definition sub-resources are not modeled** as separate streams —
  they describe a form's structure (form-builder metadata), not response data; a future
  capability-expansion pass could add them as read-only streams if a consumer needs form structure
  alongside submission data.
- **No live-API credentials were available while authoring this bundle.** Fixtures were derived
  directly from Tally's published OpenAPI examples/schemas (recorded-real-shape per
  `conventions.md` §4) with synthetic values substituted for every real identifier/timestamp; they
  have not been validated against a live Tally account. `mode: fixture` short-circuits network
  access for credential-free conformance, matching every other bundle's convention.
