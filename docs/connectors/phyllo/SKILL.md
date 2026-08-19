---
name: pm-phyllo
description: Phyllo connector knowledge and safe action guide.
---

# pm-phyllo

## Purpose

Reads Phyllo users, accounts, profiles, social content/comments, audience, and income data, and writes user/webhook/account-config mutations using Basic-auth REST endpoints.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- phyllo_account_id
- phyllo_user_id
- phyllo_work_platform_id
- client_id (secret) (required)
- client_secret (secret) (required)

## ETL Streams

- users:
  - primary key: id
  - fields: created_at(string), id(string), platform(string), status(string), updated_at(string)
- accounts:
  - primary key: id
  - fields: created_at(string), id(string), platform(string), status(string), updated_at(string)
- profiles:
  - primary key: id
  - fields: created_at(string), id(string), platform(string), status(string), updated_at(string)
- social_contents:
  - primary key: id
  - fields: created_at(string), id(string), platform(string), status(string), updated_at(string)
- work_platforms:
  - primary key: id
  - fields: category(string), created_at(string), id(string), logo_url(string), name(string), status(string), updated_at(string)
- audience:
  - primary key: account_id
  - fields: account_id(string), age_group(array), cities(array), countries(array), follower_count(integer), gender(array), languages(array), platform_username(string)
- social_content_groups:
  - primary key: id
  - fields: account_id(string), created_at(string), id(string), platform(string), status(string), title(string), type(string), updated_at(string)
- social_comments:
  - primary key: id
  - fields: account_id(string), commenter_username(string), content_id(string), created_at(string), id(string), like_count(integer), platform(string), reply_count(integer), text(string), updated_at(string)
- social_income_transactions:
  - primary key: id
  - fields: account_id(string), amount(number), created_at(string), currency_code(string), id(string), platform(string), transaction_date(string), type(string), updated_at(string)
- social_income_payouts:
  - primary key: id
  - fields: account_id(string), amount(number), created_at(string), currency_code(string), id(string), payout_date(string), platform(string), type(string), updated_at(string)
- commerce_income_transactions:
  - primary key: id
  - fields: account_id(string), amount(number), created_at(string), currency_code(string), id(string), platform(string), transaction_date(string), type(string), updated_at(string)
- commerce_income_payouts:
  - primary key: id
  - fields: account_id(string), amount(number), created_at(string), currency_code(string), id(string), payout_date(string), platform(string), updated_at(string)
- commerce_income_balances:
  - primary key: id
  - fields: account_id(string), amount(number), balance_date(string), created_at(string), currency_code(string), id(string), platform(string), updated_at(string)
- webhooks:
  - primary key: id
  - fields: created_at(string), events(array), id(string), status(string), updated_at(string), url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_user:
  - endpoint: POST /v1/users
  - required fields: name, external_id
  - risk: creates a new Phyllo end-user record that every subsequent Connect/account/profile flow is anchored to; low-risk external mutation, no destructive side effect, no approval required
- update_account:
  - endpoint: PATCH /v1/accounts/{{ record.id }}
  - required fields: id, data
  - risk: changes an account's identity/engagement/income monitoring configuration (e.g. STANDARD vs EXTENSIVE data collection level), affecting what data Phyllo collects going forward; external mutation, approval required
- disconnect_account:
  - endpoint: POST /v1/accounts/{{ record.id }}/disconnect
  - required fields: id
  - risk: revokes Phyllo's connection to the creator's linked social/creator platform account, permanently stopping all future data collection for it; destructive external mutation, approval required
- create_webhook:
  - endpoint: POST /v1/webhooks
  - required fields: url, events
  - risk: registers a new webhook endpoint that will receive Phyllo event notifications; low-risk external mutation, no approval required
- update_webhook:
  - endpoint: PUT /v1/webhooks/{{ record.id }}
  - required fields: id, url, events
  - risk: changes an existing webhook's target URL and/or subscribed event set, redirecting future event delivery; external mutation, approval required
- delete_webhook:
  - endpoint: DELETE /v1/webhooks/{{ record.id }}
  - required fields: id
  - risk: permanently removes a webhook subscription, stopping all future event delivery to it; destructive external mutation, approval required

## Security

- read risk: external Phyllo API read of user, account, profile, social content/comment, audience, and income data
- write risk: creates Phyllo users and webhooks, updates account monitoring configuration and webhook subscriptions, and disconnects linked creator accounts
- approval: required for update_account/update_webhook/disconnect_account/delete_webhook; create_user/create_webhook require no approval (low-risk, non-destructive)
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect phyllo
```

### Inspect as structured JSON

```bash
pm connectors inspect phyllo --json
```

## Agent Rules

- Run pm connectors inspect phyllo before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
