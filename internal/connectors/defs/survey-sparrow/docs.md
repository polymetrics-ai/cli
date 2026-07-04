# Overview

SurveySparrow is a Tier-1 declarative-HTTP connector reading and managing SurveySparrow surveys,
contacts, responses, questions, channels, contact lists/properties, reminders, reputation
platforms/reviews, survey folders, ticket fields, tickets, teams, roles, variables, webhooks,
users, templates, email themes, and expressions through the SurveySparrow API v3
(`https://api.surveysparrow.com/v3/...`). This bundle was Pass-B full-surface expanded against the
real documented API: `https://developers.surveysparrow.com/rest-apis/` is a Docusaurus +
`docusaurus-openapi-docs` site whose category and per-endpoint pages are server-rendered (unlike
several other migrated connectors' JS-only doc SPAs) — all 113 documented method+path endpoints
were crawled directly from their individual reference pages (method badge + path +
request/response schema, e.g. `https://developers.surveysparrow.com/rest-apis/get-v-3-surveys-id`)
— see `api_surface.json` for the per-endpoint disposition. It originally targeted capability parity
with `internal/connectors/survey-sparrow` (the hand-written `surveysparrow` package it migrates);
the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a SurveySparrow access token via the `access_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <access_token>`) and is never logged, matching legacy's
`connsdk.Bearer(token)` (`survey_sparrow.go:119`) and the real API's documented Bearer auth scheme.
`base_url` defaults to `https://api.surveysparrow.com/v3` and may be overridden for tests/proxies.

## Streams notes

22 GET list streams, all sharing the real API's pagination shape (`page`/`limit` query params,
`data` array, `has_next_page` boolean in the response — this bundle's `page_number` paginator relies
only on the short-page stop signal, matching legacy exactly; `has_next_page` is not consulted by
either side).

- `surveys`, `contacts`, `responses`, `questions` — the original 4 legacy-parity streams. Their
  schemas intentionally retain only the fields the legacy `copyRecord(...)` mappers emitted:
  `id`/`name`/`survey_type`, `id`/`email`/`name`, `id`/`completed_time`/`survey_id`, and
  `id`/`question`/`survey_id`. The real API returns additional fields, but widening these four
  schemas would emit fields legacy always dropped.
- `channels`, `contact_lists`, `contact_properties`, `reminders`, `reputation_platforms`,
  `reputation_app_platforms`, `reputation_reviews`, `survey_folders`, `ticket_fields`, `tickets`,
  `teams`, `roles`, `variables`, `webhooks`, `users`, `templates`, `email_themes`, `expressions` —
  18 new Pass-B streams covering the rest of the API's clean top-level catalog resources.

**`questions` path CORRECTED (real correctness bug found and fixed, not merely a deviation)**: the
pre-Pass-B bundle (and the legacy Go connector it was ported from) declared `questions`' path as
`/surveys/{{ config.survey_id }}/questions` — a nested, survey-scoped REST path. **The real
documented API has no such endpoint.** The actual endpoint is `GET /v3/questions` with `survey_id`
as a **required query parameter**, confirmed directly from
`https://developers.surveysparrow.com/rest-apis/get-v-3-questions`'s query-parameters section
(`survey_id number required — Id of Survey`, alongside `page`/`limit`/`tag_name`/
`language_label`). This means the legacy connector's `questions` stream has likely never worked
against the live SurveySparrow API (a request to `/surveys/{id}/questions` would 404, since that
path was never real). This bundle now declares `"path": "/questions"` with `"query": {"survey_id":
"{{ config.survey_id }}"}` — a plain (non-optional-object) query template, so an absent
`config.survey_id` still hard-errors exactly as the old path-template did, preserving the
required-for-this-stream behavior; only the WIRE SHAPE (query param vs. path segment) changed, to
match the real API instead of a never-real assumed one.

`responses` declares `incremental.cursor_field: completed_time`, matching legacy's own
`CursorFields: []string{"completed_time"}` declaration and the real API's documented
`completed_time` field (a `date` type in the OpenAPI schema); neither this bundle nor legacy sends a
server-side lower-bound filter or performs client-side filtering for this stream (legacy's `Read`
performs no incremental filtering at all, and the real `/responses` endpoint documents no
`updated_after`-style filter param) — this bundle matches that exactly (no `request_param`/
`client_filtered` declared). No other stream declares a cursor field, matching legacy and the
absence of any documented equivalent filter elsewhere.

**Per-resource `limit` maximums confirmed to differ from legacy's single global bound.** Legacy
declared one `maxPageSize = 500` constant shared by every stream. The real API documents a DIFFERENT
per-resource maximum for the `limit` query parameter on nearly every list endpoint (confirmed by
inspecting each endpoint's query-parameters section):

| Resource | real `limit` max | this bundle's `page_size` |
|---|---|---|
| `contacts` | 50 | 50 (stream-level override) |
| `users` | 50 | 50 (stream-level override) |
| `reputation_app_platforms` | 50 | 50 (stream-level override) |
| `reputation_reviews` | 50 | 50 (stream-level override) |
| `audit_logs` | 500 | not implemented (see Known limits) |
| `responses`/`targets` | 200 | 100 (base default; within range) |
| everything else confirmed | 100 | 100 (base default) |

`streams.json`'s `base.pagination.page_size: 100` is the safe default for every resource whose real
max is ≥ 100; the four resources whose real max is 50 declare a stream-level `pagination` override
(`page_size: 50`) so a full-page read never sends a `limit` value the real API would reject/clamp.
This is a genuine correctness fix, not merely documentation: legacy's `page_size` config bound
(1-500) would have permitted an operator to configure e.g. `page_size=200` for `contacts`, which the
real API caps at 50 — this bundle's per-stream `page_size` is a fixed bundle value (matching the
"not runtime-configurable" limitation already true of the pre-Pass-B bundle, see Known limits) so
this specific over-limit risk cannot occur here regardless.

## Write actions & risks

33 write actions across 12 resources (27 named actions; some act on distinct real endpoints under
the same conceptual CRUD verb), every one a plain single-record JSON-body CRUD mutation the engine's
declarative dialect expresses directly (`body_type: json`, `path_fields` naming the record's id
field for update/delete):

- **Surveys**: `create_survey`, `update_survey` (PATCH partial update; no delete endpoint exists in
  the real API for surveys at all).
- **Contacts**: `create_contact`, `update_contact` (PUT full-replace), `delete_contact`.
- **Questions**: `create_question`, `update_question` (path param is `question_id`, not `id`),
  `delete_question`.
- **Contact lists**: `create_contact_list`, `update_contact_list`, `delete_contact_list`.
- **Contact properties**: `create_contact_property`, `update_contact_property`,
  `delete_contact_property`.
- **Survey folders**: `create_survey_folder`, `update_survey_folder`, `delete_survey_folder`.
- **Teams**: `create_team` only (no update/delete endpoints exist in the real API for teams).
- **Tickets**: `create_ticket`, `update_ticket`, `delete_ticket`.
- **Webhooks**: `create_webhook`, `update_webhook`, `delete_webhook`.
- **Users**: `create_user`, `update_user`, `delete_user`.
- **Reminders**: `create_reminder`, `delete_reminder` (no update endpoint exists in the real API for
  reminders).
- **Variables**: `create_variable`, `delete_variable` (path param is `variable_id`, not `id`; no
  update endpoint exists in the real API for variables).
- **Channels**: `create_channel`, `delete_channel` (update excluded — see Known limits).

All 33 mutations are flagged `risk: "external mutation; approval required"` (or a stronger
irreversibility/live-credential note for deletes and user creation); `capabilities.write` is now
`true`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`survey_sparrow.go`'s `pageSize`/`maxPages`, bounded 1-500 / a non-negative integer or
  `all`/`unlimited`). The engine's `page_number` paginator has no config-driven page-size or
  max-pages knob (`PaginationSpec.PageSize`/`MaxPages` are static bundle JSON, never templated), so
  this bundle uses a fixed bundle value per stream (see the per-resource `limit` maximum table
  above) and does not declare `page_size`/`max_pages` in `spec.json` at all (a declared-but-
  unwireable config key is worse than an absent one, per `docs/migration/conventions.md` F6).
  Pagination is unbounded by default (reads every page until a short page), matching legacy's own
  default of `maxPages=0` (unbounded) when `max_pages` is unset.
- **`audit_logs`, `targets`, and `languages` are NOT implemented as streams.** `audit_logs`' real
  response schema documents its records array (`list`, not `data`) as a bare `string[]` rather than
  a structured object type — confirmed on both the list and single-record GET reference pages —
  fabricating a structured schema for a genuinely unconfirmed shape would be guessing, not
  documenting. `targets`' response nests a plain `string[]` of labels at `data.targets`, not
  structured per-record data. `languages`' reference page renders no field list at all. See
  `api_surface.json`'s `out_of_scope` entries for each.
- **No `update_channel`/`create_response` actions.** `PUT /channels/{id}`'s only documented body
  field re-parents the channel to a different `survey_id` — a broad-blast-radius action on
  already-collected response attribution, excluded as `destructive_admin`. Response creation is a
  multi-step stateful workflow (`POST /responses/new` → `PUT /responses/{id}/update` per-question →
  `PUT /responses/{id}/complete`), not a single-record CRUD shape; correctly modeling it needs a
  compound multi-request `WriteHook` (Tier-2), which this connector does not have and this pass may
  not add a new hook package for — excluded rather than mis-modeled as a single incomplete POST.
- **Batch/async endpoints are out of scope.** `contacts/batch`, `responses/batch`, `tickets/batch`,
  `variables/batch` (bulk multi-record creation) and their `.../status/{token}` job-polling
  companions are excluded — this dialect's write model is single-record CRUD, not
  submit-a-job-and-poll.
- **Translation, localization, Employee 360, NPS/metrics, and reports are out of scope** — see
  `api_surface.json`'s `out_of_scope` entries; none are top-level catalog CRUD resources this
  connector's stream/write model targets.
- **Ticket comments and survey sections are sub-resources of an already-covered parent, not their
  own stream/write pair** — see `api_surface.json`'s entries.
- Full API-surface disposition (every one of the 113 documented SurveySparrow API v3 method+path
  pairs) is recorded in `api_surface.json`.
