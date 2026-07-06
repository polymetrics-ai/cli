# pm connectors inspect phyllo

```text
NAME
  pm connectors inspect phyllo - Phyllo connector manual

SYNOPSIS
  pm connectors inspect phyllo
  pm connectors inspect phyllo --json
  pm credentials add <name> --connector phyllo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Phyllo users, accounts, profiles, social content/comments, audience, and income data, and writes user/webhook/account-config mutations using Basic-auth REST endpoints.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  phyllo_account_id
  phyllo_user_id
  phyllo_work_platform_id
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  users:
    primary key: id
    fields: created_at(), id(), platform(), status(), updated_at()
  accounts:
    primary key: id
    fields: created_at(), id(), platform(), status(), updated_at()
  profiles:
    primary key: id
    fields: created_at(), id(), platform(), status(), updated_at()
  social_contents:
    primary key: id
    fields: created_at(), id(), platform(), status(), updated_at()
  work_platforms:
    primary key: id
    fields: category(), created_at(), id(), logo_url(), name(), status(), updated_at()
  audience:
    primary key: account_id
    fields: account_id(), age_group(), cities(), countries(), follower_count(), gender(), languages(), platform_username()
  social_content_groups:
    primary key: id
    fields: account_id(), created_at(), id(), platform(), status(), title(), type(), updated_at()
  social_comments:
    primary key: id
    fields: account_id(), commenter_username(), content_id(), created_at(), id(), like_count(), platform(), reply_count(), text(), updated_at()
  social_income_transactions:
    primary key: id
    fields: account_id(), amount(), created_at(), currency_code(), id(), platform(), transaction_date(), type(), updated_at()
  social_income_payouts:
    primary key: id
    fields: account_id(), amount(), created_at(), currency_code(), id(), payout_date(), platform(), type(), updated_at()
  commerce_income_transactions:
    primary key: id
    fields: account_id(), amount(), created_at(), currency_code(), id(), platform(), transaction_date(), type(), updated_at()
  commerce_income_payouts:
    primary key: id
    fields: account_id(), amount(), created_at(), currency_code(), id(), payout_date(), platform(), updated_at()
  commerce_income_balances:
    primary key: id
    fields: account_id(), amount(), balance_date(), created_at(), currency_code(), id(), platform(), updated_at()
  webhooks:
    primary key: id
    fields: created_at(), events(), id(), status(), updated_at(), url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_user:
    endpoint: POST /v1/users
    risk: creates a new Phyllo end-user record that every subsequent Connect/account/profile flow is anchored to; low-risk external mutation, no destructive side effect, no approval required
  update_account:
    endpoint: PATCH /v1/accounts/{{ record.id }}
    required fields: id
    optional fields: data
    risk: changes an account's identity/engagement/income monitoring configuration (e.g. STANDARD vs EXTENSIVE data collection level), affecting what data Phyllo collects going forward; external mutation, approval required
  disconnect_account:
    endpoint: POST /v1/accounts/{{ record.id }}/disconnect
    required fields: id
    risk: revokes Phyllo's connection to the creator's linked social/creator platform account, permanently stopping all future data collection for it; destructive external mutation, approval required
  create_webhook:
    endpoint: POST /v1/webhooks
    risk: registers a new webhook endpoint that will receive Phyllo event notifications; low-risk external mutation, no approval required
  update_webhook:
    endpoint: PUT /v1/webhooks/{{ record.id }}
    required fields: id
    risk: changes an existing webhook's target URL and/or subscribed event set, redirecting future event delivery; external mutation, approval required
  delete_webhook:
    endpoint: DELETE /v1/webhooks/{{ record.id }}
    required fields: id
    risk: permanently removes a webhook subscription, stopping all future event delivery to it; destructive external mutation, approval required

SECURITY
  read risk: external Phyllo API read of user, account, profile, social content/comment, audience, and income data
  write risk: creates Phyllo users and webhooks, updates account monitoring configuration and webhook subscriptions, and disconnects linked creator accounts
  approval: required for update_account/update_webhook/disconnect_account/delete_webhook; create_user/create_webhook require no approval (low-risk, non-destructive)
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect phyllo

  # Inspect as structured JSON
  pm connectors inspect phyllo --json

AGENT WORKFLOW
  - Run pm connectors inspect phyllo before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
