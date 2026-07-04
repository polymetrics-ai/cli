# Overview

UpPromote is a Shopify affiliate-marketing platform. The original migration read affiliates from a
single legacy-shaped endpoint (`GET {base_url}/api/affiliates`), ported from the hand-written
`internal/connectors/uppromote` legacy package at capability parity. This revision is a Pass B
full-surface expansion against UpPromote's REAL, currently-published API v2 reference
(`https://aff-api.uppromote.com/docs/v2`, an Apidog-hosted OpenAPI spec — the legacy connector's own
`docs_url`, `https://uppromote.com/api`, redirects to marketing copy with no endpoint reference of
its own): 5 new read streams and 13 new write actions covering programs, coupons, referrals,
payments, and the full affiliate/referral/payment/webhook-subscription lifecycle. The legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Requires one secret: `api_key` (UpPromote API key), sent as a Bearer token on every request via
`streams.json` `base.auth`'s `bearer` mode — matching legacy's `connsdk.Bearer(apiKey)` exactly.
`base_url` defaults to `https://api.uppromote.com` (legacy's `defaultBaseURL`), overridable for
tests or proxies. **This auth shape is carried forward unchanged for every Pass B addition** — see
Known limits for why (the real API's own documented auth header format differs).

## Streams notes

`affiliates` (`GET api/affiliates`, records at `affiliates`) is the original parity stream: no
pagination — legacy's `Read` issues a single unpaginated request and emits every record from the one
response, so this stream declares no `pagination` block either (parity, not an omission).

The optional `start_date` config value is sent as the `start_date` query parameter only when set,
using the opt-in optional-query object dialect (`{"template": "{{ config.start_date }}",
"omit_when_absent": true}`) — matching legacy's own conditional `query.Set("start_date", start)`
(only called `if start := strings.TrimSpace(req.Config.Config["start_date"]); start != ""`).
`start_date` is a plain static passthrough config value, not a stateful incremental cursor: legacy
never tracks or advances a read cursor across syncs, so this bundle deliberately declares no
`incremental` block on this stream (declaring one would add cursor-based `incremental_append` sync
modes legacy never supported — an accepted-input-behavior change the migration meta-rule forbids).
The schema still declares `x-cursor-field: created_at` for manifest-surface parity with legacy's
`CursorFields: []string{"created_at"}` (`uppromote.go:128`) — this is purely descriptive metadata;
sync-mode derivation is gated on the stream's `incremental` block, not on `x-cursor-field` alone, so
no new sync mode is introduced by declaring it. `start_date` applies only to this stream, not to any
Pass B addition below (none of the new streams accept a caller-supplied filter in this revision).

**Pass B additions** (new in this revision, all real `api/v2/...` paths per the Apidog reference):

- `programs` (`GET api/v2/programs`, records at `data`) — commission programs (rate/rule/payment
  method configuration); `page_number` pagination (`page`/`per_page`, default page size 10, matching
  the real API's own documented default).
- `coupons` (`GET api/v2/coupons`, records at `data`) — affiliate-assigned discount coupons; same
  pagination shape.
- `referrals` (`GET api/v2/referrals`, records at `data`) — individual commission-bearing referral
  events (order-linked or manual); same pagination shape.
- `unpaid_payments` (`GET api/v2/payments/unpaid`, records at `data`) — per-affiliate aggregates of
  approved-but-not-yet-paid commission; same pagination shape.
- `paid_payments` (`GET api/v2/payments/paid`, records at `data`) — payout history; same pagination
  shape.

All five Pass B streams share the real API's envelope (`{status, message, data: [...]}` ), which
differs from the legacy `affiliates` stream's bare `{"affiliates": [...]}` shape — each stream's
`records.path` is set accordingly (`data` vs `affiliates`).

## Write actions & risks

`capabilities.write` flips to `true` in this revision (legacy shipped none). 13 actions, all newly
added, none destructive-delete (UpPromote's v2 API exposes no delete on affiliates, referrals,
programs, coupons, or payments — only status/lifecycle transitions):

- `create_affiliate` (`POST api/v2/affiliates`) — creates a new affiliate account (capped by
  UpPromote itself at 150/day); no approval required.
- `approve_deny_affiliate` (`POST api/v2/affiliate/active`) — approves/denies a pending affiliate.
- `set_upline_affiliate` (`POST api/v2/affiliate/set-upline`) — sets multi-tier referral attribution.
- `move_affiliate_to_program` (`POST api/v2/affiliate/move-affiliate-to-program`) — reassigns an
  affiliate's commission program.
- `connect_customer_to_affiliate` (`POST api/v2/affiliate/create-connect-customer`) — links a
  Shopify customer email to an affiliate for future referral attribution.
- `assign_coupon_to_affiliate` (`POST api/v2/coupons/assign`) — assigns/creates a coupon code tied
  to an affiliate.
- `create_referral` (`POST api/v2/referrals`) — creates a manual referral, either tied to a Shopify
  `order_id` or as a `fixed_amount`; affects payout totals.
- `approve_deny_referral` (`POST api/v2/referral/{id}/status`) — approves/denies a pending referral.
- `add_referral_adjustment` (`POST api/v2/referral/{id}/adjustment`) — adds a positive or negative
  commission adjustment to an existing referral.
- `mark_as_paid_manual_payment` (`POST api/v2/payments/mark-as-paid`) — records a manual payout
  outside UpPromote's own payment processing.
- `subscribe_webhook_event` / `update_webhook_subscription` / `delete_webhook_subscription`
  (`POST`/`PUT`/`DELETE api/v2/webhook-subscriptions`) — full webhook-subscription lifecycle for the
  9 documented event types (`referral.*`, `affiliate.*`, `payment.paid`). `delete_webhook_subscription`
  is `kind: delete` with `missing_ok_status: [404]` (idempotent) even though it is a configuration
  removal, not a data-record delete — no destructive data loss results.

None of the 13 actions require approval per this bundle's risk framing: every one is a create,
status-transition, or configuration-registration mutation, not an irreversible destructive delete.

## Known limits

- **The `affiliates` stream's path/auth/envelope is a pre-existing legacy behavior, not a Pass B
  deviation.** Legacy's hand-written connector was built against `GET {base_url}/api/affiliates`
  with a bare `{"affiliates": [...]}` response and `Authorization: Bearer <api_key>` — none of which
  match UpPromote's real, currently-published API v2 (base `https://aff-api.uppromote.com/api/v2`,
  envelope `{status, message, data: [...]}`, and `Authorization: <api_key>` with **no** `Bearer`
  prefix — see `aff-api.uppromote.com/docs/v2/api-overview-1615961m0`). This bundle preserves the
  `affiliates` stream's legacy-accepted request shape byte-for-byte (the migration meta-rule forbids
  changing accepted-input behavior); it is not a Pass B regression, and correcting it (if ever
  desired) is out of scope for a surface-expansion pass — it would be a behavior-changing repair,
  tracked separately.
- **Every Pass B addition inherits the SAME `bearer` auth mode**, because `HTTPBase.Auth` is a
  single base-level list shared by every stream/write in a bundle — there is no per-stream auth
  override in this dialect. Against UpPromote's real API (which expects the raw key with no `Bearer`
  prefix), every request this bundle sends — old and new alike — currently includes an extra
  `Bearer ` prefix UpPromote's real API does not document expecting. Operators integrating against
  the real live API today need a thin proxy in front of `base_url` that strips the `Bearer ` prefix
  until the engine grows a per-stream/per-action auth override or this bundle's auth mode is
  deliberately migrated as a tracked, reviewed behavior change (not a silent Pass B side effect).
  This is documented here rather than "fixed" because fixing it would alter the `affiliates` stream's
  already-accepted request shape, which Pass B's scope explicitly forbids.
- **No incremental sync on any Pass B stream.** None of the 5 new streams accept a caller-supplied
  date filter in this revision (the real API's `from_date`/`to_date` query params exist per-endpoint
  but are not wired here) — every read re-fetches the full first page. A future capability-expansion
  pass could wire `from_date`/`to_date` using the same optional-query dialect `start_date` already
  uses on `affiliates`.
- **Webhook-subscription listing (`GET api/v2/webhook-subscriptions`) is excluded** (see
  `api_surface.json`) — the subscribe/update/delete writes already round-trip the full subscription
  lifecycle for a small, operator-managed set; a dedicated list stream was judged low read value for
  this pass.
- Full UpPromote API surface beyond what is enumerated in `api_surface.json` (single-object detail
  endpoints duplicating list-stream shapes, the total-paid-payments scalar aggregate, and
  program-scoped excluded-products sub-object) is out of scope — see `api_surface.json`'s `excluded`
  entries for the specific reason on each.
