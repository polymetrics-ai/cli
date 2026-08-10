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
  api_key (secret)

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

COMMAND SURFACE
  Run UpPromote's declared streams and reverse-ETL actions.
  Usage: pm uppromote <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    add referral adjustment apply - Plan and execute the add referral adjustment reverse-ETL action [intent=reverse_etl availability=implemented write=add_referral_adjustment]; approval: requires plan, preview, approval, and execute; risk: adds a positive or negative commission adjustment to an existing referral, directly changing affiliate payout amounts; no approval required; flags: --adjustment (required), --id (required)
    affiliates list - Run the affiliates ETL stream [intent=etl availability=implemented stream=affiliates]
    api get api v2 affiliates - Documented GET /api/v2/affiliates (not implemented) [intent=direct_read availability=not_implemented operation=uppromote.get.api-v2-affiliates]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 affiliates id - Documented GET /api/v2/affiliates/{id} (not implemented) [intent=direct_read availability=not_implemented operation=uppromote.get.api-v2-affiliates-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 payments id - Documented GET /api/v2/payments/{id} (not implemented) [intent=direct_read availability=not_implemented operation=uppromote.get.api-v2-payments-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 payments total-paid - Documented GET /api/v2/payments/total-paid (not implemented) [intent=direct_read availability=not_implemented operation=uppromote.get.api-v2-payments-total-paid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 programs id - Documented GET /api/v2/programs/{id} (not implemented) [intent=direct_read availability=not_implemented operation=uppromote.get.api-v2-programs-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 programs id excluded - Documented GET /api/v2/programs/{id}/excluded (not implemented) [intent=direct_read availability=not_implemented operation=uppromote.get.api-v2-programs-id-excluded]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 referrals id - Documented GET /api/v2/referrals/{id} (not implemented) [intent=direct_read availability=not_implemented operation=uppromote.get.api-v2-referrals-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 webhook-subscriptions - Documented GET /api/v2/webhook-subscriptions (not implemented) [intent=direct_read availability=not_implemented operation=uppromote.get.api-v2-webhook-subscriptions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post api affiliates - Documented POST /api/affiliates (not implemented) [intent=direct_write availability=not_implemented operation=uppromote.post.api-affiliates]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    approve deny affiliate apply - Plan and execute the approve deny affiliate reverse-ETL action [intent=reverse_etl availability=implemented write=approve_deny_affiliate]; approval: requires plan, preview, approval, and execute; risk: approves or denies a pending affiliate application; low-risk, no approval required; flags: --affiliate_email (required), --status (required)
    approve deny referral apply - Plan and execute the approve deny referral reverse-ETL action [intent=reverse_etl availability=implemented write=approve_deny_referral]; approval: requires plan, preview, approval, and execute; risk: approves or denies a pending referral, affecting affiliate payout eligibility; no approval required; flags: --id (required), --status (required)
    assign coupon to affiliate apply - Plan and execute the assign coupon to affiliate reverse-ETL action [intent=reverse_etl availability=implemented write=assign_coupon_to_affiliate]; approval: requires plan, preview, approval, and execute; risk: assigns a discount coupon code to an affiliate for referral tracking; low-risk, no approval required; flags: --affiliate_email (required)
    connect customer to affiliate apply - Plan and execute the connect customer to affiliate reverse-ETL action [intent=reverse_etl availability=implemented write=connect_customer_to_affiliate]; approval: requires plan, preview, approval, and execute; risk: links a Shopify customer email to an affiliate for future referral attribution; low-risk, no approval required; flags: --affiliate_email (required), --customer_email (required)
    coupons list - Run the coupons ETL stream [intent=etl availability=implemented stream=coupons]
    create affiliate apply - Plan and execute the create affiliate reverse-ETL action [intent=reverse_etl availability=implemented write=create_affiliate]; approval: requires plan, preview, approval, and execute; risk: creates a new affiliate account; low-risk, no approval required (UpPromote caps this at 150 affiliates/day per its own API); flags: --email (required)
    create referral apply - Plan and execute the create referral reverse-ETL action [intent=reverse_etl availability=implemented write=create_referral]; approval: requires plan, preview, approval, and execute; risk: creates a manual commission-bearing referral for an affiliate, either tied to a Shopify order or as a fixed amount; affects payout totals, no approval required; flags: --affiliate_email (required), --type (required)
    delete webhook subscription apply - Plan and execute the delete webhook subscription reverse-ETL action [intent=reverse_etl availability=implemented write=delete_webhook_subscription]; approval: requires plan, preview, approval, and execute; risk: removes a webhook subscription; the external endpoint stops receiving that event type, no approval required; flags: --event (required)
    mark as paid manual payment apply - Plan and execute the mark as paid manual payment reverse-ETL action [intent=reverse_etl availability=implemented write=mark_as_paid_manual_payment]; approval: requires plan, preview, approval, and execute; risk: marks approved referrals as manually paid outside UpPromote's own payout processing; affects financial records, no approval required; flags: --affiliate_email (required)
    move affiliate to program apply - Plan and execute the move affiliate to program reverse-ETL action [intent=reverse_etl availability=implemented write=move_affiliate_to_program]; approval: requires plan, preview, approval, and execute; risk: reassigns an affiliate to a different commission program, changing future commission rules; no approval required; flags: --affiliate_email (required), --program_id (required)
    paid payments list - Run the paid payments ETL stream [intent=etl availability=implemented stream=paid_payments]
    programs list - Run the programs ETL stream [intent=etl availability=implemented stream=programs]
    referrals list - Run the referrals ETL stream [intent=etl availability=implemented stream=referrals]
    set upline affiliate apply - Plan and execute the set upline affiliate reverse-ETL action [intent=reverse_etl availability=implemented write=set_upline_affiliate]; approval: requires plan, preview, approval, and execute; risk: sets the referring (upline) affiliate for a downline affiliate, affecting multi-tier commission attribution; no approval required; flags: --affiliate_email (required), --upline_affiliate_email (required)
    subscribe webhook event apply - Plan and execute the subscribe webhook event reverse-ETL action [intent=reverse_etl availability=implemented write=subscribe_webhook_event]; approval: requires plan, preview, approval, and execute; risk: registers a new outbound webhook subscription that will deliver event payloads to an external URL; low-risk, no approval required; flags: --event (required), --target_url (required)
    unpaid payments list - Run the unpaid payments ETL stream [intent=etl availability=implemented stream=unpaid_payments]
    update webhook subscription apply - Plan and execute the update webhook subscription reverse-ETL action [intent=reverse_etl availability=implemented write=update_webhook_subscription]; approval: requires plan, preview, approval, and execute; risk: updates an existing webhook subscription's target URL; low-risk, no approval required; flags: --event (required), --target_url (required)

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
