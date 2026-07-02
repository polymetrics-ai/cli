# Overview

Sendinblue (Brevo) is a wave2 fan-out declarative-HTTP migration. It reads Brevo contacts, email
campaigns, contact lists, and senders through the Brevo API v3 (`GET
https://api.brevo.com/v3/...`). This bundle is engine-vs-legacy parity-tested against
`internal/connectors/sendinblue` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Brevo API key via the `api_key` secret; it is sent as the `api-key` request header
(never a Bearer token), matching legacy's `connsdk.APIKeyHeader("api-key", key, "")`
(`sendinblue.go:102`) — no prefix is prepended to the header value. `base_url` defaults to
`https://api.brevo.com/v3` and may be overridden for tests or proxies.

## Streams notes

All 4 streams share the same `offset_limit` pagination shape (`limit`/`offset` query params,
`page_size: 100`), matching legacy's `connsdk.OffsetPaginator{LimitParam: "limit", OffsetParam:
"offset", PageSize: pageSize}` at its default `page_size` of 100 (`sendinblue.go:19,87`). Record
envelopes: `contacts` → `contacts` key, `email_campaigns` → `campaigns` key (note the API's own
"campaigns" key differs from both the stream name `email_campaigns` (renamed from legacy's
camelCase `emailCampaigns` to satisfy this dialect's `snake_case` stream-naming rule,
conventions.md §2 — the underlying HTTP resource path is unaffected, still `/emailCampaigns`) and
the raw resource path — legacy's own `streamEndpoints` map records this exact key mismatch,
`sendinblue.go:109`), `contacts_lists` (path
`/contacts/lists`) → `lists` key, `senders` → `senders` key. `contacts` and `email_campaigns`
declare `x-cursor-field: modifiedAt` matching legacy's declared `CursorFields`, but neither stream
sends an incremental request filter — legacy declares the cursor field for catalog/UI purposes
only and never wires a `updatedAtFilter`-style query param anywhere in `sendinblue.go`; this bundle
matches that by declaring no `incremental` block on any stream (full refresh only), consistent
with the schema still carrying `x-cursor-field` per conventions.md §2 wherever a cursor concept
exists on the legacy stream. `id` is typed `integer` (not string) on every stream, matching
legacy's own declared `Field{Name: "id", Type: "integer"}` and Brevo's real numeric wire shape.

## Write actions & risks

None. Legacy's package declares `Capabilities.Write: false` and its `Write` method always returns
`connectors.ErrUnsupportedOperation`; `capabilities.write` is `false` here and this bundle ships no
`writes.json`.

## Known limits

- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) stamps a `previous_cursor` field (echoing
  `req.State["cursor"]` when a prior cursor happens to be set) onto fixture-mode records
  (`sendinblue.go:123-137`). This is not part of the LIVE record shape; this bundle's schemas and
  fixtures target the live path only.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`sendinblue.go:79-86`). The engine's `offset_limit` paginator's `PageSize` is a fixed
  bundle-authored value (`streams.json`'s `base.pagination.page_size`), not resolved from a
  `spec.json` config template, and the `limit`/`offset` query params it sends come from that same
  fixed value — declaring a `{{ config.page_size }}` query template alongside it would be dead
  code (the paginator's own query entry unconditionally wins per `mergeQuery`, `read.go`). This
  bundle sends legacy's own default `limit=100` on every request, matching legacy's un-overridden
  behavior; a caller wanting a different page size cannot express it through this bundle today.
  `max_pages` (legacy's hard request-count cap, defaulting to 1) is likewise not modeled; this
  bundle relies solely on the short-page stop signal (fewer than `page_size` records on a page) for
  termination.
