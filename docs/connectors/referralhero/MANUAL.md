# pm connectors inspect referralhero

```text
NAME
  pm connectors inspect referralhero - ReferralHero connector manual

SYNOPSIS
  pm connectors inspect referralhero
  pm connectors inspect referralhero --json
  pm credentials add <name> --connector referralhero [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads ReferralHero lists, subscribers, referrals, rewards, coupon groups, and campaign-scoped subscriber resources, and performs approved ReferralHero API v2 mutations.

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
  coupon_group_id
  list_uuid
  subscriber_email
  subscriber_id
  subscriber_mwr
  subscriber_name
  api_key (secret) (required)

ETL STREAMS
  lists:
    primary key: id
    fields: created_at(string), id(string), name(string), status(string)
  subscribers:
    primary key: id
    cursor: updated_at
    fields: email(string), id(string), name(string), referral_code(string), status(string), updated_at(string)
  referrals:
    primary key: id
    cursor: created_at
    fields: created_at(string), email(string), id(string), status(string), subscriber_id(string)
  rewards:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), updated_at(string)
  list_leaderboard:
    primary key: id
    fields: code(string), created_at(integer), crypto_wallet_address(string), device(string), email(string), extra_field(string), extra_field_2(string), extra_field_3(string), extra_field_4(string), host(string), id(string), last_updated_at(integer), level_2_confirmed_referrals(integer), level_3_confirmed_referrals(integer), name(string), option_field(string), other_identifier_value(string), pending_referrals(integer), people_referred(integer), phone_number(string), points(integer), position(integer), promoted(boolean), promoted_at(integer), referral_link(string), referral_status(string), referral_status_at(integer), referred(boolean), referred_by(object), response(string), risk_level(integer), source(string), stripe_customer_id(string), tags(array), unconfirmed_referrals(integer), universal_link(string), verified(boolean), verified_at(integer), visitors(integer)
  list_bonuses:
    fields: description(string), referrals(integer), title(string)
  subscribers_search_by_name:
    primary key: id
    fields: code(string), created_at(integer), crypto_wallet_address(string), device(string), email(string), extra_field(string), extra_field_2(string), extra_field_3(string), extra_field_4(string), host(string), id(string), last_updated_at(integer), level_2_confirmed_referrals(integer), level_3_confirmed_referrals(integer), name(string), option_field(string), other_identifier_value(string), pending_referrals(integer), people_referred(integer), phone_number(string), points(integer), position(integer), promoted(boolean), promoted_at(integer), referral_link(string), referral_status(string), referral_status_at(integer), referred(boolean), referred_by(object), response(string), risk_level(integer), source(string), stripe_customer_id(string), tags(array), unconfirmed_referrals(integer), universal_link(string), verified(boolean), verified_at(integer), visitors(integer)
  campaign_subscribers:
    primary key: id
    fields: code(string), created_at(integer), crypto_wallet_address(string), device(string), email(string), extra_field(string), extra_field_2(string), extra_field_3(string), extra_field_4(string), host(string), id(string), last_updated_at(integer), level_2_confirmed_referrals(integer), level_3_confirmed_referrals(integer), name(string), option_field(string), other_identifier_value(string), pending_referrals(integer), people_referred(integer), phone_number(string), points(integer), position(integer), promoted(boolean), promoted_at(integer), referral_link(string), referral_status(string), referral_status_at(integer), referred(boolean), referred_by(object), response(string), risk_level(integer), source(string), stripe_customer_id(string), tags(array), unconfirmed_referrals(integer), universal_link(string), verified(boolean), verified_at(integer), visitors(integer)
  subscriber_detail:
    primary key: id
    fields: code(string), created_at(integer), crypto_wallet_address(string), device(string), email(string), extra_field(string), extra_field_2(string), extra_field_3(string), extra_field_4(string), host(string), id(string), last_updated_at(integer), level_2_confirmed_referrals(integer), level_3_confirmed_referrals(integer), name(string), option_field(string), other_identifier_value(string), pending_referrals(integer), people_referred(integer), phone_number(string), points(integer), position(integer), promoted(boolean), promoted_at(integer), referral_link(string), referral_status(string), referral_status_at(integer), referred(boolean), referred_by(object), response(string), risk_level(integer), source(string), stripe_customer_id(string), tags(array), unconfirmed_referrals(integer), universal_link(string), verified(boolean), verified_at(integer), visitors(integer)
  subscriber_by_email:
    primary key: id
    fields: code(string), created_at(integer), crypto_wallet_address(string), device(string), email(string), extra_field(string), extra_field_2(string), extra_field_3(string), extra_field_4(string), host(string), id(string), last_updated_at(integer), level_2_confirmed_referrals(integer), level_3_confirmed_referrals(integer), name(string), option_field(string), other_identifier_value(string), pending_referrals(integer), people_referred(integer), phone_number(string), points(integer), position(integer), promoted(boolean), promoted_at(integer), referral_link(string), referral_status(string), referral_status_at(integer), referred(boolean), referred_by(object), response(string), risk_level(integer), source(string), stripe_customer_id(string), tags(array), unconfirmed_referrals(integer), universal_link(string), verified(boolean), verified_at(integer), visitors(integer)
  subscriber_by_mwr:
    primary key: id
    fields: code(string), created_at(integer), crypto_wallet_address(string), device(string), email(string), extra_field(string), extra_field_2(string), extra_field_3(string), extra_field_4(string), host(string), id(string), last_updated_at(integer), level_2_confirmed_referrals(integer), level_3_confirmed_referrals(integer), name(string), option_field(string), other_identifier_value(string), pending_referrals(integer), people_referred(integer), phone_number(string), points(integer), position(integer), promoted(boolean), promoted_at(integer), referral_link(string), referral_status(string), referral_status_at(integer), referred(boolean), referred_by(object), response(string), risk_level(integer), source(string), stripe_customer_id(string), tags(array), unconfirmed_referrals(integer), universal_link(string), verified(boolean), verified_at(integer), visitors(integer)
  subscriber_referrals:
    primary key: id
    fields: code(string), created_at(integer), crypto_wallet_address(string), device(string), email(string), extra_field(string), extra_field_2(string), extra_field_3(string), extra_field_4(string), host(string), id(string), last_updated_at(integer), level_2_confirmed_referrals(integer), level_3_confirmed_referrals(integer), name(string), option_field(string), other_identifier_value(string), pending_referrals(integer), people_referred(integer), phone_number(string), points(integer), position(integer), promoted(boolean), promoted_at(integer), referral_link(string), referral_status(string), referral_status_at(integer), referred(boolean), referred_by(object), response(string), risk_level(integer), source(string), stripe_customer_id(string), tags(array), unconfirmed_referrals(integer), universal_link(string), verified(boolean), verified_at(integer), visitors(integer)
  subscriber_level_2_all_referrals:
    primary key: id
    fields: code(string), created_at(integer), crypto_wallet_address(string), device(string), email(string), extra_field(string), extra_field_2(string), extra_field_3(string), extra_field_4(string), host(string), id(string), last_updated_at(integer), level_2_confirmed_referrals(integer), level_3_confirmed_referrals(integer), name(string), option_field(string), other_identifier_value(string), pending_referrals(integer), people_referred(integer), phone_number(string), points(integer), position(integer), promoted(boolean), promoted_at(integer), referral_link(string), referral_status(string), referral_status_at(integer), referred(boolean), referred_by(object), response(string), risk_level(integer), source(string), stripe_customer_id(string), tags(array), unconfirmed_referrals(integer), universal_link(string), verified(boolean), verified_at(integer), visitors(integer)
  subscriber_level_3_all_referrals:
    primary key: id
    fields: code(string), created_at(integer), crypto_wallet_address(string), device(string), email(string), extra_field(string), extra_field_2(string), extra_field_3(string), extra_field_4(string), host(string), id(string), last_updated_at(integer), level_2_confirmed_referrals(integer), level_3_confirmed_referrals(integer), name(string), option_field(string), other_identifier_value(string), pending_referrals(integer), people_referred(integer), phone_number(string), points(integer), position(integer), promoted(boolean), promoted_at(integer), referral_link(string), referral_status(string), referral_status_at(integer), referred(boolean), referred_by(object), response(string), risk_level(integer), source(string), stripe_customer_id(string), tags(array), unconfirmed_referrals(integer), universal_link(string), verified(boolean), verified_at(integer), visitors(integer)
  subscriber_level_1_referrals:
    primary key: id
    fields: code(string), created_at(integer), crypto_wallet_address(string), device(string), email(string), extra_field(string), extra_field_2(string), extra_field_3(string), extra_field_4(string), host(string), id(string), last_updated_at(integer), level_2_confirmed_referrals(integer), level_3_confirmed_referrals(integer), name(string), option_field(string), other_identifier_value(string), pending_referrals(integer), people_referred(integer), phone_number(string), points(integer), position(integer), promoted(boolean), promoted_at(integer), referral_link(string), referral_status(string), referral_status_at(integer), referred(boolean), referred_by(object), response(string), risk_level(integer), source(string), stripe_customer_id(string), tags(array), unconfirmed_referrals(integer), universal_link(string), verified(boolean), verified_at(integer), visitors(integer)
  subscriber_level_2_referrals:
    primary key: id
    fields: code(string), created_at(integer), crypto_wallet_address(string), device(string), email(string), extra_field(string), extra_field_2(string), extra_field_3(string), extra_field_4(string), host(string), id(string), last_updated_at(integer), level_2_confirmed_referrals(integer), level_3_confirmed_referrals(integer), name(string), option_field(string), other_identifier_value(string), pending_referrals(integer), people_referred(integer), phone_number(string), points(integer), position(integer), promoted(boolean), promoted_at(integer), referral_link(string), referral_status(string), referral_status_at(integer), referred(boolean), referred_by(object), response(string), risk_level(integer), source(string), stripe_customer_id(string), tags(array), unconfirmed_referrals(integer), universal_link(string), verified(boolean), verified_at(integer), visitors(integer)
  subscriber_level_3_referrals:
    primary key: id
    fields: code(string), created_at(integer), crypto_wallet_address(string), device(string), email(string), extra_field(string), extra_field_2(string), extra_field_3(string), extra_field_4(string), host(string), id(string), last_updated_at(integer), level_2_confirmed_referrals(integer), level_3_confirmed_referrals(integer), name(string), option_field(string), other_identifier_value(string), pending_referrals(integer), people_referred(integer), phone_number(string), points(integer), position(integer), promoted(boolean), promoted_at(integer), referral_link(string), referral_status(string), referral_status_at(integer), referred(boolean), referred_by(object), response(string), risk_level(integer), source(string), stripe_customer_id(string), tags(array), unconfirmed_referrals(integer), universal_link(string), verified(boolean), verified_at(integer), visitors(integer)
  campaign_rewards:
    primary key: id
    fields: coupon_code(string), coupon_group(string), created_at(integer), id(string), image_url(string), name(string), recurring_count(integer), referral(string), referrals(integer), referrals_type(string), sent_date(integer), signup_type(string), status(string), subscriber_email(string), subscriber_id(string), total(string), unlocked_date(integer), value(number)
  subscriber_rewards:
    primary key: id
    fields: coupon_code(string), coupon_group(string), created_at(integer), id(string), image_url(string), name(string), recurring_count(integer), referral(string), referrals(integer), referrals_type(string), sent_date(integer), signup_type(string), status(string), subscriber_email(string), subscriber_id(string), total(string), unlocked_date(integer), value(number)
  coupon_groups:
    primary key: id
    fields: active(boolean), coupons(array), created_at(integer), id(string), name(string)
  coupon_group_coupons:
    primary key: code
    fields: available(boolean), code(string), created_at(integer), email_id(string), sent_at(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_list:
    endpoint: POST /lists
    required fields: website, name
    risk: creates a live ReferralHero campaign/list in the account; external mutation, approval required
  add_subscriber:
    endpoint: POST /lists/{{ record.uuid }}/subscribers
    required fields: uuid, email
    risk: creates or registers a live subscriber in a ReferralHero campaign and may trigger campaign email/referral workflows; approval required
  track_referral_conversion_event:
    endpoint: POST /lists/{{ record.uuid }}/subscribers/track_referral_conversion_event
    required fields: uuid, email
    risk: confirms/unconfirms referral conversion state and may create a referral when a referrer is provided; external mutation, approval required
  confirm_subscriber_by_id:
    endpoint: POST /lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id }}/confirm
    required fields: uuid, subscriber_id
    risk: confirms a verified referral/subscriber conversion in the campaign; external mutation, approval required
  confirm_subscriber_by_identifier:
    endpoint: POST /lists/{{ record.uuid }}/subscribers/confirm
    required fields: uuid, email
    risk: confirms a verified referral/subscriber conversion by unique identifier; external mutation, approval required
  update_subscriber:
    endpoint: POST /lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id }}
    required fields: uuid, subscriber_id
    risk: updates profile, identifier, points, address, or tag fields for a verified subscriber; external mutation, approval required
  add_points:
    endpoint: POST /lists/{{ record.uuid }}/subscribers/add_points
    required fields: uuid, email, points
    risk: adds points to a subscriber, changing contest/reward standings; external mutation, approval required
  add_transaction:
    endpoint: POST /lists/{{ record.uuid }}/subscribers/add_transactions
    required fields: uuid, email, amount
    risk: records a transaction against a subscriber and may affect conversion/reward calculations; external mutation, approval required
  add_bulk_transactions:
    endpoint: POST /lists/{{ record.uuid }}/subscribers/add_bulk_transactions
    required fields: uuid, transactions
    risk: records up to 500 transactions in one call and emails an admin CSV result; high-blast-radius external mutation, approval required
  promote_subscriber:
    endpoint: POST /lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id }}/promote
    required fields: uuid, subscriber_id
    risk: promotes a subscriber into the campaign winners/promoted state; external mutation, approval required
  unlock_promoted_reward:
    endpoint: POST /lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id }}/unlock_promoted_reward
    required fields: uuid, subscriber_id
    risk: unlocks a promoted reward for a subscriber, changing reward fulfillment state; external mutation, approval required
  delete_subscriber:
    endpoint: DELETE /lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id }}
    required fields: uuid, subscriber_id
    risk: permanently deletes a subscriber from a live campaign; destructive external mutation, approval required
  update_reward_status:
    endpoint: POST /lists/{{ record.uuid }}/subscribers/update_reward_status
    required fields: uuid, reward_id, status
    risk: changes fulfillment status for an unlocked reward; external mutation, approval required
  create_coupon_group:
    endpoint: POST /lists/{{ record.uuid }}/coupon_groups
    required fields: uuid, name, coupons, active
    risk: creates a campaign coupon group and coupon inventory; external mutation, approval required
  create_coupons:
    endpoint: POST /lists/{{ record.uuid }}/coupons
    required fields: uuid, coupon_group_id, coupons
    risk: adds redeemable coupon codes to an existing campaign coupon group; external mutation, approval required
  unqualify_referral:
    endpoint: POST /lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id }}/unqualify
    required fields: uuid, subscriber_id
    risk: marks a referral/subscriber as unqualified, changing campaign qualification and reward state; external mutation, approval required
  qualify_referral:
    endpoint: POST /lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id }}/qualify
    required fields: uuid, subscriber_id
    risk: marks a referral/subscriber as qualified, changing campaign qualification and reward state; external mutation, approval required

SECURITY
  read risk: external ReferralHero API read of referral program list, subscriber, referral, reward, and coupon data
  write risk: creates and mutates live ReferralHero campaign, subscriber, transaction, reward, coupon, and qualification state; approval required before execution
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect referralhero

  # Inspect as structured JSON
  pm connectors inspect referralhero --json

AGENT WORKFLOW
  - Run pm connectors inspect referralhero before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
