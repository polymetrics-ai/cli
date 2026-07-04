# Overview

ChargeDesk is a wave2 fan-out migration, expanded in Pass B to the full documented API surface.
This bundle reads ChargeDesk charges, customers, subscriptions, products, activity logs, and
subscription-cancellation logs, and a static catalog of webhook notification types, through the
ChargeDesk REST API, migrating `internal/connectors/chargedesk` (the legacy hand-written connector,
which stays registered and unchanged until wave6's registry flip). It also writes 13 mutations:
customer/charge/webhook/agent create-update-delete lifecycle actions, plus 4 live "gateway methods"
(refund/capture/void a charge, cancel a subscription) that mutate the connected payment gateway
itself, not just ChargeDesk's own records. Legacy is entirely read-only (`Write` stub returning
`connectors.ErrUnsupportedOperation`); every write action here is new capability, not a parity port.

The four legacy streams are field-built from the Go mappers; their schemas intentionally project
only those legacy output fields even when ChargeDesk's current API returns additional attributes.

## Auth setup

ChargeDesk authenticates with HTTP Basic auth
(https://chargedesk.com/api-docs/authentication). Provide the ChargeDesk secret API key via the
`password` secret; by default it is sent as the Basic auth USERNAME with a blank password
(`Authorization: Basic base64(<password>:)`), matching legacy's default scheme. An optional
`username` config value overrides this: when set, `username` is the Basic auth username and
`password` is its Basic auth password instead (`Authorization: Basic base64(<username>:<password>)`).

This is a **dual-auth candidate list**, evaluated first-match-wins (`base.auth`'s declaration
order is load-bearing — see `docs/migration/conventions.md` §3's dual-auth-ordering rule): the
`username`-gated candidate is declared FIRST (`when: "{{ config.username }}"`, true only when
`username` is configured), falling through to the secret-as-username default candidate when
`username` is absent — reproducing legacy's exact `switch { case username != "": ...; case secret
!= "": ...}` precedence (username override wins whenever both are present).

## Streams notes

`charges`/`customers`/`subscriptions`/`products`/`log_activity`/`log_cancellations` share the same
shape: `GET` against the ChargeDesk list endpoint, records at `data`, offset/count pagination
(`pagination.type: offset_limit`, `limit_param: count`, `offset_param: offset`, `page_size: 100` —
matches legacy's `chargedeskDefaultPageSize` and ChargeDesk's own documented `count` param, whose
default is 20 but whose maximum is 500; 100 matches legacy's choice), stopping on a short page
(fewer than 100 records), matching legacy's `len(records) < pageSize` rule exactly (every one of
these envelopes has no `has_more`/next-page-token field — offset/count is the API's real, only
documented pagination shape for all 6). `webhook_notifications` declares `pagination.type: none`
and `records.path: "."`: `/webhooks/notifications` returns a bare JSON array (not an object with a
named field), a static reference catalog of every possible webhook notification type ChargeDesk can
send, with no pagination at all.

`charges`/`customers`/`subscriptions`/`products` all keep the legacy catalog cursor field
`occurred`. `log_activity`/`log_cancellations` both declare `incremental.cursor_field: occurred`
(their own documented timestamp field). None of these 6 streams declare `request_param` or
`client_filtered` — this bundle never sends any incremental filter to the API and never
client-side filters either (matching legacy's own unfiltered-full-walk behavior for the original
4 streams); the bare `cursor_field` declaration exists only so the engine derives
`incremental_append` sync-mode eligibility from the schema's own cursor field. `webhook_notifications`
has no incremental cursor (it is a static, non-time-ordered catalog). Primary keys: `charges` uses
`charge_id`, `customers` uses `customer_id`, `subscriptions` uses `subscription_id`, `products`
uses `product_id`, `webhook_notifications` uses `notification` — matching each resource's real
unique identifier field. `log_activity`/`log_cancellations` declare NO `x-primary-key`: neither
resource's documented response fields include any unique id (they are append-only event logs keyed
only by `occurred` + the referenced `object_id`/`subscription_id`), so these two streams support
only `full_refresh_append`/`incremental_append` sync modes, never a deduped variant — this is
Chargedesk's real API shape, not an authoring gap (§2 sync-mode derivation is schema-driven: no
declared `x-primary-key` means no `*_deduped` mode is offered, correctly).

`spec.json` intentionally does NOT declare `page_size`/`max_pages` as runtime-configurable
properties (unlike legacy, which accepts config overrides for both): `PaginationSpec.PageSize`/
`MaxPages` are read exclusively from `streams.json`'s static `pagination` JSON literal, never from
a `config.*`-templated value (F6, `conventions.md`). See Known limits.

## Write actions & risks

Thirteen write actions were added in Pass B, all gated by approval (`metadata.json`'s
`capabilities.write: true`, `risk.write`). None existed in legacy, which is entirely read-only:

- `create_customer`/`update_customer`/`delete_customer` (`POST /customers`, `POST
  /customers/{id}`, `DELETE /customers/{id}`) — ordinary customer-record lifecycle.
  `delete_customer` is irreversible and, by ChargeDesk's own documented default
  (`delete_all` unset), also deletes all associated charges/tickets.
- `update_charge`/`delete_charge` (`POST /charges/{id}`, `DELETE /charges/{id}`) — updates a
  charge record's stored data (amount/currency/status/customer_id), or deletes the record.
  `create_charge` (a plain, non-gateway record-only create) is deliberately excluded — see
  `api_surface.json`.
- `refund_charge`/`capture_charge`/`void_charge` (`POST /gateway/charges/{id}/{refund,capture,void}`)
  — these are ChargeDesk's "gateway methods": each mutates the charge on the ORIGINATING PAYMENT
  GATEWAY itself (Stripe, Braintree, etc.), not just ChargeDesk's own record, and each may fail per
  the connected gateway's own rules. `refund_charge` is irreversible.
  `void_charge` doubles as "cancel a payment request" per ChargeDesk's own docs.
- `cancel_subscription` (`POST /gateway/subscriptions/{id}/cancel`) — another gateway method:
  irreversibly cancels future recurring charges for a subscription on the connected gateway.
- `create_webhook`/`delete_webhook` (`POST /webhooks`, `DELETE /webhooks/{id}`) — creates or
  removes an outbound webhook subscription; `create_webhook`'s `notifications`/`all` params
  select which of the `webhook_notifications` stream's catalog entries to subscribe to.
- `create_agent`/`delete_agent` (`POST /agents`, `DELETE /agents/{email}`) — invites (or updates
  the role of) a support agent with ChargeDesk account access, or revokes it. `delete_agent` is
  addressed by email (ChargeDesk has no separate numeric agent id in its documented API).

Every other documented ChargeDesk mutation is excluded — see `api_surface.json` for the specific
category+reason per endpoint: the highest-risk live gateway methods (`gateway/products/charge`
— directly charges a customer's card — and `gateway/subscriptions/{id}/plans` — changes a live
billing plan) are deliberately excluded pending real demand, as are record-only (non-gateway)
create/update paths for charges/subscriptions/products that this connector's write surface does not
need to duplicate.

## Known limits

- Additional current-API fields on the four legacy streams are intentionally not projected; legacy
  emitted field-built records and remains the fidelity target for this pass.
- `page_size`/`max_pages` runtime overrides are not exposed (see Streams notes above) — every
  read uses the fixed `page_size: 100`/unbounded-pages shape baked into `streams.json`. This never
  changes any single emitted record's DATA, only how many requests a sync issues and at what page
  size — parity-deviation ledger candidate, ACCEPTABLE under the meta-rule.
- `log_activity`/`log_cancellations` have no primary key (ChargeDesk's own API documents none for
  either resource) — see Streams notes above for the sync-mode consequence.
- `/charges/{id}/items` (per-charge line-item/tax breakdown), `/customers/grouped` (an
  identity-search lookup, not a list), and `/charges/preview` (a stateless tax/total calculator)
  are excluded — none of the three is a list resource with an independent primary key to project as
  a stream. See `api_surface.json` for every other excluded endpoint and its specific reason.
