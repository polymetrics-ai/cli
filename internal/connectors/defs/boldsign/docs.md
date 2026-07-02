# Overview

BoldSign is a wave2 fan-out declarative-HTTP migration. It reads BoldSign documents, templates,
teams, contacts, and brands through the BoldSign REST API
(`GET https://api.boldsign.com/v1/<resource>/list`). This bundle is migrated from
`internal/connectors/boldsign` (the hand-written connector); the legacy package stays registered
and unchanged until wave6's registry flip.

## Auth setup

Provide a BoldSign API key via the `api_key` secret; it is sent as the `X-API-KEY` header
(`streams.json` `base.auth`'s `api_key_header` mode), matching legacy's
`connsdk.APIKeyHeader(boldsignAPIKeyHeader, secret, "")` (`boldsign.go:241`). Never logged.
`base_url` defaults to `https://api.boldsign.com` and may be overridden for tests/proxies.

## Streams notes

All five streams are page-numbered list endpoints (`Page`/`PageSize` query params, `page_size: 50`
— legacy's own `boldsignDefaultPageSize`); the engine stops on a short/empty page, matching
legacy's own `len(records) < pageSize` stop rule. Four streams (`documents`, `templates`,
`contacts`, `brands`) wrap their records in a `result` envelope; `teams` is the documented
exception using `results` (legacy's own comment, `boldsign/streams.go:11`) — each stream's
`records.path` captures this per-stream difference exactly as legacy's per-endpoint
`recordsPath` table does.

Every stream's record mapper renames camelCase raw API fields to this bundle's snake_case schema
fields via `computed_fields` (e.g. `document_id` from `record.documentId`, `sender_email` from
`record.senderEmail`); each is a bare single `{{ record.<path> }}` reference, so the engine's typed
extraction preserves the raw JSON type for `is_deleted`/`enable_signing_order`/`is_shared_template`/
`is_default` (booleans) and `sender_detail` (object), `signer_details`/`labels`/`users` (arrays) —
matching legacy's `mapRecord` functions field-for-field. Fields whose raw and schema names already
match (`status`, `labels`, `users`) pass through via plain schema projection with no
`computed_fields` entry needed.

`documents` and `templates` declare `created_date` as `x-cursor-field`, matching legacy's
`CursorFields: []string{"created_date"}`; `teams` likewise declares `created_date`. None of the
three actually filter server-side or client-side (legacy declares the cursor field for
manifest-surface parity only — BoldSign's list endpoints have no incremental filter parameter),
so no `incremental` block is declared on any stream, matching legacy exactly. `contacts` and
`brands` declare no cursor field at all, matching legacy (`PrimaryKey` only, no `CursorFields`).

## Write actions & risks

None. BoldSign's write surface (sending signature requests, uploading documents) is not a safe
generic reverse-ETL target; legacy's own package doc makes this explicit. `capabilities.write` is
`false` and this bundle ships no `writes.json`. Legacy's `Write` additionally returns
`RecordsFailed: len(records)` alongside `ErrUnsupportedOperation` (`boldsign.go:96-98`) — a detail
of legacy's in-process error-accounting contract with no declarative equivalent (the engine's own
unsupported-write path for a `capabilities.write: false` bundle governs this uniformly across
every read-only connector), not modeled here.

## Known limits

- **The mixed-case id fallback (`documentID`/`teamID`/`brandID`) is not modeled.** Legacy's
  `firstString(item, "id", "contactId")`-style helper (`boldsign/streams.go:187-194`) falls back to
  an alternate-cased key (e.g. `documentId` OR `documentID`) when the primary key is absent. The
  engine's `computed_fields` dialect has no multi-path fallback filter (only a single `{{ record.
  <path> }}` reference per field). This bundle wires the PRIMARY key observed in every available
  fixture/test (`documentId`/`teamId`/`brandId`/`id` — camelCase, confirmed against legacy's own
  `boldsign_test.go` fixtures, which never exercise the alternate-cased fallback), and does not
  model the defensive alternate-casing fallback. If BoldSign's live API ever emits the alternate
  casing for a given record, that record's primary-key field would resolve to `null` here where
  legacy would have recovered it — an `ACCEPTABLE` deviation only because the fallback is
  observably dead in every test/fixture on record; flagged here rather than silently dropped.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size`
  (1-100, default 50) and `max_pages` (0/all/unlimited default) as config-driven overrides
  (`boldsignPageSize`/`boldsignMaxPages`, `boldsign.go:274-302`). The engine's `page_number`
  paginator's `page_size`/`max_pages` are fixed values baked into `streams.json`'s
  `base.pagination` block at bundle-author time, matching the identical, already-documented
  searxng/bitly precedent. Neither is declared in `spec.json` (F6, REVIEW.md).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only
  reached when `config.mode == "fixture"`) stamps `connector: "boldsign"`, `fixture: true`, and a
  conditional `previous_cursor` onto every fixture-mode record (`boldsign.go:213-219`). None of
  these are part of the LIVE record shape; this bundle's schemas and fixtures target the live
  `harvest` path only.
