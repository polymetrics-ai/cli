# Overview

Timely reads Timely users, projects, clients, calendar/time events, time entries (`hours`), tags
(`labels`), and teams through the Timely API v1.1
(`GET https://api.timelyapp.com/1.1/<account_id>/<resource>`). This bundle originated as a
capability-parity migration from `internal/connectors/timely` (the hand-written connector it
migrates; the legacy package stays registered and unchanged until wave6's registry flip) and was
then expanded to Timely's practical time-tracking-core read surface (Pass B), per Timely's own
published OpenAPI 3.1 document (served inline on https://developer.timely.com/). Every mutation
endpoint in that document (client/project/user/team/label/time-entry create+update) is out of scope
— see "Write actions & risks" below.

## Auth setup

Provide a Timely OAuth access token via the `bearer_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <bearer_token>`), matching legacy's `connsdk.Bearer(token)`
(`timely.go:123`) exactly, and is never logged. `account_id` is required and prefixes every
stream's path (`<account_id>/<resource>`), matching legacy's `accountPath` helper
(`timely.go:166-172`). `base_url` defaults to `https://api.timelyapp.com/1.1`, matching legacy's
`defaultBaseURL` fallback.

## Streams notes

The original 4 streams (`users`, `projects`, `clients`, `events`) read the full JSON array response
(`records.path: "."`) with no pagination, matching legacy's single unpaginated `Do` request per
stream (`timely.go:91`, `RecordsAt(resp.Body, ".")`) — unchanged by the Pass B expansion below.
Primary key `["id"]` on every stream.

**`events` targets an endpoint no longer present in Timely's own current OpenAPI document** —
`/1.1/{account_id}/events` (singular) does not appear anywhere in the 106-path document served at
https://developer.timely.com/; the documented modern equivalent is `/1.1/{account_id}/hours` (time
entries with linked work items). Since this bundle must never change accepted-input behavior for
an existing stream (conventions.md §5's meta-rule), `events` is left completely untouched — it
still targets `/events` exactly as legacy did, on the working assumption (consistent with every
available secondary source; no deprecation notice for this path was found) that Timely kept it as
a live backward-compatible alias. The modern, fully-documented `hours` resource is added as a
brand-new, ADDITIONAL stream (see below) rather than replacing `events`, so a caller already
depending on the `events` stream's exact shape sees no change at all.

**`hours`** (`GET /{account_id}/hours`) is the new Pass B stream for Timely's documented time-entry
resource. It supports real `page`/`per_page` pagination (`page_number`, `page_size: 100`, matching
Timely's own documented `per_page` cap guidance) — the ONLY stream in this bundle that paginates.
`project_id`/`user_id` are extracted from the raw response's nested `project.id`/`user.id` objects
via typed `computed_fields` bare-reference extraction (conventions.md §3's "typed extraction" rule
— these copy the raw integer id, not a stringified one), matching the naming convention the
legacy-parity `events` stream's own flat `project_id`/`user_id` fields already use. No `incremental`
block is declared: the real `since`/`upto` query params filter time entries by their **calendar day**
(`day`), not by `updated_at`/last-modified — there is no server-side "updated since X" filter to
back a genuine incremental cursor, so declaring `x-cursor-field` without a matching `incremental`
block would misrepresent this as change-tracking capability the API doesn't actually offer
(conventions.md §8 rule 2's incremental truth table).

**`labels`** (`GET /{account_id}/labels`) and **`teams`** (`GET /{account_id}/teams`) are new Pass B
streams for Timely's documented tag and team resources; both are small, unpaginated per-account
lists (no pagination block declared, matching every other non-`hours` stream in this bundle).

`events` additionally sends a `since` query param sourced from the `start_date` config value via
the opt-in optional-query dialect (`{"template": "{{ config.start_date }}", "omit_when_absent":
true}`) — present only when `start_date` is configured, omitted entirely otherwise, matching
legacy's own stream-specific gating exactly (`timely.go:86-90`: `since` is only ever set for the
`events` stream, and only when `start_date` is non-empty). This is a plain optional config
passthrough, not a true `incremental` block: legacy tracks no persisted cursor and applies no
client-side re-filtering — it is a one-shot "start from this timestamp" hint the Timely API
itself interprets, so this bundle deliberately does not declare an `incremental` block (which
would imply cursor-based state tracking legacy never had).

All 4 streams declare `"projection": "passthrough"` (conventions.md §8 rule 1). Legacy's `Read`
extracts records via `connsdk.RecordsAt(resp.Body, spec.recordsPath)` and does
`emit(connectors.Record(rec))` on each one with no field-building step at all
(`timely.go:95-103`); `streamSpecs[...].fields` (e.g. `id`/`name`/`email`/`created_at`/
`updated_at` for `users`) is Catalog-only decoration (`timely.go:126-146`, consumed solely by
`Catalog()`'s `connectors.Stream` construction), never applied to the emitted record itself.
Default `"schema"` projection mode would silently drop every real Timely field not named in each
stream's declared schema properties (the live API returns considerably more per-object detail
than legacy's catalog list names, e.g. `users`' `external_id`/`role`/`holiday_calendar`,
`projects`' `budget`/`color`/`active`/`billable`, `events`' `note`/`billed`/`from`/`to`/`user_ids`),
an undocumented silent data-shape change relative to legacy's raw passthrough. Each schema still
declares the fields legacy's own catalog names (for `x-primary-key` typing and
`records_match_schema` coverage), but passthrough mode means any other real field Timely returns
still survives unfiltered, matching legacy exactly.

## Write actions & risks

None. Timely is read-only (`capabilities.write: false`, no `writes.json`). This is an `ENGINE_GAP`,
not a scope choice: EVERY mutation endpoint in Timely's real, documented API (client/project/
user/team/label/time-entry create+update) requires the request body to be wrapped in a
resource-named single-key JSON envelope — e.g. `POST /clients` expects `{"client": {"name": ...}}`,
`POST /hours` expects `{"event": {...}}`, `POST /teams` expects `{"team": {...}}`, and so on,
identically across all 6 mutation resource families (well past the ≥3-occurrence recurrence
threshold conventions.md §6 uses to trigger an engine mini-wave). The engine's declarative write
dialect (`body_type: json`/`form`/`none`) can only construct a FLAT top-level JSON object from a
record's fields; it has no primitive for nesting that object under a named envelope key. See
`api_surface.json`'s per-endpoint `ENGINE_GAP` notes for the full endpoint list this blocks.

## Known limits

- **No writes are implemented (`ENGINE_GAP`).** See "Write actions & risks" above —
  every Timely mutation endpoint needs a nested single-key body envelope the engine's write
  dialect cannot express. This is the correct outcome per conventions.md §6: a real, recurring,
  correctness-relevant engine gap is documented as a blocker, not worked around with an invented
  Go hook whose only job would be wrapping a body (not a sanctioned Tier-2 trigger).
- **`events`'s `since` is a one-shot config hint, not a stateful incremental cursor.** Legacy
  never persists or advances a cursor value for this stream — every sync either passes the
  configured `start_date` verbatim or omits `since` entirely. This bundle reproduces that exact
  behavior; it is not eligible for `incremental_append[_deduped]` sync modes since no
  `incremental` block is declared (conventions.md §2, "Sync-mode derivation — never declared").
- **`events` targets a path absent from Timely's current published OpenAPI document.** See
  "Streams notes" above — kept unchanged for parity; `hours` is the documented modern equivalent
  and is the stream to prefer for any new integration.
- **`hours` has no incremental/cursor support.** Timely's `since`/`upto` query params filter by
  calendar day, not by last-modified time; there is no server-side "updated since X" filter to back
  a real incremental cursor for this stream (see "Streams notes" above).
- **Billing/invoicing, forecasting, AI timesheet-proposal, third-party app/workspace integration,
  webhooks, and account-administration resources are out of scope** — outside this connector's
  time-tracking-core domain (see `api_surface.json`'s resource-family exclusions for the full
  list and reasoning).
