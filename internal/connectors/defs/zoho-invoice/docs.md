# Overview

Zoho Invoice is a wave2 fan-out migration from `internal/connectors/zoho-invoice` (the
hand-written Go connector it replaces). It reads Zoho Invoice customers, invoices, and payments
through the Zoho Invoice REST API v3. Read-only, matching legacy's capabilities exactly (`Write`
returns `ErrUnsupportedOperation`). The legacy package stays registered and unchanged until wave6's
registry flip.

## Auth setup

Provide a Zoho OAuth access token via the `access_token` secret. It is sent as the `Authorization`
header with legacy's exact non-standard prefix (`Zoho-oauthtoken <access_token>`, NOT the standard
`Bearer <token>` shape) via `streams.json` base `auth`'s `api_key_header` mode
(`header: Authorization`, `prefix: "Zoho-oauthtoken "`) — never logged. An optional
`organization_id` config value is sent as the `organization_id` query parameter on every stream
request when set; when unset, it is omitted entirely (not sent empty), matching legacy's
`baseQuery` behavior exactly.

## Streams notes

All 3 streams share the same shape: `GET` against the Zoho Invoice list endpoint, `page_number`
pagination (`page`/`per_page` query params, default page size 200, configurable 1-200 via the
`page_size` config value, stop on a short page — identical to legacy's `harvest` loop). `max_pages`
(default unbounded; `0`/`all`/`unlimited` all mean unbounded) caps the number of pages read,
matching legacy's `maxPages` parsing exactly.

- `customers`: records at the top-level `customers` array key.
- `invoices`: records at the top-level `invoices` array key.
- `payments`: legacy names this stream `payments` but calls the real Zoho Invoice
  `customerpayments` endpoint and reads records at the top-level `customerpayments` array key — the
  bundle reproduces this exact path/records-path mismatch (stream name `payments`, `path:
  "/customerpayments"`, `records.path: "customerpayments"`).

Every stream uses `projection: "passthrough"` (every raw API field survives, matching legacy's
`mapRecord`, which copies every input field verbatim) plus a `computed_fields` alias for each
stream's authoritative primary-key/name/cursor field to the parity fields `id`/`name`/`updated_at`
legacy also synthesizes (`customer_id`->`id`, `customer_name`->`name`, `last_modified_time`->
`updated_at` for `customers`; `invoice_id`->`id`/`invoice_number`->`name` for `invoices`;
`payment_id`->`id`/`payment_number`->`name` for `payments`). `computed_fields`' bare
`{{ record.<path> }}` shape gets typed extraction (native JSON type preserved) and is silently
skipped when the source path is absent on a given record, matching legacy's own
`out["id"] == nil` fallback-only-if-absent semantics for the common case where the field is
already present.

No stream declares an `incremental` block: legacy's `harvest`/`readFixture` never sends a
server-side updated-since filter (there is no `updated_at`-style query parameter anywhere in
`baseQuery`) despite exposing an `updated_at` cursor field on the catalog — the cursor is
informational/state-tracking only in legacy, never used to filter requests. This bundle reproduces
that exact behavior: every sync is effectively a full refresh over the paginated list.

## Write actions & risks

None. Legacy `zoho-invoice` is read-only (`Capabilities.Write: false`); this bundle ships no
`writes.json`.

## Known limits

- Legacy's `firstValue` fallback tried MULTIPLE candidate keys per derived field in priority order
  (e.g. `customers.id`: `customer_id`, then `contact_id`, then bare `id`; `payments.id`:
  `payment_id` (declared twice), then bare `id`) to guard against alternate/legacy API response
  shapes. This bundle's `computed_fields` aliases only the FIRST (authoritative, real-wire-shape)
  candidate key per field — the engine's `computed_fields` dialect has no multi-key fallback-chain
  primitive (each entry is a single template, not an ordered candidate list). The secondary
  fallback keys are legacy defensive coding for a shape the real, documented Zoho Invoice v3 API
  never actually emits (confirmed against `docs_url` and the legacy connector's own tests, which
  only ever exercise the first key). Documented scope narrowing, not a silent behavior change for
  any real Zoho Invoice response. See `docs/migration/conventions.md`'s parity-deviation ledger
  meta-rule.
- No server-side incremental filtering is modeled (see Streams notes above) — this matches legacy
  exactly, not a narrowing.
- Full Zoho Invoice API surface (estimates, credit notes, recurring invoices, expenses, items,
  projects) is out of scope for this wave; see `api_surface.json`'s `excluded: {category:
  out_of_scope}` entries. Only the 3 legacy-parity streams are implemented.
- `fixtures/streams/customers/{page_1,page_2}.json` is the required 2-page pagination proof (200
  full-size records on page 1 to trigger a genuine second request under the real `page_size`
  default, 1 record on page 2 to stop) — `invoices`/`payments` ship single-page fixtures, matching
  the stripe golden's "only the first declared stream proves 2-page termination" pattern.
