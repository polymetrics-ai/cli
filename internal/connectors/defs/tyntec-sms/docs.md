# Overview

tyntec SMS reads messages, templates, sender IDs, and delivery reports from the tyntec Messaging
API (`GET {base_url}/sms/v1/<resource>`). Pass B also adds the documented SMS send mutation as an
approval-gated write action. The legacy package stays registered and unchanged until wave6's
registry flip.

## Auth setup

Requires one secret: `api_key` (tyntec API key), sent as the `apikey` header on every request via
`streams.json` `base.auth`'s `api_key_header` mode — matching legacy's
`connsdk.APIKeyHeader("apikey", key, "")` exactly (no prefix). `base_url` defaults to
`https://api.tyntec.com` (legacy's `defaultBaseURL`), overridable for tests or proxies.

## Streams notes

Four streams, all sharing the identical `page_number` pagination shape (legacy's own
`limit`/`page` query convention): `messages` (`GET sms/v1/messages`, records at `messages`),
`templates` (`GET sms/v1/templates`, records at `templates`), `sender_ids` (`GET
sms/v1/sender-ids`, records at `sender_ids`), and `delivery_reports` (`GET sms/v1/reports`, records
at `reports`). Pagination sends `limit=<page_size>&page=<n>` and stops on a short page (fewer than
`page_size` records), matching legacy's `harvest` loop exactly; `page_size` defaults to 100
(legacy's `defaultPageSize`) and is a fixed bundle-authored value (see Known limits).

`messages` and `delivery_reports` declare `created_at` as their cursor field (manifest-surface
parity with legacy's `CursorFields: []string{"created_at"}`); a `computed_fields` rename maps the
raw API's `createdAt` field to `created_at` (`"created_at": "{{ record.createdAt }}"`). Legacy's
`first(item, "createdAt", "created_at")` helper also tolerates a record that already uses the
`created_at` key directly (no `createdAt`) — this bundle reproduces that fallback for free: when
`computed_fields`' `record.createdAt` reference is absent on a given record, the field is silently
skipped and whatever the schema-projection step already copied from a same-named `created_at`
source field survives untouched, exactly matching legacy's two-key fallback. Neither legacy nor
this bundle actually filters or advances reads by the cursor field server-side; the API supports no
incremental filter parameter, so a full stream read is always performed (no `incremental` block is
declared).

## Write actions & risks

`send_message` posts to `POST sms/v1/messages` and sends an SMS through tyntec. Records must provide
`to`, `from`, and `text`; optional `reference` and `callbackUrl` fields are forwarded when present.
This is a billable, externally visible side effect and requires reverse-ETL plan preview and
approval before execution.

## Known limits

- **`page_size`/`max_pages` are not exposed as runtime config.** Legacy accepts `config["page_size"]`
  (1-1000, default 100) and `config["max_pages"]` (default 1; `"all"`/`"unlimited"` for unbounded)
  as caller-overridable values. The engine's `PaginationSpec.PageSize`/`MaxPages` fields are plain
  JSON integers fixed at bundle-authoring time in `streams.json`'s `base.pagination` block — there is
  no templated/config-driven override mechanism for either field (`bundle.go`'s `PaginationSpec`
  carries no `{{ }}`-templated variant of these fields). Declaring `page_size`/`max_pages` as
  `spec.json` properties that no template in the bundle ever consumes would be dead config (F6,
  REVIEW.md; see also searxng's identical precedent), so neither is declared. This bundle bakes in
  legacy's own DEFAULT values instead: `page_size: 100`, `max_pages: 1` — reproducing the exact
  behavior a caller who never overrides either config key already gets from legacy. A caller who
  previously relied on setting `max_pages=all`/a larger `page_size` to pull more than one page's
  worth of records per read loses that override; out of scope for this wave, not silently wrong for
  the common (unconfigured) case.
- The stale metadata URL `https://api.tyntec.com/reference/messaging` returned 404 during Pass B.
  The official reference index at `https://api.tyntec.com/reference/` was reachable and links the
  SMS API reference; HTTPS fetches of the linked `sms/current.html` page failed from this sandbox,
  so the write schema is kept to the stable documented SMS send fields (`to`, `from`, `text`) plus
  two common optional fields.
