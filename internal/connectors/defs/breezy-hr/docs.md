# Overview

Breezy HR is a wave2 fan-out declarative-HTTP migration, expanded in Pass B to the connector's
full documented v3 REST API surface (research this pass extracted the complete, real OpenAPI 3.1
spec embedded in the ReadMe-hosted reference page markup — 46 distinct operations across `/v3` —
rather than relying on prose summaries alone). It reads Breezy HR positions (job openings), hiring
pipelines, per-position candidates, departments, categories, custom attribute definitions,
questionnaires, and message templates for a configured company; it writes position create/update/
state-change and candidate create/update/pipeline-stage-move mutations, through the Breezy v3 REST
API. The legacy package (`internal/connectors/breezy-hr`) stays registered and unchanged until
wave6's registry flip. The `candidates` sub-resource fan-out read is expressed via the engine's
`fan_out` dialect (S4 engine mini-wave item 2) — the `ENGINE_GAP` that previously blocked this
stream is closed.

## Auth setup

Provide a Breezy HR raw API key via the `api_key` secret; it is sent verbatim as the `Authorization`
header value (`auth: api_key_header`, no `Bearer`/other prefix), and a `company_id` secret naming
the Breezy company to scope every request to. Both are never logged. The effective base URL is
`{{ config.base_url }}/company/{{ secrets.company_id }}` (`base_url` defaults to
`https://api.breezy.hr/v3`).

`company_id` is marked `x-secret: true` to match the legacy catalog's own field classification,
even though it is a path segment rather than a credential used for request signing/authentication.

Breezy's own documentation also exposes a separate, session-token-based `/signin`/`/signout`
email+password auth flow (a Breezy dashboard user's own login credentials, not an API key) — this
connector's raw-API-key model is a fundamentally different credential shape and does not use that
flow at all; see Known limits.

## Streams notes

- `positions` — `GET /positions`, records at the response root (`records.path: "."`; Breezy returns
  a bare top-level JSON array, not an enveloped object). Paginated via `pagination.type:
  page_number` (`page_param: page`, `size_param: limit`, `start_page: 1`, `page_size: 100`),
  matching legacy's `harvestPositions` exactly. `position_id` is computed from the raw API's `_id`
  field (`computed_fields`); `type` is computed from the raw nested `type.name` object field;
  `country_id`/`country_name` are computed from the raw nested `location.country.id`/
  `location.country.name` object fields.
- `pipelines` — `GET /pipelines`, records at the response root, unpaginated. `id` is computed from
  the raw API's `_id` field.
- `candidates` — a sub-resource fan-out over positions: `fan_out.ids_from.request` issues the SAME
  paginated `GET /positions` sequence the `positions` stream itself uses, then
  `into.path_var: "position_id"` threads each discovered position id into `/position/{{ fanout.id
  }}/candidates`, and `stamp_field: "position_id"` writes it onto every emitted candidate record
  after projection. `id` is computed from the raw API's `_id` field; `stage` is computed from the
  raw nested `stage.name` object field. `query: {"sort": "updated_date"}` matches legacy verbatim.
- `departments` **(new this pass)** — `GET /departments`, records at the response root (a bare
  array, matching every other simple-list Breezy endpoint), unpaginated. `id` computed from `_id`.
- `categories` **(new this pass)** — `GET /categories`, same shape as `departments`.
- `custom_attributes_candidate` **(new this pass)** — `GET /custom-attributes/candidate`, the
  candidate custom-FIELD-DEFINITION list (`id`/`name`/`secure`), records at the response root,
  unpaginated. The per-candidate custom field VALUES (`.../candidate/{id}/custom-fields`) are not
  modeled — see Known limits.
- `custom_attributes_position` **(new this pass)** — `GET /custom-attributes/position`, the
  position custom-field-DEFINITION list, same shape as its candidate counterpart.
- `questionnaires` **(new this pass)** — `GET /questionnaires`, records at the response root,
  unpaginated. `id` computed from `_id`.
- `templates` **(new this pass)** — `GET /templates` (message/email templates), same shape as
  `questionnaires`.

None of the 9 streams declare an incremental cursor: Breezy's public API exposes no updated-since
filter for any of them; all are full-refresh only.

## Write actions & risks

- `create_position` — `POST /positions`; creates a new job opening. If not left in draft state, may
  become publicly visible on the company's careers page and job boards depending on the configured
  `state` (defaults to `published` server-side per Breezy's own docs).
- `update_position` — `PUT /position/{position_id}`; mutates an existing position's title/
  description/location/department.
- `update_position_state` — `PUT /position/{position_id}/state`; changes a position's lifecycle
  state (`published`/`draft`/`closed`/`archived`) — the highest-visibility write action this
  connector exposes, since `published` makes the job publicly listed.
- `create_candidate` — `POST /position/{position_id}/candidates`; adds a new candidate to a
  position's pipeline. Low risk, additive.
- `update_candidate` — `PUT /position/{position_id}/candidate/{candidate_id}`; mutates an existing
  candidate's contact/profile fields.
- `move_candidate_stage` — `PUT /position/{position_id}/candidate/{candidate_id}/stage`; moves a
  candidate to a different pipeline stage WITHIN THE SAME position (`body_fields: ["stage_id"]`
  restricts the body to just the target stage id). Moving to a terminal stage (hired/disqualified)
  may trigger the position's configured stage actions (auto-emails, webhook notifications) if
  `stage_actions_enabled` was set on that position.

## Known limits

- **`candidates`'s per-position request is now paginated; legacy's was not.** (Unchanged from
  wave2.) Legacy's `readCandidates` issues exactly ONE unpaginated `GET /position/{id}/candidates`
  request per position; the engine's `fan_out` dialect reuses the SAME pagination spec for both the
  id-listing request AND every per-id child sub-sequence, so this bundle's `candidates` stream also
  (harmlessly, in the common case) paginates each per-position candidates read. Documented parity
  deviation (§5, ACCEPTABLE): for any position with 100 or fewer candidates, exactly one request is
  issued per position — byte-identical to legacy; only a position with MORE than 100 candidates
  diverges, and only by emitting additional true records legacy's own unpaginated read would have
  silently truncated.
- **Real, confirmed query-parameter naming differences vs. this bundle's existing (legacy-parity)
  shape are NOT changed here, per the parity meta-rule.** The Breezy v3 OpenAPI spec (extracted this
  pass) documents `/positions` as accepting NO pagination query params at all (only an optional
  `state` filter) and `/position/{id}/candidates` as using `page_size` (not `limit`) and
  `sort=updated|created` (not `sort=updated_date`) — all three differ from what this bundle (and
  legacy before it) actually sends. Since legacy's own `harvestPositions`/`readCandidates` already
  send `page`/`limit`/`sort=updated_date` unconditionally and have done so in production, and Breezy
  appears to silently ignore unrecognized query params (returning the same unpaginated/default-sort
  result either way — the same "harmless unused param" class as this connector's own fan_out
  deviation above), changing these param names now would be an unreviewed, out-of-band behavior
  change to already-shipped legacy-parity request shapes, not a Pass B surface-coverage fix. Flagged
  here as a genuine finding for a future dedicated review, not silently corrected.
- **The root-level `GET /companies` and `GET /company/{company_id}` endpoints are a genuine
  ENGINE_GAP, not merely out of scope.** Every stream in this bundle shares one company-scoped
  `base.url` (`{{ config.base_url }}/company/{{ secrets.company_id }}`); reaching the root-level
  `/companies` listing would require a stream path that escapes back above that prefix. The
  dialect's `InterpolatePath` applies `urlencode` by DEFAULT to every resolved `{{ }}` reference
  (conventions.md §3), so a stream path templating `{{ config.base_url }}` directly to reconstruct
  an absolute root-level URL would corrupt the scheme/host into a percent-encoded string — and a
  hardcoded absolute literal duplicating the DEFAULT `base_url` value would silently ignore a
  caller-configured `base_url` override, a real accepted-input divergence the parity meta-rule
  forbids. There is no per-stream base-URL-override primitive in this dialect today. Legacy stays
  authoritative for this capability (which it also never implemented, so no capability is actually
  lost relative to legacy) until the engine gains one.
- **Candidate/position sub-resource detail endpoints are not modeled** (assessments,
  background-checks, conversation threads, custom-field VALUES, documents, education/work-history
  sub-objects, questionnaire responses, resumes, scorecards, activity streams, team members) — each
  requires an already-known candidate_id/position_id with no incremental sync value beyond the
  `candidates`/`positions` streams' own records, and several return binary file content or are
  themselves communication actions (send an email, post a comment) rather than data reads/writes.
  See `api_surface.json` for the complete per-endpoint accounting.
- Full Breezy v3 API surface still excluded as genuinely out of scope: the session-token
  `/signin`+`/signout`+password-reset auth flow (a different credential model this connector
  doesn't use), single-user profile introspection, and candidate cross-position search (an
  on-demand lookup requiring an already-known email address, not an enumerable stream).
