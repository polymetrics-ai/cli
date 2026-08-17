---
name: pm-uppromote
description: UpPromote connector knowledge and safe action guide.
---

# pm-uppromote

## Purpose

Reads affiliates, programs, coupons, referrals, and payments from the UpPromote API, and writes affiliate/referral/coupon/payment/webhook-subscription lifecycle mutations.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- start_date
- api_key (secret)

## ETL Streams

- affiliates:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), email(), id(), status()
- programs:
  - primary key: id
  - fields: commission_amount(), commission_type(), created_at(), description(), exclude_product_tax(), exclude_self_referral(), exclude_shipping(), exclude_shipping_tax(), exclude_tip(), id(), is_default(), name(), payment_default(), payment_methods(), rule(), status()
- coupons:
  - primary key: id
  - cursor: created_at
  - fields: affiliate_email(), affiliate_id(), coupon(), created_at(), description(), id()
- referrals:
  - primary key: id
  - fields: commission(), commission_adjustment(), customer_id(), id(), order_id(), order_number(), quantity(), status(), total_sales(), tracking_type()
- unpaid_payments:
  - primary key: affiliate_id
  - fields: affiliate_email(), affiliate_id(), payment_method(), total_commission(), total_products(), total_referrals(), total_sales()
- paid_payments:
  - primary key: payment_id
  - cursor: processed_at
  - fields: affiliate_email(), affiliate_id(), payment_id(), payment_method(), processed_at(), status(), total_processed(), total_referrals()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_affiliate:
  - endpoint: POST api/v2/affiliates
  - risk: creates a new affiliate account; low-risk, no approval required (UpPromote caps this at 150 affiliates/day per its own API)
- approve_deny_affiliate:
  - endpoint: POST api/v2/affiliate/active
  - risk: approves or denies a pending affiliate application; low-risk, no approval required
- set_upline_affiliate:
  - endpoint: POST api/v2/affiliate/set-upline
  - risk: sets the referring (upline) affiliate for a downline affiliate, affecting multi-tier commission attribution; no approval required
- move_affiliate_to_program:
  - endpoint: POST api/v2/affiliate/move-affiliate-to-program
  - risk: reassigns an affiliate to a different commission program, changing future commission rules; no approval required
- connect_customer_to_affiliate:
  - endpoint: POST api/v2/affiliate/create-connect-customer
  - risk: links a Shopify customer email to an affiliate for future referral attribution; low-risk, no approval required
- assign_coupon_to_affiliate:
  - endpoint: POST api/v2/coupons/assign
  - risk: assigns a discount coupon code to an affiliate for referral tracking; low-risk, no approval required
- create_referral:
  - endpoint: POST api/v2/referrals
  - risk: creates a manual commission-bearing referral for an affiliate, either tied to a Shopify order or as a fixed amount; affects payout totals, no approval required
- approve_deny_referral:
  - endpoint: POST api/v2/referral/{{ record.id }}/status
  - required fields: id
  - optional fields: status
  - risk: approves or denies a pending referral, affecting affiliate payout eligibility; no approval required
- add_referral_adjustment:
  - endpoint: POST api/v2/referral/{{ record.id }}/adjustment
  - required fields: id
  - optional fields: adjustment
  - risk: adds a positive or negative commission adjustment to an existing referral, directly changing affiliate payout amounts; no approval required
- mark_as_paid_manual_payment:
  - endpoint: POST api/v2/payments/mark-as-paid
  - risk: marks approved referrals as manually paid outside UpPromote's own payout processing; affects financial records, no approval required
- subscribe_webhook_event:
  - endpoint: POST api/v2/webhook-subscriptions
  - risk: registers a new outbound webhook subscription that will deliver event payloads to an external URL; low-risk, no approval required
- update_webhook_subscription:
  - endpoint: PUT api/v2/webhook-subscriptions
  - risk: updates an existing webhook subscription's target URL; low-risk, no approval required
- delete_webhook_subscription:
  - endpoint: DELETE api/v2/webhook-subscriptions
  - optional fields: event
  - risk: removes a webhook subscription; the external endpoint stops receiving that event type, no approval required

## Security

- read risk: external UpPromote API read of affiliate, program, coupon, referral, and payment data
- write risk: external mutation of UpPromote affiliates, referrals, coupons, payments, and webhook subscriptions; no destructive deletes are modeled
- approval: none required; every modeled write is a create/approve/assign/mark-paid style mutation, not a destructive delete
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect uppromote
```

### Inspect as structured JSON

```bash
pm connectors inspect uppromote --json
```

## Agent Rules

- Run pm connectors inspect uppromote before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
