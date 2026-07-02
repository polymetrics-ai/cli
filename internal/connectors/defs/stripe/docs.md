# Overview

Stripe is the reference declarative-HTTP golden migration (wave0 `wave:F`). It reads Stripe
customers, charges, invoices, subscriptions, and products, and writes approved reverse-ETL
customer actions through the Stripe REST API. This bundle is engine-vs-legacy parity-tested
against `internal/connectors/stripe` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Stripe secret API key (`sk_...`) via the `client_secret` secret; it is used only for
Bearer auth (`Authorization: Bearer <client_secret>`) and is never logged. An optional `account_id`
config value is sent as the `Stripe-Account` header for Stripe Connect accounts; when unset, the
header is omitted entirely (not sent empty).

## Streams notes

All 5 streams (`customers`, `charges`, `invoices`, `subscriptions`, `products`) share the same
shape: `GET` against the Stripe list endpoint, records at `data`, primary key `["id"]`, incremental
cursor field `created`. Pagination follows Stripe's `starting_after`/`has_more` id-cursor
convention (`pagination.type: cursor` with `last_record_field: id` and `stop_path: has_more`): the
next page's `starting_after` is the `id` of the last record on the current page, and pagination
stops when `has_more` is falsy or a page yields no records. Every request sends `limit=100`
(matches legacy's default `page_size`). Incremental reads send `created[gte]` as a Unix-seconds
value (`param_format: unix_seconds`), computed either from the sync's persisted cursor or, on a
fresh sync, from the RFC3339 `start_date` config value — identical to legacy
`incrementalLowerBound`/`formatParam`.

## Write actions & risks

`create_customer` (`POST /customers`, form-encoded body) creates a Stripe customer from
`email`/`name`/`description`/`phone`. `update_customer` (`POST /customers/{id}`, form-encoded
body, `id` carried in the path) updates an existing customer's mutable fields. Both are external
mutations requiring reverse-ETL plan approval before execution, matching legacy
`stripe/write.go`'s allow-list semantics exactly (method, path, and form field set).

Documented parity deviation: legacy's `create_customer` validation requires "email OR name"
present (a named-field OR-rule). The engine's draft-07 subset validator has no `anyOf`/`oneOf`
keyword, so this bundle approximates the rule with `minProperties: 1` over the four optional
customer fields — strictly more permissive than legacy (e.g. a record with only `phone` set would
pass here but fail legacy), never stricter, and never diverges for any record legacy itself would
accept. See `docs/migration/conventions.md`'s parity-deviation ledger.

## Known limits

- Full Stripe API surface (payment intents, refunds, payouts, disputes, webhooks) is out of scope
  for wave0; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries. Only the 5 legacy-parity streams and 2 legacy-parity write
  actions are implemented.
- This bundle's own `fixtures/streams/customers/*.json` (used by `TestConformance/stripe` and
  `connectorgen validate`, not by the engine-vs-legacy parity suite) represent the `created` cursor
  field as an RFC3339 string rather than Stripe's real Unix-seconds integer wire format. This is a
  fixture-authoring accommodation for `internal/connectors/conformance`'s `cursor_advances` check,
  which recognizes a cursor value only via a Go `string` type assertion and then parses it as
  RFC3339 for `unix_seconds` formatting; it does not affect engine-vs-legacy record parity (the
  schema type is declared as `["integer","string"]` and the read path performs no type coercion).
