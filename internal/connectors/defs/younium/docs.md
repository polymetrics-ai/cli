# Overview

Younium is a subscription billing platform. This bundle reads Younium accounts, subscriptions,
invoices, products, payment terms, currencies, and webhooks through the Younium REST API v2.1
(`GET {base_url}/Accounts|Subscriptions|Invoices|Products|PaymentTerms|Currency|Webhooks`), and
writes account/subscription/invoice lifecycle mutations. It migrates
`internal/connectors/younium` (the hand-written connector, which was read-only with only the 3
legacy-parity streams); the legacy package stays registered and unchanged until wave6's registry
flip. Pass B full-surface expansion (`docs/migration/conventions.md`) added the 4 new streams and
5 write actions and flipped `capabilities.write` to `true` — see `api_surface.json` for the full
documented Younium API v2.1 surface (29 resource categories) and why everything not covered here
is excluded.

## Auth setup

Provide `username` (config) and `password` (secret) for HTTP Basic auth
(`Authorization: Basic base64(username:password)`), matching legacy's `connsdk.Basic(username,
password)`. An optional `legal_entity` config value is sent as the `X-Younium-Legal-Entity` request
header when set; when unset, the header is omitted entirely (not sent empty) — `legal_entity` is
declared in `spec.json` but not in `required[]`, so the engine's conditional-header omission
semantics apply (`docs/migration/conventions.md` §3).

## Streams notes

The 3 legacy-parity streams (`accounts`, `subscriptions`, `invoices`) share the same shape: `GET`
against the Younium list endpoint (`/Accounts`, `/Subscriptions`, `/Invoices`), records at `data`,
primary key `["id"]`, cursor field `updated_at`. No pagination is declared for these 3 — legacy
issues a single unpaginated request per stream and emits every record in the response's `data`
array, so this bundle's `streams.json` omits any `pagination` block (defaulting to `none`) for
them, matching legacy exactly.

Every legacy-parity stream declares `projection: passthrough`. Legacy's `mapRecord` first copies
every raw response field verbatim (`for k, v := range in { out[k] = v }`) and only THEN overlays
the 4 derived aliases (`id`, `name`, `account_id`, `updated_at`) on top — the real emitted record
is the raw field set plus those aliases, not just the 4 aliases alone. `"schema"` (default)
projection would restrict every stream to only the declared properties, silently dropping every
other raw field (`accountId`, `accountName`, `invoiceId`, `invoiceNumber`, `updated`, etc.) that
legacy's copy-first loop preserves under its own raw name. `passthrough` reproduces that copy-first
behavior; `computed_fields` then overlay the renamed aliases on top of the passed-through raw
fields, matching legacy's overlay-after-copy order exactly.

`computed_fields` mirror legacy's first-non-null fallback chains with `coalesce`: accounts use
`id`/`accountId` and `name`/`accountName`; invoices use `id`/`invoiceId` and
`invoiceNumber`/`number`/`name`; all 3 legacy streams use `accountId`/`account_id` for
`account_id` and `updated`/`updatedAt`/`updated_at` for `updated_at`.

**Pass B new streams** (`products`, `payment_terms`, `currencies`, `webhooks`) also use
`projection: passthrough` for the same undisguised-raw-field-preservation reasoning, with a
minimal `id`/`name` computed-field overlay (no legacy behavior to match here, since legacy never
read these resources — this is new coverage, so the overlay is authored fresh, not ported). Their
real wire ids: `products.id`/`products.modified` (renamed to `updated_at`), `payment_terms.id`/
`payment_terms.name`, `currencies.id`/`currencies.code` (renamed to `name`, matching the wire's own
ISO-4217 currency code field), `webhooks.id`/`webhooks.url` (renamed to `name`, since Younium
webhook records identify themselves by target URL rather than a separate display name). `products`,
`payment_terms`, and `currencies` paginate via `offset_limit` (`Take`/`Skip` query params, matching
the Younium API's own parameter casing, `page_size: 100`) since Younium's own docs show these as
plain list endpoints with no documented upper bound; `webhooks` is left unpaginated (Younium
tenants typically register a small, bounded number of webhook subscriptions).

## Write actions & risks

5 write actions, all requiring approval; `cancel_subscription` and `cancel_invoice` additionally
require explicit destructive confirmation (`confirm: "destructive"`):

- **`create_account`** (`POST /Accounts`) — creates a new billing account. Requires `name` and
  `currency`.
- **`update_account`** (`PATCH /Accounts/{id}`) — mutates an existing account's billing/contact/
  tax metadata.
- **`cancel_subscription`** (`POST /Subscriptions/cancel/{id}`) — irreversibly schedules or
  immediately cancels an active subscription (`cancellationMode`: Younium's `CancelOrderMode` enum,
  e.g. `EndOfTerm`/`Immediate`), ending future billing. Destructive.
- **`post_invoice`** (`POST /Invoices/{id}/post`) — finalizes a draft invoice, making it official/
  sendable to the customer. No request body (`body_type: none`); the invoice `id` is fully carried
  by the path.
- **`cancel_invoice`** (`POST /Invoices/{id}/cancel`) — irreversibly cancels a posted invoice.
  Destructive. No request body.

The riskier account-deletion (`DELETE /Accounts/{accountId}`) and every other invoice/subscription
lifecycle mutation (renew, activate, revert, credit memos, Stripe payment processing, on-account
invoices) are excluded this pass — see `api_surface.json` for the full per-endpoint reasoning; the
5 actions above were chosen as the highest-value, cleanly-single-request mutations the dialect can
express without a compound multi-section body (subscription creation/mid-term change need an
order-line-shaped nested body no `record_schema` here attempts to validate).

## Known limits

- Legacy's `mapRecord` still assigns alias keys with `nil` when every candidate field in a
  fallback chain is absent. The declarative engine can now express the fallback chain itself, but
  cannot force emission of a null-valued computed key when every candidate path is absent. This is
  the same per-field emit-null limitation tracked in the shared fidelity follow-ups.
- Full Younium API surface (~23 remaining resource categories — Bookings, ChartOfAccounts,
  Orders, Payments, Quotes, Usage, Users, and more) is out of scope for this pass; see
  `api_surface.json`'s per-endpoint `excluded` entries for the specific reason each was left out
  (breadth-vs-cost triage, PCI-adjacent payment-detail scope, destructive-admin deletes, binary
  attachment downloads, compound-body creates the dialect cannot validate).
- `create_account`/`update_account`/`cancel_subscription`/`post_invoice`/`cancel_invoice` are new
  Pass B write actions with no legacy Go counterpart to match against (legacy was read-only) — their
  `record_schema` and risk classification are authored fresh from the live Younium API docs, not
  ported from an existing implementation.
