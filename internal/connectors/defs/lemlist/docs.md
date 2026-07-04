# Overview

Lemlist is a cold outreach / sales engagement platform. This bundle reads the legacy team,
campaigns, activities, and unsubscribes streams plus additional no-required-input API collections:
team senders/credits/CRM users, schedules, People database filters, tasks, inbox labels, contacts,
contact lists, companies, webhooks, unsubscribe variables, Signal Agent signals, channel status,
and contact/company field definitions. Requests use the Lemlist REST API
(`https://api.lemlist.com/api`). It migrates `internal/connectors/lemlist` (the hand-written
connector); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Lemlist API key via the `api_key` secret. It is sent as the `access_token` query
parameter on every request (`api_key_query` auth mode) — Lemlist's documented auth convention;
never logged.

## Streams notes

`campaigns`, `activities`, `unsubscribes`, `schedules`, `contacts`, `companies`,
`unsubscribed_variables`, and `watchlist_signals` use offset/limit pagination and stop on a short
page. `tasks` uses lemlist's page-number shape starting at page 0. Single-object and unpaginated
array endpoints (`team`, `team_senders`, `team_credits`, `team_crm_users`, `database_filters`,
`inbox_labels`, `contact_lists`, `webhooks`, `user_channels`, `fields_contact`, and
`fields_company`) override pagination to `none`.

The legacy streams still preserve the original record contracts: root arrays for `campaigns`,
`activities`, and `unsubscribes`, and a single root object for `team`. New streams use documented
response envelopes: `schedules` reads `schedules`, `tasks` reads `results`, CRM contacts/companies
read `data`, Signal Agent signals read `signals`, and field-definition streams split the `/fields`
response into `data.contact` and `data.company`. No stream has a legacy incremental filter
mechanism, so no `incremental` block is declared anywhere in this bundle.

## Write actions & risks

None. This bundle is read-only (`capabilities.write: false`); Lemlist exposes no reverse-ETL write
surface here (matches legacy's `Write` returning `connectors.ErrUnsupportedOperation`
unconditionally). Lemlist's documented create/update/delete endpoints can add leads, send messages,
change campaign/schedule/task state, alter CRM/contact/company records, configure webhooks, manage
email accounts, or unsubscribe/resubscribe contacts; those mutations are excluded from this
read-only source bundle with concrete risk categories in `api_surface.json`.

## Known limits

- Legacy exposed runtime-configurable `page_size` (1-100) and `max_pages` (0/all/unlimited) config
  knobs. Neither is expressible in this dialect: `PaginationSpec`'s `page_size`/`max_pages` fields
  are fixed JSON literals in `streams.json`, never resolved from `RuntimeConfig.Config` at read
  time (no existing bundle in this codebase templates them from config; this mirrors the stripe
  golden's own documented `page_size`/`max_pages`-is-dead-config precedent, ledger item 3,
  `docs/migration/conventions.md` §5). This bundle fixes the page size at 100 (Lemlist's own
  documented max and legacy's own default) and leaves pagination unbounded (`max_pages` absent =
  no cap), matching legacy's own default behavior for a caller that never overrides either knob.
  Documented scope narrowing, not a data-shape deviation — the emitted records for any given page
  are identical either way.
- GET endpoints that require caller-supplied IDs or required query parameters are not global
  streams. Campaign reports require a `campaignIds` query value, CRM filters require CRM/user
  selectors, inbox conversations require a user/contact ID, and detail endpoints require campaign,
  schedule, lead, contact, user, export, enrichment, or unsubscribe identifiers.
- Export endpoints that return or initiate CSV/file downloads are excluded as binary/report
  payloads rather than JSON object streams.
