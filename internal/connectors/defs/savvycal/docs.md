# Overview

SavvyCal is a Pass B full-surface-expansion declarative-HTTP migration. It reads SavvyCal events,
scheduling links, contacts, time zones, webhooks, and workflows, and writes scheduling-link and
webhook lifecycle mutations, through the SavvyCal API (`https://api.savvycal.com/v1/...`). This
bundle targets capability parity with `internal/connectors/savvycal` (the hand-written connector it
migrates) as a **superset**: legacy's original 3 read streams are preserved unchanged, and 4 new
read streams plus 9 write actions are added against the documented SavvyCal REST API surface
(`https://developers.savvycal.com/`, fetched 2026-07-03). The legacy package stays registered and
unchanged until wave6's registry flip.

## Auth setup

Provide a SavvyCal API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's `connsdk.Bearer(token)`
(`savvycal.go:178`). `base_url` defaults to `https://api.savvycal.com` and may be overridden for
tests/proxies.

## Streams notes

**Legacy-parity streams** (`events`, `links`, `contacts`) are unchanged from the prior wave: `GET`
against the SavvyCal list endpoint, records at `data`, primary key `["id"]`, `next_url` pagination
on `links.next`, `projection: passthrough` (legacy performs zero record shaping —
`savvycal.go:126`'s `emit(connectors.Record(rec))` passes the raw decoded record straight through).

**`contacts` reads `/v1/contacts`, an endpoint no longer present in SavvyCal's current public API
documentation** (`https://developers.savvycal.com/` enumerates Events, Scheduling Links, Current
User, Time Zones, Webhooks, and Workflows only — no Contacts category, and `GET /v1/contacts`
returns `404` against the live docs site's own route table check). This stream is preserved
unchanged for legacy parity (never remove functionality a prior wave already shipped) but is not
re-verified against current documentation; if SavvyCal has fully retired this endpoint, a live
account's real API call would 404 — this is a pre-existing legacy behavior, not something
introduced by this pass. Flagged here rather than silently dropped or silently re-endorsed.

**New streams added this pass** (all `GET`, `projection: passthrough` for the same reason as the
legacy-parity streams — none of these list endpoints are documented as returning a shape this
bundle's own code reshapes):

- `time_zones` (`GET /v1/time_zones`) — the complete list of IANA time zone definitions SavvyCal
  publishes (localized display name, abbreviation, UTC offset, DST flag); unpaginated
  (`pagination: none`), matching the docs' own description (no "paginated" language, unlike every
  other list endpoint here).
- `webhooks` (`GET /v1/webhooks`) — the authenticated user's registered webhook subscriptions;
  `next_url` pagination on `links.next`, matching the other list streams' documented convention.
- `workflows` (`GET /v1/workflows`) — marketing-automation-style workflows for every scope the
  authenticated user manages; `next_url` pagination.
- `workflow_rules` (`GET /v1/workflows/{workflow_id}/rules`) — a **`fan_out` stream**
  (conventions.md §3): there is no top-level endpoint to list rules across every workflow, so this
  stream fans out over every id returned by `workflows`'s own list request
  (`fan_out.ids_from.request`), issuing one unpaginated rules sub-request per workflow and stamping
  `workflow_id` onto every emitted rule record.

None of the 7 streams declare an `incremental` block, matching legacy's `Catalog` (no
`CursorFields`) and the fact that no SavvyCal list endpoint documents an updated-since filter
parameter.

Pagination for every `next_url` stream follows the SavvyCal-documented `links.next` body-path
convention (`savvycal.go:130`'s `firstStringAt(resp.Body, "links.next", "next")`): the engine's
`next_url` dialect supports only a single `next_url_path`, so this bundle declares `links.next` and
does not model legacy's secondary `next` top-level fallback path (never observed in SavvyCal's real
API responses). `page=1`/`per_page={{ config.page_size }}` (default `100`) is a static per-stream
`query`, re-sent on every page request; the engine's `next_url` paginator re-merges `stream.Query`
onto every absolute next-page URL, a benign wire-request-shape divergence from legacy's own
`path/query = next; query = nil` reset, since SavvyCal's own `links.next` URL already encodes the
correct pagination state server-side.

`metadata.json` declares no `rate_limit` — legacy's own SavvyCal package enforces no client-side
rate limiting either, so this bundle adds none, matching legacy's real (lack of) behavior.

## Write actions & risks

`capabilities.write` is now `true` (9 actions added; legacy shipped none):

- `create_personal_link` (`POST /v1/links`, JSON body) — creates a new scheduling link in the
  authenticated user's personal scope. Approval required.
- `create_scope_link` (`POST /v1/scopes/{scope_slug}/links`, JSON body) — creates a new scheduling
  link under a specific team or individual scope (`scope_slug` is a record field threaded into the
  path, e.g. `acme-inc` or `john`). Approval required.
- `update_link` (`PATCH /v1/links/{id}`, JSON body) — updates an existing scheduling link's
  name/slug/duration. Approval required.
- `delete_link` (`DELETE /v1/links/{id}`) — **destructive/irreversible**: permanently deletes a
  scheduling link. Approval required.
- `duplicate_link` (`POST /v1/links/{id}/duplicate`) — creates a copy of an existing scheduling
  link. Low-risk, no approval required (matching the create-action precedent elsewhere in this
  migration wave: a plain, non-destructive create-shaped action).
- `toggle_link` (`POST /v1/links/{id}/toggle`) — flips a scheduling link between active and
  disabled state, changing its public bookability. Approval required.
- `cancel_event` (`POST /v1/events/{id}/cancel`) — **destructive/irreversible**: cancels a
  scheduled event and notifies attendees. Approval required.
- `create_webhook` (`POST /v1/webhooks`, JSON body) — creates a new webhook subscription that
  will POST event notifications to an external URL. Approval required (an external URL supplied
  by write-record data will receive live account event data).
- `delete_webhook` (`DELETE /v1/webhooks/{id}`) — **destructive/irreversible**: permanently deletes
  a webhook subscription. Approval required.

## Known limits

- **`max_pages` is not modeled.** Legacy exposes a config-driven `max_pages` override
  (`savvycal.go:99`, default `100`) as a hard request-count cap. The engine's `next_url` paginator
  has no `MaxPages`-equivalent knob wired to a config value; pagination here is bounded only by the
  short/empty-`links.next` stop signal, matching SavvyCal's own real termination behavior. Not
  declared in `spec.json` at all (F6, REVIEW.md precedent: a declared-but-unwireable config key is
  worse than an absent one).
- **Legacy's fixture-mode-only marker field is not modeled** (`savvycal.go:162`'s synthetic
  `fixture: true` marker) — a legacy-only testing convenience superseded by the engine's own
  conformance/fixture-replay harness.
- **Event creation (`POST /v1/links/{link_id}/events`) is excluded.** Booking an event against a
  scheduling link's availability is an attendee + time-slot-selection flow, not a plain flat-field
  CRM-style record create; no documented flat request schema is suited to a reverse-ETL write
  action here.
- **The link-slots availability lookup (`GET /v1/links/{link_id}/slots`) is excluded.** Its real
  request requires a start/end availability-window query whose exact parameter names are rendered
  client-side by the docs site's OpenAPI viewer (not scrapeable via a static fetch); implementing a
  fan_out stream on a guessed query shape risks silently returning an always-empty or malformed
  result rather than an honest, verified read. Revisit if/when the parameter names can be confirmed
  (a live account trial or a published OpenAPI spec file).
- **`contacts` (`GET /v1/contacts`) is not present in SavvyCal's current public API
  documentation** — see "Streams notes" above. Preserved for legacy parity; not independently
  re-verified.
- **Every stream declares `projection: "passthrough"`** (conventions.md §8 rule 1): none of these
  list endpoints are documented as needing field renaming/filtering versus their raw wire shape,
  and legacy's own 3 streams already established this precedent (`savvycal.go:126`). Schemas stay
  intentionally minimal/documentation-only (conventions.md's minimal-honest depth precedent) rather
  than exhaustively enumerating every nested field SavvyCal's resource objects may carry (e.g.
  `Event`'s `EventAttendee`/`Payment` sub-objects, `Link`'s full booking-configuration fields) —
  `passthrough` guarantees no real field is dropped regardless of schema breadth.
- No incremental filtering is modeled for any stream, matching legacy's `Catalog` (no declared
  `CursorFields`) and the absence of a documented updated-since filter parameter on any SavvyCal
  list endpoint.
