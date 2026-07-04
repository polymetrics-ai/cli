# Overview

Wufoo started as a wave2 fan-out declarative-HTTP migration (3 legacy-parity read streams) and was
expanded to full API-surface coverage in Pass B (`api_surface.json` `reviewed_at: 2026-07-04`). It
now reads 8 resources (forms, form fields, entries, form comments, reports, report fields, report
entries, report widgets) and writes 3 actions (submit a form entry, add a webhook, delete a
webhook) through the Wufoo API v3 (`GET {{ config.base_url }}/...`), re-derived directly from the
live docs page (https://wufoo.github.io/docs/). This bundle originally migrated from
`internal/connectors/wufoo` (the hand-written connector); the legacy package stays registered and
unchanged until wave6's registry flip, and is frozen at its original 3-stream, read-only surface —
it never gained the 5 additional streams or any write capability.

## Auth setup

Provide a Wufoo API key via the `api_key` secret; it is sent as the HTTP Basic username with the
literal password `pass` (`mode: basic`), matching Wufoo's documented convention exactly — Wufoo's
API accepts any password value when authenticating with an API key as the username. `base_url`
defaults to `https://example.wufoo.com/api/v3` (a placeholder subdomain) and must be overridden with
the account's actual subdomain (`https://<subdomain>.wufoo.com/api/v3`).

## Streams notes

Wufoo's REST surface is small and split cleanly into a Forms family and a Reports family, each with
list + detail + fields + entries/comments/widgets sub-resources; every stream shares HTTP Basic
auth from `base.auth`.

**Forms family**: `forms` (`GET /forms.json`, `page_number` pagination via `page`/`pageSize`,
matching legacy's paginator defaults) emits the top-level `Forms` array. `form_fields` (`GET
/forms/{{ config.form_hash }}/fields.json`) emits the `Fields` array — Wufoo does not paginate this
endpoint (a form's field list is always returned whole), so the stream overrides the base
pagination to `"type": "none"`. `entries` (`GET /forms/{{ config.form_hash }}/entries.json`) emits
the `Entries` array using the same legacy `page`/`pageSize` page-number paginator as `forms`;
`max_pages` is fixed at 1 to match legacy's default read. `form_comments` (`GET
/forms/{{ config.form_hash }}/comments.json`) emits the `Comments` array with the same
`pageStart`/`pageSize` offset convention, default page size 25 (Wufoo's documented default for this
endpoint, half the entries/forms default).

**Reports family** mirrors the Forms family exactly: `reports` (`GET /reports.json`, `page_number`)
emits `Reports`; `report_fields` (`GET /reports/{{ config.report_hash }}/fields.json`,
unpaginated) emits `Fields`; `report_entries` (`GET /reports/{{ config.report_hash }}/entries.json`,
`offset_limit` via `pageStart`/`pageSize`) emits `Entries`; `report_widgets` (`GET
/reports/{{ config.report_hash }}/widgets.json`, unpaginated — only Chart/Graph/Number widgets are
returned, per Wufoo's docs) emits `Widgets`.

All 8 streams declare `"projection": "passthrough"` — Wufoo's field set varies per-form/per-report
(custom `Field##` columns on `entries`), so passthrough is the only faithful representation; the
schema remains a documentation surface of the well-known common fields (`Hash`/`Name`/`DateUpdated`
for forms/reports, `EntryId`/`DateCreated`/`DateUpdated` for entries) rather than an exhaustive
per-account field enumeration. No stream declares an `incremental` block: none of Wufoo's list
endpoints accept a time-range query parameter (confirmed absent from every endpoint's documented
Query Parameters table) — `DateUpdated`/`DateCreated` are declared as `x-cursor-field` for manifest
documentation only; every read is a full sync.

## Write actions & risks

- **`submit_entry`** (`POST /forms/{{ config.form_hash }}/entries.json`, `body_type: form`) —
  submits a new entry to the configured form. The record schema is an open `additionalProperties:
  string` map (Wufoo's own convention: form fields are named `Field1`, `Field2`, ... `Field##`,
  unique per form and not statically enumerable without a live `form_fields` read first) rather
  than a fixed property list. External mutation; approval required.
- **`add_webhook`** (`PUT /forms/{{ config.form_hash }}/webhooks.json`, `body_type: form`) —
  registers a webhook callback URL (`url` required; `handshakeKey`/`metadata` optional) on the
  configured form. Wufoo's Webhooks resource is PUT-to-add/DELETE-to-remove only — there is no GET
  list endpoint at all (confirmed absent from the docs' resource table: "Webhooks ... PUT or
  DELETE"), so webhooks cannot be modeled as a read stream. External mutation; approval required.
- **`delete_webhook`** (`DELETE /forms/{{ config.form_hash }}/webhooks/{{ record.hash }}.json`,
  `path_fields: ["hash"]`) — removes a previously registered webhook by its hash.
  `delete.missing_ok_status: [404]` treats an already-removed webhook as a successful (idempotent)
  delete. Irreversible external deletion; approval required.

## Known limits

- **No incremental filter is modeled** on any stream: Wufoo's list endpoints accept no time-range
  query parameter (confirmed absent from every endpoint's documented Query Parameters table); every
  sync is a full read. `DateUpdated`/`DateCreated` are declared as `x-cursor-field` purely for
  manifest-surface documentation.
- **`entries` keeps legacy pagination.** Wufoo's public docs describe `pageStart`/`pageSize`, but
  the legacy connector used the same `page`/`pageSize` page-number paginator for `entries` as it
  used for `forms` and `reports`. The bundle preserves that legacy request shape for data fidelity;
  the newly added `form_comments` and `report_entries` streams use the documented offset paginator.
- **`webhooks` cannot be a read stream**: Wufoo's Webhooks resource genuinely has no list/GET
  endpoint (see Write actions above); only `add_webhook`/`delete_webhook` writes exist. An operator
  who registers a webhook via `add_webhook` and later needs to enumerate existing webhooks must
  track the returned hash out-of-band (e.g. from `add_webhook`'s own response body) — this is a
  real Wufoo API limitation, not an omission in this bundle.
- **`users.json`** (account sub-user administration) and **`/login`** (session-cookie auth for
  Wufoo's legacy widget/embed tooling) are excluded — see `api_surface.json` for the specific
  reasons. Both are out of scope for a business-data sync/write connector.
- **`page_size`/`max_pages` config-driven per-request overrides are not modeled**: the engine's
  `page_number`/`offset_limit` paginators read their page-size from the static `streams.json`
  pagination block only — there is no per-request config-driven override mechanism in the current
  dialect. `page_size`/`max_pages` remain declared in `spec.json` as documentation of the original
  legacy-parity bundle's accepted config surface, but neither is wired into any template.
