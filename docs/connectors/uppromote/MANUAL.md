# pm connectors inspect uppromote

```text
NAME
  pm connectors inspect uppromote - UpPromote connector manual

SYNOPSIS
  pm connectors inspect uppromote
  pm connectors inspect uppromote --json
  pm credentials add <name> --connector uppromote [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads affiliates, programs, coupons, referrals, and payments from the UpPromote API, and writes affiliate/referral/coupon/payment/webhook-subscription lifecycle mutations.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  start_date
  api_key (secret) (required)

ETL STREAMS
  affiliates:
    primary key: id
    cursor: created_at
    fields: created_at(string), email(string), id(string), status(string)
  programs:
    primary key: id
    fields: commission_amount(string), commission_type(string), created_at(string), description(string), exclude_product_tax(boolean), exclude_self_referral(boolean), exclude_shipping(boolean), exclude_shipping_tax(boolean), exclude_tip(boolean), id(integer), is_default(string), name(string), payment_default(string), payment_methods(array), rule(string), status(string)
  coupons:
    primary key: id
    cursor: created_at
    fields: affiliate_email(string), affiliate_id(integer), coupon(string), created_at(string), description(string), id(integer)
  referrals:
    primary key: id
    fields: commission(string), commission_adjustment(string), customer_id(string), id(integer), order_id(integer), order_number(integer), quantity(integer), status(string), total_sales(string), tracking_type(string)
  unpaid_payments:
    primary key: affiliate_id
    fields: affiliate_email(string), affiliate_id(integer), payment_method(string), total_commission(number), total_products(integer), total_referrals(integer), total_sales(number)
  paid_payments:
    primary key: payment_id
    cursor: processed_at
    fields: affiliate_email(string), affiliate_id(integer), payment_id(integer), payment_method(string), processed_at(string), status(string), total_processed(number), total_referrals(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_affiliate:
    endpoint: POST api/v2/affiliates
    required fields: email
    risk: creates a new affiliate account; low-risk, no approval required (UpPromote caps this at 150 affiliates/day per its own API)
  approve_deny_affiliate:
    endpoint: POST api/v2/affiliate/active
    required fields: affiliate_email, status
    risk: approves or denies a pending affiliate application; low-risk, no approval required
  set_upline_affiliate:
    endpoint: POST api/v2/affiliate/set-upline
    required fields: affiliate_email, upline_affiliate_email
    risk: sets the referring (upline) affiliate for a downline affiliate, affecting multi-tier commission attribution; no approval required
  move_affiliate_to_program:
    endpoint: POST api/v2/affiliate/move-affiliate-to-program
    required fields: affiliate_email, program_id
    risk: reassigns an affiliate to a different commission program, changing future commission rules; no approval required
  connect_customer_to_affiliate:
    endpoint: POST api/v2/affiliate/create-connect-customer
    required fields: affiliate_email, customer_email
    risk: links a Shopify customer email to an affiliate for future referral attribution; low-risk, no approval required
  assign_coupon_to_affiliate:
    endpoint: POST api/v2/coupons/assign
    required fields: affiliate_email
    risk: assigns a discount coupon code to an affiliate for referral tracking; low-risk, no approval required
  create_referral:
    endpoint: POST api/v2/referrals
    required fields: type, affiliate_email
    risk: creates a manual commission-bearing referral for an affiliate, either tied to a Shopify order or as a fixed amount; affects payout totals, no approval required
  approve_deny_referral:
    endpoint: POST api/v2/referral/{{ record.id }}/status
    required fields: id, status
    risk: approves or denies a pending referral, affecting affiliate payout eligibility; no approval required
  add_referral_adjustment:
    endpoint: POST api/v2/referral/{{ record.id }}/adjustment
    required fields: id, adjustment
    risk: adds a positive or negative commission adjustment to an existing referral, directly changing affiliate payout amounts; no approval required
  mark_as_paid_manual_payment:
    endpoint: POST api/v2/payments/mark-as-paid
    required fields: affiliate_email
    risk: marks approved referrals as manually paid outside UpPromote's own payout processing; affects financial records, no approval required
  subscribe_webhook_event:
    endpoint: POST api/v2/webhook-subscriptions
    required fields: target_url, event
    risk: registers a new outbound webhook subscription that will deliver event payloads to an external URL; low-risk, no approval required
  update_webhook_subscription:
    endpoint: PUT api/v2/webhook-subscriptions
    required fields: target_url, event
    risk: updates an existing webhook subscription's target URL; low-risk, no approval required
  delete_webhook_subscription:
    endpoint: DELETE api/v2/webhook-subscriptions
    required fields: event
    risk: removes a webhook subscription; the external endpoint stops receiving that event type, no approval required

SECURITY
  read risk: external UpPromote API read of affiliate, program, coupon, referral, and payment data
  write risk: external mutation of UpPromote affiliates, referrals, coupons, payments, and webhook subscriptions; no destructive deletes are modeled
  approval: none required; every modeled write is a create/approve/assign/mark-paid style mutation, not a destructive delete
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect uppromote

  # Inspect as structured JSON
  pm connectors inspect uppromote --json

AGENT WORKFLOW
  - Run pm connectors inspect uppromote before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
