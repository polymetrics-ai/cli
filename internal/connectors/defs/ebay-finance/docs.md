# Overview

eBay Finance is a wave2 fan-out declarative-HTTP migration. It reads a seller's monetary records —
transactions, payouts, transfers, and the seller funds summary — through the eBay Sell Finances
REST API (`GET {{ config.base_url }}/...`). This bundle is migrated from
`internal/connectors/ebay-finance` (the hand-written connector it replaces); the legacy package
stays registered and unchanged until wave6's registry flip. Read-only (`capabilities.write` is
`false`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`).

## Auth setup

Provide an eBay OAuth user access token (scope `sell.finances`) via the `client_access_token`
secret; it is sent as a Bearer token (`mode: bearer`), matching legacy's
`connsdk.Bearer(secret)`. `base_url` defaults to `https://apiz.ebay.com/sell/finances/v1`
(eBay's production endpoint) and may be overridden for eBay's sandbox environment or test proxies.

## Streams notes

`transactions` (`GET /transaction`, records at `transactions`), `payouts` (`GET /payout`, records
at `payouts`), and `transfers` (`GET /transfer`, records at `transfers`) share `offset_limit`
pagination (`limit`/`offset` query params); `seller_funds_summary` (`GET
/seller_funds_summary`) is a single-object endpoint with no array wrapper (`records.path: ""`,
matching `connsdk.RecordsAt(resp.Body, "")`'s single-object-becomes-one-record behavior) and no
pagination (`pagination.type: "none"` overrides the base spec for this stream only).

Each of the three list streams sends an eBay date-range `filter` query param
(`omit_when_absent: true`, so it is left off entirely on a request with no resolvable lower
bound) built from `{{ incremental.lower_bound }}` — the persisted state cursor, falling back to
`start_date` — wrapped in eBay's `<field>:[<value>..]` range-filter syntax, matching legacy's
`ebayDateFilter`. **The filter field name for `transfers` is `transactionDate`, not
`transferDate`**, reproducing legacy's own `dateFilterField` mapping exactly (legacy uses the same
field name for both `transactions` and `transfers` filters, even though the transfer record's own
timestamp is called `transferDate`) — this looks like a legacy quirk, but per
`docs/migration/conventions.md`'s meta-rule, an existing legacy accepted-input behavior is
reproduced, not "corrected," during migration.

`amount`/`payoutInstrument`/`totalFunds`/`availableFunds`/`fundsOnHold`/`processingFunds` money
and nested objects are flattened into `<field>_value`/`<field>_currency` (or
`payoutInstrument_nickname`/`payoutInstrument_accountLastFourDigits`) scalar fields via
`computed_fields`, matching legacy's `flattenAmount` helper exactly. Primary keys are each
stream's natural id (`transactionId`/`payoutId`/`transferId`); `seller_funds_summary` declares no
primary key, matching legacy's empty `PrimaryKey: []string{}`. `x-cursor-field` on each list
stream's schema names the record's own timestamp field (`transactionDate`/`payoutDate`/
`transferDate`) even where it differs from the filter's field-name literal, above.

## Write actions & risks

None. Legacy `ebay_finance.go`'s `Write` returns `connectors.ErrUnsupportedOperation`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` config overrides are not modeled.** Legacy accepts optional
  `page_size` (1-1000, default 200) and `max_pages` (default unlimited, `all`/`unlimited`/`0`
  synonyms) config keys read at request time (`ebayPageSize`/`ebayMaxPages`). The engine's
  `PaginationSpec.PageSize`/`MaxPages` fields are plain fixed JSON integers baked into
  `streams.json` — there is no templating/config-driven override mechanism for them.
  `base.pagination.page_size` is set to legacy's real production default, `200`
  (`ebayDefaultPageSize`) — this is the actual value a live deployment's paginator sends; it is
  not a fixture convenience. The `transactions` stream declares a stream-level `pagination`
  override (`page_size: 2`) so its required 2-page conformance fixture
  (`fixtures/streams/transactions/{page_1,page_2}.json`, §4 of `docs/migration/conventions.md`)
  can stay small and readable; since stream-level `pagination` replaces the base spec wholesale,
  this is an intentional, ledgered per-stream deviation from legacy's uniform 200-record page
  size — `transactions` reads in smaller, more numerous pages than legacy would, everywhere else
  identical. `payouts` and `transfers` are unaffected and use legacy's true 200-record page size
  end-to-end, matching their single-page fixtures' `limit=200` request/response. No `max_pages`
  cap is declared (unbounded, matching legacy's own default). Neither key is declared in
  `spec.json` (F6, `docs/migration/conventions.md`: dead, unwireable config is worse than absent
  config). This never changes which records are emitted for an in-range request — only request
  cadence.
- **Legacy's secondary stop condition (`offset >= total` after a full page) is not modeled.**
  The engine's `offset_limit` paginator stops on a single signal: a page returning fewer records
  than `page_size`. When the true last page happens to be exactly `page_size` records long, the
  engine issues one additional (empty) request before stopping, where legacy could stop
  immediately using the response's own `total` field. This never changes the set of emitted
  records (the extra page is empty) and never loops — a strictly data-neutral extra-request
  difference, acceptable per `docs/migration/conventions.md` §5's meta-rule.
- **`base_url`/`api_host` scheme/host validation is enforced by legacy in Go** with dedicated
  error messages (`ebayBaseURL`); the engine has no equivalent declarative URL-shape validator, so
  a malformed `base_url` here surfaces as a generic request-construction/connection error rather
  than legacy's specific messages. This bundle also drops legacy's separate `api_host` config
  knob (a bare host that legacy suffixes with `/sell/finances/v1`) in favor of a single directly
  settable `base_url` (default `https://apiz.ebay.com/sell/finances/v1`) — functionally
  equivalent, a config-surface narrowing rather than an emitted-data change.
- **Pass B full-surface review (2026-07-04): already_full.** The eBay Sell Finances API's entire
  documented surface is exactly 8 endpoints across 4 resources; this bundle's 4 streams already
  cover every syncable record shape. The remaining 4 endpoints (`getTransactionSummary`,
  `getPayoutSummary`, `getPayout`, `getTransfer`) are aggregate-summary objects or single-object-
  by-id lookups that duplicate an already-covered stream's record shape — see `api_surface.json`'s
  `excluded` entries for the specific reason per endpoint. The API has no write/mutation method at
  all in this resource group, so `capabilities.write` stays `false` with no `writes.json`.
