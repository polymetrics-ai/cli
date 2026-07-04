# Overview

Circa is a Pass B full-surface declarative-HTTP migration. It reads and writes Circa events,
contacts, companies, teams, custom fields, and event/company sub-resources through the Circa REST
API (`https://app.circa.co/api/v1/...`), verified against the real Postman-hosted API reference at
`docs.circa.co` (collection fetched 2026-07-04). This bundle originally targeted capability parity
with `internal/connectors/circa` (the hand-written connector it migrates, read-only); the legacy
package stays registered and unchanged until wave6's registry flip. This bundle now goes beyond
legacy's read-only scope to cover Circa's full documented CRUD surface via `writes.json`.

## Auth setup

Provide a Circa API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged. `base_url` defaults to
`https://app.circa.co/api/v1` and may be overridden for tests/proxies.

## Streams notes

`events`/`contacts`/`companies`/`teams`/`fields` read Circa's `{data:[...]}` list envelope directly.
Pagination is `page_number` (`page_param: page`, `size_param: ""` — Circa's API never accepts/needs
a page-size query parameter; legacy's own `harvest` only ever sends `page`, matching the engine's
`page_number` paginator with an empty `size_param`, `start_page: 1`, static `page_size: 25` matching
legacy's `circaDefaultPageSize`). The engine's short-page stop (`recordCount < page_size`) is
identical to legacy's own stop condition (`len(records) < pageSize`).

`events`/`contacts`/`companies` are incremental (`x-cursor-field: updated_at`), sent as
`updated_at[min]` (`incremental.request_param`) computed from the sync's persisted cursor or, on a
fresh sync, from the RFC3339 `start_date` config value (`incremental.start_config_key`) — identical
to legacy's `incrementalLowerBound`. `teams`/`fields` are full-refresh-only streams (no
`incremental` block); Circa's `/fields` endpoint has no `updated_at`/`created_at` timestamp at all.

`event_contacts`/`event_staff`/`event_expenses` are **event-scoped sub-resources**: this bundle uses
the engine's `stream.fan_out` dialect (`ids_from.request` against `/events`, `into.path_var:
event_id`, `stamp_field: event_id`) to list every event, then read each event's
`/events/{event_id}/contacts|staff|expenses` sub-resource, stamping `event_id` onto every emitted
record — the same pattern already used by this wave's cisco-meraki bundle for organization-scoped
sub-resources. `company_contacts` uses the identical fan_out pattern scoped to `/companies` instead.
`event_staff` has no natural `id` field in Circa's own API response (its primary key is
`[event_id, email]`); `event_contacts`/`event_expenses`/`company_contacts` use `[<parent>_id, id]`.

## Write actions & risks

`capabilities.write` is `true`. Nine actions cover Circa's full documented mutation surface:
`create_contact`/`update_contact`/`delete_contact`, `create_event`/`update_event`/`delete_event`,
`create_company`/`update_company`/`delete_company` (all direct JSON-body CRUD against their
respective resource paths), and `add_event_contact`/`update_event_contact`/`remove_event_contact`
(the event-contact-registration lifecycle: `POST /events/{event_id}/contacts` with a `contact_id`
body registers an existing contact onto an event; `PATCH .../contacts/{contact_id}` updates
`attendance_status`/`registration_status`; `DELETE .../contacts/{contact_id}` removes the
registration). All three delete-kind actions (`delete_contact`/`delete_event`/`delete_company`/
`remove_event_contact`) declare `delete.missing_ok_status: [404]` (idempotent delete) and carry an
irreversible-mutation risk note requiring approval; the remaining six creates/updates carry an
external-mutation risk note. This exceeds legacy's own read-only scope (`Write` unconditionally
returned `ErrUnsupportedOperation`) — a deliberate Pass B capability expansion, not a parity
deviation, since it only ADDS capability legacy never had.

## Known limits

- `page_size` (legacy's config-driven page-size override, 1-100, default 25) and `max_pages`
  (legacy's 0/all/unlimited-or-positive-integer request-count cap) are not runtime-configurable
  here: the engine's `page_number` paginator's `PageSize` is a static value set once in
  `streams.json` (25, matching legacy's `circaDefaultPageSize`), and `PaginationSpec` has no
  `MaxPages` field this paginator type reads. `spec.json` intentionally omits both `page_size` and
  `max_pages` — a declared-but-unwireable key is worse than an absent one (conventions.md F6),
  matching the aha/appfigures/cin7 wave2 precedent for the same paginator-type limitation.
- Legacy's `base_url` SSRF-guard scheme/host validation (https/http only, host required) is
  reproduced by the engine's own base-URL handling; no bundle-level behavior change.
- The async event-contacts-export job flow (`POST/GET/DELETE
  /events/{event_id}/contacts/exports[/{export_id}]`) is out of scope: it is a
  create-then-poll-then-download async job pattern with no declarative equivalent in this dialect,
  and its underlying data (event contacts) is already covered by the `event_contacts` stream. See
  `api_surface.json`.
- Every single-object detail `GET` endpoint (`/contacts/{id}`, `/events/{id}`, `/companies/{id}`,
  `/teams/{id}`, `/fields/{id}`, `/events/{id}/contacts/{id}`) is excluded as `duplicate_of` its
  corresponding list stream's already-covered per-item record shape — see `api_surface.json`.
