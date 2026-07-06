---
name: pm-referralhero
description: ReferralHero connector knowledge and safe action guide.
---

# pm-referralhero

## Purpose

Reads ReferralHero lists, subscribers, referrals, rewards, coupon groups, and campaign-scoped subscriber resources, and performs approved ReferralHero API v2 mutations.

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
- coupon_group_id
- list_uuid
- subscriber_email
- subscriber_id
- subscriber_mwr
- subscriber_name
- api_key (secret)

## ETL Streams

- lists:
  - primary key: id
  - fields: created_at(), id(), name(), status()
- subscribers:
  - primary key: id
  - cursor: updated_at
  - fields: email(), id(), name(), referral_code(), status(), updated_at()
- referrals:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), email(), id(), status(), subscriber_id()
- rewards:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), status(), updated_at()
- list_leaderboard:
  - primary key: id
  - fields: code(), created_at(), crypto_wallet_address(), device(), email(), extra_field(), extra_field_2(), extra_field_3(), extra_field_4(), host(), id(), last_updated_at(), level_2_confirmed_referrals(), level_3_confirmed_referrals(), name(), option_field(), other_identifier_value(), pending_referrals(), people_referred(), phone_number(), points(), position(), promoted(), promoted_at(), referral_link(), referral_status(), referral_status_at(), referred(), referred_by(), response(), risk_level(), source(), stripe_customer_id(), tags(), unconfirmed_referrals(), universal_link(), verified(), verified_at(), visitors()
- list_bonuses:
  - fields: description(), referrals(), title()
- subscribers_search_by_name:
  - primary key: id
  - fields: code(), created_at(), crypto_wallet_address(), device(), email(), extra_field(), extra_field_2(), extra_field_3(), extra_field_4(), host(), id(), last_updated_at(), level_2_confirmed_referrals(), level_3_confirmed_referrals(), name(), option_field(), other_identifier_value(), pending_referrals(), people_referred(), phone_number(), points(), position(), promoted(), promoted_at(), referral_link(), referral_status(), referral_status_at(), referred(), referred_by(), response(), risk_level(), source(), stripe_customer_id(), tags(), unconfirmed_referrals(), universal_link(), verified(), verified_at(), visitors()
- campaign_subscribers:
  - primary key: id
  - fields: code(), created_at(), crypto_wallet_address(), device(), email(), extra_field(), extra_field_2(), extra_field_3(), extra_field_4(), host(), id(), last_updated_at(), level_2_confirmed_referrals(), level_3_confirmed_referrals(), name(), option_field(), other_identifier_value(), pending_referrals(), people_referred(), phone_number(), points(), position(), promoted(), promoted_at(), referral_link(), referral_status(), referral_status_at(), referred(), referred_by(), response(), risk_level(), source(), stripe_customer_id(), tags(), unconfirmed_referrals(), universal_link(), verified(), verified_at(), visitors()
- subscriber_detail:
  - primary key: id
  - fields: code(), created_at(), crypto_wallet_address(), device(), email(), extra_field(), extra_field_2(), extra_field_3(), extra_field_4(), host(), id(), last_updated_at(), level_2_confirmed_referrals(), level_3_confirmed_referrals(), name(), option_field(), other_identifier_value(), pending_referrals(), people_referred(), phone_number(), points(), position(), promoted(), promoted_at(), referral_link(), referral_status(), referral_status_at(), referred(), referred_by(), response(), risk_level(), source(), stripe_customer_id(), tags(), unconfirmed_referrals(), universal_link(), verified(), verified_at(), visitors()
- subscriber_by_email:
  - primary key: id
  - fields: code(), created_at(), crypto_wallet_address(), device(), email(), extra_field(), extra_field_2(), extra_field_3(), extra_field_4(), host(), id(), last_updated_at(), level_2_confirmed_referrals(), level_3_confirmed_referrals(), name(), option_field(), other_identifier_value(), pending_referrals(), people_referred(), phone_number(), points(), position(), promoted(), promoted_at(), referral_link(), referral_status(), referral_status_at(), referred(), referred_by(), response(), risk_level(), source(), stripe_customer_id(), tags(), unconfirmed_referrals(), universal_link(), verified(), verified_at(), visitors()
- subscriber_by_mwr:
  - primary key: id
  - fields: code(), created_at(), crypto_wallet_address(), device(), email(), extra_field(), extra_field_2(), extra_field_3(), extra_field_4(), host(), id(), last_updated_at(), level_2_confirmed_referrals(), level_3_confirmed_referrals(), name(), option_field(), other_identifier_value(), pending_referrals(), people_referred(), phone_number(), points(), position(), promoted(), promoted_at(), referral_link(), referral_status(), referral_status_at(), referred(), referred_by(), response(), risk_level(), source(), stripe_customer_id(), tags(), unconfirmed_referrals(), universal_link(), verified(), verified_at(), visitors()
- subscriber_referrals:
  - primary key: id
  - fields: code(), created_at(), crypto_wallet_address(), device(), email(), extra_field(), extra_field_2(), extra_field_3(), extra_field_4(), host(), id(), last_updated_at(), level_2_confirmed_referrals(), level_3_confirmed_referrals(), name(), option_field(), other_identifier_value(), pending_referrals(), people_referred(), phone_number(), points(), position(), promoted(), promoted_at(), referral_link(), referral_status(), referral_status_at(), referred(), referred_by(), response(), risk_level(), source(), stripe_customer_id(), tags(), unconfirmed_referrals(), universal_link(), verified(), verified_at(), visitors()
- subscriber_level_2_all_referrals:
  - primary key: id
  - fields: code(), created_at(), crypto_wallet_address(), device(), email(), extra_field(), extra_field_2(), extra_field_3(), extra_field_4(), host(), id(), last_updated_at(), level_2_confirmed_referrals(), level_3_confirmed_referrals(), name(), option_field(), other_identifier_value(), pending_referrals(), people_referred(), phone_number(), points(), position(), promoted(), promoted_at(), referral_link(), referral_status(), referral_status_at(), referred(), referred_by(), response(), risk_level(), source(), stripe_customer_id(), tags(), unconfirmed_referrals(), universal_link(), verified(), verified_at(), visitors()
- subscriber_level_3_all_referrals:
  - primary key: id
  - fields: code(), created_at(), crypto_wallet_address(), device(), email(), extra_field(), extra_field_2(), extra_field_3(), extra_field_4(), host(), id(), last_updated_at(), level_2_confirmed_referrals(), level_3_confirmed_referrals(), name(), option_field(), other_identifier_value(), pending_referrals(), people_referred(), phone_number(), points(), position(), promoted(), promoted_at(), referral_link(), referral_status(), referral_status_at(), referred(), referred_by(), response(), risk_level(), source(), stripe_customer_id(), tags(), unconfirmed_referrals(), universal_link(), verified(), verified_at(), visitors()
- subscriber_level_1_referrals:
  - primary key: id
  - fields: code(), created_at(), crypto_wallet_address(), device(), email(), extra_field(), extra_field_2(), extra_field_3(), extra_field_4(), host(), id(), last_updated_at(), level_2_confirmed_referrals(), level_3_confirmed_referrals(), name(), option_field(), other_identifier_value(), pending_referrals(), people_referred(), phone_number(), points(), position(), promoted(), promoted_at(), referral_link(), referral_status(), referral_status_at(), referred(), referred_by(), response(), risk_level(), source(), stripe_customer_id(), tags(), unconfirmed_referrals(), universal_link(), verified(), verified_at(), visitors()
- subscriber_level_2_referrals:
  - primary key: id
  - fields: code(), created_at(), crypto_wallet_address(), device(), email(), extra_field(), extra_field_2(), extra_field_3(), extra_field_4(), host(), id(), last_updated_at(), level_2_confirmed_referrals(), level_3_confirmed_referrals(), name(), option_field(), other_identifier_value(), pending_referrals(), people_referred(), phone_number(), points(), position(), promoted(), promoted_at(), referral_link(), referral_status(), referral_status_at(), referred(), referred_by(), response(), risk_level(), source(), stripe_customer_id(), tags(), unconfirmed_referrals(), universal_link(), verified(), verified_at(), visitors()
- subscriber_level_3_referrals:
  - primary key: id
  - fields: code(), created_at(), crypto_wallet_address(), device(), email(), extra_field(), extra_field_2(), extra_field_3(), extra_field_4(), host(), id(), last_updated_at(), level_2_confirmed_referrals(), level_3_confirmed_referrals(), name(), option_field(), other_identifier_value(), pending_referrals(), people_referred(), phone_number(), points(), position(), promoted(), promoted_at(), referral_link(), referral_status(), referral_status_at(), referred(), referred_by(), response(), risk_level(), source(), stripe_customer_id(), tags(), unconfirmed_referrals(), universal_link(), verified(), verified_at(), visitors()
- campaign_rewards:
  - primary key: id
  - fields: coupon_code(), coupon_group(), created_at(), id(), image_url(), name(), recurring_count(), referral(), referrals(), referrals_type(), sent_date(), signup_type(), status(), subscriber_email(), subscriber_id(), total(), unlocked_date(), value()
- subscriber_rewards:
  - primary key: id
  - fields: coupon_code(), coupon_group(), created_at(), id(), image_url(), name(), recurring_count(), referral(), referrals(), referrals_type(), sent_date(), signup_type(), status(), subscriber_email(), subscriber_id(), total(), unlocked_date(), value()
- coupon_groups:
  - primary key: id
  - fields: active(), coupons(), created_at(), id(), name()
- coupon_group_coupons:
  - primary key: code
  - fields: available(), code(), created_at(), email_id(), sent_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_list:
  - endpoint: POST /lists
  - risk: creates a live ReferralHero campaign/list in the account; external mutation, approval required
- add_subscriber:
  - endpoint: POST /lists/{{ record.uuid }}/subscribers
  - required fields: uuid
  - risk: creates or registers a live subscriber in a ReferralHero campaign and may trigger campaign email/referral workflows; approval required
- track_referral_conversion_event:
  - endpoint: POST /lists/{{ record.uuid }}/subscribers/track_referral_conversion_event
  - required fields: uuid
  - risk: confirms/unconfirms referral conversion state and may create a referral when a referrer is provided; external mutation, approval required
- confirm_subscriber_by_id:
  - endpoint: POST /lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id }}/confirm
  - required fields: uuid, subscriber_id
  - risk: confirms a verified referral/subscriber conversion in the campaign; external mutation, approval required
- confirm_subscriber_by_identifier:
  - endpoint: POST /lists/{{ record.uuid }}/subscribers/confirm
  - required fields: uuid
  - risk: confirms a verified referral/subscriber conversion by unique identifier; external mutation, approval required
- update_subscriber:
  - endpoint: POST /lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id }}
  - required fields: uuid, subscriber_id
  - risk: updates profile, identifier, points, address, or tag fields for a verified subscriber; external mutation, approval required
- add_points:
  - endpoint: POST /lists/{{ record.uuid }}/subscribers/add_points
  - required fields: uuid
  - risk: adds points to a subscriber, changing contest/reward standings; external mutation, approval required
- add_transaction:
  - endpoint: POST /lists/{{ record.uuid }}/subscribers/add_transactions
  - required fields: uuid
  - risk: records a transaction against a subscriber and may affect conversion/reward calculations; external mutation, approval required
- add_bulk_transactions:
  - endpoint: POST /lists/{{ record.uuid }}/subscribers/add_bulk_transactions
  - required fields: uuid
  - risk: records up to 500 transactions in one call and emails an admin CSV result; high-blast-radius external mutation, approval required
- promote_subscriber:
  - endpoint: POST /lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id }}/promote
  - required fields: uuid, subscriber_id
  - risk: promotes a subscriber into the campaign winners/promoted state; external mutation, approval required
- unlock_promoted_reward:
  - endpoint: POST /lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id }}/unlock_promoted_reward
  - required fields: uuid, subscriber_id
  - risk: unlocks a promoted reward for a subscriber, changing reward fulfillment state; external mutation, approval required
- delete_subscriber:
  - endpoint: DELETE /lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id }}
  - required fields: uuid, subscriber_id
  - risk: permanently deletes a subscriber from a live campaign; destructive external mutation, approval required
- update_reward_status:
  - endpoint: POST /lists/{{ record.uuid }}/subscribers/update_reward_status
  - required fields: uuid
  - risk: changes fulfillment status for an unlocked reward; external mutation, approval required
- create_coupon_group:
  - endpoint: POST /lists/{{ record.uuid }}/coupon_groups
  - required fields: uuid
  - risk: creates a campaign coupon group and coupon inventory; external mutation, approval required
- create_coupons:
  - endpoint: POST /lists/{{ record.uuid }}/coupons
  - required fields: uuid
  - risk: adds redeemable coupon codes to an existing campaign coupon group; external mutation, approval required
- unqualify_referral:
  - endpoint: POST /lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id }}/unqualify
  - required fields: uuid, subscriber_id
  - risk: marks a referral/subscriber as unqualified, changing campaign qualification and reward state; external mutation, approval required
- qualify_referral:
  - endpoint: POST /lists/{{ record.uuid }}/subscribers/{{ record.subscriber_id }}/qualify
  - required fields: uuid, subscriber_id
  - risk: marks a referral/subscriber as qualified, changing campaign qualification and reward state; external mutation, approval required

## Security

- read risk: external ReferralHero API read of referral program list, subscriber, referral, reward, and coupon data
- write risk: creates and mutates live ReferralHero campaign, subscriber, transaction, reward, coupon, and qualification state; approval required before execution
- approval: reverse ETL plan approval required before writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect referralhero
```

### Inspect as structured JSON

```bash
pm connectors inspect referralhero --json
```

## Agent Rules

- Run pm connectors inspect referralhero before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
