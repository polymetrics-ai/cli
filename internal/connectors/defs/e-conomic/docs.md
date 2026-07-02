# Overview

e-conomic is a wave2 fan-out declarative-HTTP migration. It reads e-conomic customers, products,
suppliers, accounts, and booked invoices through the e-conomic REST API
(`GET https://restapi.e-conomic.com/...`). This bundle migrates
`internal/connectors/e-conomic` (the hand-written connector); the legacy package stays registered
and unchanged until wave6's registry flip. e-conomic is read-only: legacy exposes no reverse-ETL
writes, so `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Auth setup

Provide two secrets: `app_secret_token` (the e-conomic app's secret) and `agreement_grant_token`
(the per-agreement grant). Both are sent as plain request headers —
`X-AppSecretToken: <app_secret_token>` and `X-AgreementGrantToken: <agreement_grant_token>` — via
`streams.json`'s `base.headers`, matching legacy's `requester` construction exactly
(`e_conomic.go`'s `appSecretHeader`/`agreementGrantHead` constants). Both secrets are required;
neither has a fallback. `base_url` defaults to `https://restapi.e-conomic.com` and may be
overridden for tests/proxies.

## Streams notes

All 5 streams (`customers`, `products`, `suppliers`, `accounts`, `invoices`) share the same list
shape: `GET` against the e-conomic collection path, records at `collection`, and e-conomic's own
`skippages`/`pagesize` pagination convention with `pagination.nextPage` (an absolute URL) driving
the next page (`pagination.type: next_url`, `next_url_path: "pagination.nextPage"`) — matching
legacy's `harvest` function exactly (follow the absolute `nextPage` URL directly; stop when it is
empty). `invoices` reads from `/invoices/booked` (legacy only ever implements booked invoices, not
drafts). Every stream's initial request sends `skippages=0&pagesize=100` (legacy's
`defaultPageSize`); `page_size`/`max_pages` are not modeled as runtime config (see Known limits).

Every stream requires a `computed_fields` rename because e-conomic's wire shape uses camelCase
field names (`customerNumber`, `salesPrice`, `creditLimit`, ...) while this bundle's schemas use
snake_case, and schema-as-projection matches only by exact raw-key equality (conventions.md §2) —
a plain projection with no renames would silently drop every e-conomic field. Nested business-key
references (e.g. `{"vatZone":{"vatZoneNumber":1}}`) are flattened via dotted `record.<path>`
references (`{{ record.vatZone.vatZoneNumber }}`), matching legacy's `refNumber` helper — a bare
single reference like this receives the engine's typed extraction (conventions.md §3), preserving
the real integer type, exactly like legacy's `refNumber` returning the raw `any` value.

## Write actions & risks

None. e-conomic is a read-only source in legacy (`e_conomic.go`'s package doc: "The e-conomic
source is read-only (full-refresh); it exposes no reverse-ETL writes"); `capabilities.write` is
`false` and no `writes.json` is shipped.

## Known limits

- **No incremental cursor.** Legacy exposes no incremental cursor field for any e-conomic stream
  (`economicStreams()` declares no `CursorFields` anywhere) — every stream is full-refresh only in
  both legacy and this bundle; no `incremental` block is declared on any stream, matching legacy
  exactly.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size` (default
  100, capped at 1000) and `max_pages` (0/all/unlimited = unbounded) as config-driven overrides
  (`pageSize`/`maxPages` in `e_conomic.go`). The engine's `next_url` paginator has no
  config-driven page-size or max-pages knob (it never reads `PaginationSpec.PageSize`/`MaxPages`
  the way `page_number`/`offset_limit` do), so this bundle sends legacy's own default
  (`pagesize=100`) as a static per-stream query literal and does not declare `page_size`/`max_pages`
  in `spec.json` at all (a declared-but-unwireable config key is worse than an absent one, per
  conventions.md F6 precedent). Pagination is bounded only by e-conomic's own empty-`nextPage` stop
  signal, matching e-conomic's real termination behavior.
- **Legacy's fixture-mode-only synthetic fields are not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`, a credential-free conformance-harness affordance
  in the legacy Go connector) stamps deterministic placeholder records with fields shaped after —
  but not identical to — the live wire shape. This bundle's schemas and fixtures target the LIVE
  record shape only (`e_conomic.go`'s `harvest`/`mapRecord` functions), per the bitly-pilot
  precedent (`docs/migration/conventions.md`'s worked example): the engine's own
  fixture-replay conformance harness supersedes the need for an in-connector fixture-mode branch.
- **`next_url` pagination ships single-page conformance fixtures for every stream** (the sanctioned
  exception, conventions.md §4): e-conomic's `pagination.nextPage` is an absolute URL whose host is
  the live API (or, in a real sync, whatever `base_url` resolves to) — a static fixture file cannot
  embed the conformance replay server's own dynamically-assigned address ahead of time. All 5
  streams share the identical base-level `next_url` pagination (unlike bitly, which has one
  non-paginated stream available to serve as the `pagination_terminates` candidate), so every
  stream fixture here is single-page; `pagination_terminates` exercises `customers` (the first
  declared stream) against its one-page fixture, which trivially proves the read terminates and
  consumes exactly the one recorded page. A live two-page proof (an `httptest.Server` asserting the
  engine correctly follows a `nextPage` URL across two real pages) is out of scope for this wave
  (JSON+docs only, no `paritytest` packages) and is a documented follow-up: extend this bundle with
  a `paritytest/e-conomic` suite (mirroring bitly's/calendly's `next_url` parity tests) in a
  subsequent wave.
