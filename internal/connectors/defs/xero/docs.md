# Overview

Xero is a wave2 fan-out declarative-HTTP migration. It reads Xero accounting data (invoices,
contacts, accounts, bank transactions, items, and payments) from the Xero Accounting API
(`GET https://api.xero.com/api.xro/2.0/...`). This bundle migrates `internal/connectors/xero` (the
hand-written connector it replaces); the legacy package stays registered and unchanged until
wave6's registry flip. Xero is read-only here: the Accounting API supports writes, but reverse-ETL
writes are not enabled for this connector, matching legacy.

## Auth setup

Provide a Xero OAuth2 access token via the `access_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <access_token>`), matching legacy's `connsdk.Bearer(token)`
(`xero.go:247`). Provide the Xero tenant (organisation) id via the `tenant_id` secret; it is sent
as the `Xero-tenant-id` header on every request, matching legacy's `DefaultHeaders`
(`xero.go:249-251`). Neither value is ever logged. `base_url` defaults to
`https://api.xero.com/api.xro/2.0` and may be overridden for tests/proxies, matching legacy's
`xeroDefaultBaseURL` constant.

## Streams notes

All six streams (`invoices`, `contacts`, `accounts`, `bank_transactions`, `items`, `payments`) hit
their respective Xero Accounting API resource paths (`Invoices`, `Contacts`, `Accounts`,
`BankTransactions`, `Items`, `Payments`); Xero's list envelope shape is
`{"Id":..., "Status":"OK", "<Resource>":[...]}`, so `records.path` names the same resource segment
as the URL path, matching legacy's `xeroStreamEndpoints` routing table. Pagination is 1-based
page-number pagination (`page_number`, no size query param — Xero's own fixed 100-per-page
envelope), stopping on a short page (fewer than 100 records), matching legacy's `harvest` exactly
(`if len(records) < xeroDefaultPageSize`).

Every stream's primary key is the resource's own native GUID field (`InvoiceID`, `ContactID`,
`AccountID`, `BankTransactionID`, `ItemID`, `PaymentID`), matching legacy's per-resource
`idField`/`PrimaryKey`. `computed_fields` reproduces legacy's lower-cased convenience aliases
(`id`, `type`, `status`, `updated_at`) via bare single-reference templates (typed extraction
preserves each field's native JSON type — `id`/`type`/`status`/`updated_at` are all Xero strings,
so no widening is needed), and reproduces legacy's `contactID(item)` nested-extraction helper for
`invoices`/`bank_transactions` (`ContactID: "{{ record.Contact.ContactID }}"`, since Xero's real
wire shape nests the contact reference under a `Contact` object rather than exposing a top-level
`ContactID`). The cursor field for every stream is `UpdatedDateUTC` (Xero's own last-modified
timestamp), matching legacy's `xeroStreams()` catalog declaration; as with legacy, no stream
actually performs incremental cursor-based filtering (there is no `If-Modified-Since`
header/query-param wiring in either `xero.go`'s `harvest` or this bundle) — both connectors always
perform a full stream read, and `UpdatedDateUTC` is declared solely for manifest-surface parity
(matching legacy's own `InitialState`, which always seeds an empty cursor with no read-time
consumption).

## Write actions & risks

None. This connector is read-only in both legacy and this bundle (`capabilities.write: false`); no
`writes.json` is shipped, even though the underlying Xero Accounting API does support writes.

## Known limits

- **`tenant_id`'s config-fallback convenience is not modeled.** Legacy resolves `tenant_id` from
  secrets first, falling back to a plain (non-secret) config value for convenience
  (`xeroTenantID`, `xero.go:265-275`). The engine's header-templating dialect resolves a single
  template per header with no multi-source fallback mechanism (unlike `auth`'s first-match-wins
  candidate list), so this bundle declares `tenant_id` as a single required `x-secret` field only
  wired from `secrets.tenant_id`. This is a documented config-surface narrowing: a caller who
  previously supplied `tenant_id` via plain config must now supply it as a secret instead; the
  emitted header value and every downstream request/response are byte-identical once they do.
- **The declarative `Check` request cannot reproduce legacy's authorization pre-flight ordering.**
  Legacy's `Check` explicitly validates `access_token`/`tenant_id` presence with connector-specific
  error messages before ever issuing a request (`xero.go:74-79`); the engine's `auth`/header
  resolution instead hard-errors on an absent `secrets.access_token`/`secrets.tenant_id` reference
  at request-build time with the engine's own generic unresolved-key error text. Both sides reject
  the identical missing-credential input; only the error message differs (conventions.md §5's
  established config-validation-parity precedent — bucketed by reason, not byte-matched text).
- **Legacy's fixture-mode-only fields (`readFixture`, reached only when `config.mode ==
  "fixture"`) are not modeled**, including the static `connector: "xero"` / `fixture: true` marker
  fields and the `previous_cursor` echo. This bundle's schemas and fixtures target the live wire
  shape only; the engine's own fixture-replay conformance harness supplies the credential-free
  test affordance legacy's fixture mode existed for, matching the precedent recorded in
  `docs/migration/conventions.md`'s bitly ledger entry.
