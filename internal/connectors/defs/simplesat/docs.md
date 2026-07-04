# Overview

Simplesat reads and writes surveys, answers, questions, customers, and responses (including nested
ticket data) through the Simplesat v1 API (`https://api.simplesat.io/api/v1/...`). This is a Pass B
full-surface migration onto v1 — legacy (`internal/connectors/simplesat`) targeted the deprecated
v0 API (bare `answers/`, `surveys/`, `questions/`, `customers/`, `tickets/` paths, a uniform
`results` envelope, no write capability at all); this bundle migrates onto v1 throughout rather than
porting v0 verbatim, since v0 is documented by Simplesat itself as deprecated. See `api_surface.json`
for the full v0-vs-v1 reasoning.

## Auth setup

Provide a Simplesat API token via the `api_key` secret; it is sent as the `X-Simplesat-Token`
header (`api_key_header` auth mode). `base_url` defaults to `https://api.simplesat.io/api/v1` and
may be overridden for tests/proxies. `api_key` is never logged.

## Streams notes

All five streams (`answers`, `surveys`, `questions`, `customers`, `responses`) declare
`"projection": "passthrough"` — v1's response envelope varies per resource (a `next`/`previous`/
`count` + a per-resource plural key, e.g. `{"answers": [...], "next": ..., "count": ...}`) and this
bundle preserves every field verbatim rather than gating on a fixed schema-declared field set;
`schemas/*.json` stay a documentation surface of the well-known fields.

- `answers`: `POST /answers/search` (v1 exposes answer listing as a POST search endpoint, not a
  GET list — no filter body is sent, so it lists every answer), records at `answers`.
- `surveys`: `GET /surveys`, records at `surveys`.
- `questions`: `GET /questions`, records at `questions`.
- `customers`: `GET /customers`, records at `customers`. The only incremental stream:
  `incremental.cursor_field: modified` + `request_param: modified_after` sends v1's own
  `modified_after` server-side filter when a lower bound resolves (state cursor or the optional
  `created_after` config passthrough for the FIRST sync's lower bound — see below); `created_after`
  is a separate, always-available config passthrough filter independent of incremental state.
- `responses`: `POST /responses/search` (no filter body — lists every response), records at
  `responses`, including each response's nested `ticket`/`customer`/`team_members`/`answers`
  sub-objects verbatim. This supersedes legacy's `tickets` stream concept: v1 has no standalone
  ticket list endpoint at all (a v1 "ticket" only exists as a nested field of a response) —
  `schemas/responses.json`'s `ticket` property carries the exact same ticket data
  (`id`/`external_id`/`subject`/`custom_attributes`) legacy's `tickets` stream exposed, just nested
  rather than top-level.

Pagination is `next_url` (`next_url_path: "next"`) for every stream — v1's list envelope carries an
absolute `next` URL (or `null` on the last page), matching a `next_url`-shaped API exactly; per
conventions.md §4's sanctioned exception, `next_url` streams ship a single-page conformance fixture
(the replay server's own absolute URL is unknown until runtime) — `pagination_terminates` exercises
`answers` (see `fixtures/streams/answers/page_1.json`, whose `next: null` proves natural
termination without a second page).

`page_size` defaults to `100` and is always sent; `created_after`/`modified_after` are optional
passthrough/incremental query params, omitted entirely when unset (`omit_when_absent: true`).

## Write actions & risks

Pass B adds 6 write actions covering v1's create/update surface:

- `create_or_update_customer` (`kind: upsert`, `POST /customers`) — creates a new customer or
  updates the existing one matched by `external_id`/`email`; low-risk, no approval required.
- `update_customer` (`kind: update`, `PUT /customers/{{ record.id }}`) — mutates an existing
  customer's profile fields by id; overwrites `tags`/`custom_attributes` wholesale with the
  submitted value.
- `create_or_update_team_member` (`kind: upsert`, `POST /team-members`) — creates a new team
  member or updates the existing one matched by `external_id`/`email`; low-risk.
- `update_answer` (`kind: update`, `PUT /answers/{{ record.id }}`) — mutates an existing survey
  answer's recorded `choice`/`comment`/follow-up fields; changes customer-submitted response data.
- `create_or_update_response` (`kind: upsert`, `POST /responses/create-or-update`) — creates a new
  survey response (or updates one matched by the API's own dedup rule) including its nested
  `answers`/`customer`/`ticket`/`team_members` sub-objects; commonly used to import/backfill
  historical survey data with an explicit `created` timestamp.
- `update_response` (`kind: update`, `PUT /responses/{{ record.id }}/update`) — mutates an
  existing response's `tags`/`answers`/`team_members` by id; overwrites the identified response's
  recorded data.
- `send_survey_email` (`kind: custom`, `POST /surveys/{{ record.survey_token }}/email`) — sends a
  **live survey invitation email to the named customer's real inbox**; each call generates one
  outbound email delivery, not a reversible data mutation. Review before enabling in a caller with
  untrusted input.

**Deliberately NOT implemented**: single-record detail GETs (`/answers/{id}`, `/customers/{id}`,
`/responses/{id}`) are excluded as `duplicate_of` their respective search/list streams, which
already list every record in full; `/customers/bulk` is excluded as `duplicate_of`
`create_or_update_customer` (the engine's write dialect has no batch/array-body primitive — see
`api_surface.json`); team members have no list endpoint at all (`GET /team-members/{id}` is the
only read, with no way to enumerate ids), so no `team_members` read stream exists even though the
write action does.

## Known limits

- **`page_size`'s upper bound is not enforced by this bundle.** v1's documented range (1-250 for
  `customers`, provider default elsewhere) is not validated client-side; the engine's declarative
  query dialect has no numeric-range validation primitive. An out-of-range `page_size` is sent to
  the API as-is; the API itself is the ultimate arbiter of an invalid value.
- **No `team_members` read stream** — see Write actions & risks above; `create_or_update_team_member`
  exists as a write-only capability with no corresponding read.
- Legacy's `Check` targeted the deprecated v0 `answers/` endpoint; this bundle's `check` targets v1
  `GET /surveys` (an arbitrary lightweight v1 endpoint proving auth/connectivity — v1 has no
  dedicated health-check endpoint of its own).
