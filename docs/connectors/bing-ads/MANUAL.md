# pm connectors inspect bing-ads

```text
NAME
  pm connectors inspect bing-ads - Bing Ads connector manual

SYNOPSIS
  pm connectors inspect bing-ads
  pm connectors inspect bing-ads --json
  pm credentials add <name> --connector bing-ads [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Microsoft Advertising (Bing Ads) accounts, users, campaigns, ad groups, and ads through the v13 Customer Management and Campaign Management REST APIs. Read-only.

ICON
  id: bingads
  asset: icons/bingads.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://learn.microsoft.com/en-us/advertising/guides/release-notes

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id
  account_ids
  ad_group_id
  base_url
  campaign_base_url
  campaign_id
  customer_account_id
  customer_id
  token_url
  client_id (secret) (required)
  client_secret (secret)
  developer_token (secret) (required)
  refresh_token (secret) (required)
  tenant_id (secret)

ETL STREAMS
  accounts:
    primary key: Id
    fields: AccountLifeCycleStatus(string), Id(string), Name(string), Number(string), PauseReason(string)
  users:
    primary key: Id
    fields: CustomerId(string), Id(string), JobTitle(string), LastModifiedTime(string), UserLifeCycleStatus(string), UserName(string)
  campaigns:
    primary key: Id
    fields: BudgetType(string), CampaignType(string), DailyBudget(number), Id(string), Name(string), Status(string), TimeZone(string)
  ad_groups:
    primary key: Id
    fields: AdRotation(string), EndDate(string), Id(string), Name(string), Network(string), StartDate(string), Status(string)
  ads:
    primary key: Id
    fields: DevicePreference(string), EditorialStatus(string), Id(string), Status(string), Type(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Microsoft Advertising REST API read of account/user/campaign/ad-group/ad metadata
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect bing-ads

  # Inspect as structured JSON
  pm connectors inspect bing-ads --json

AGENT WORKFLOW
  - Run pm connectors inspect bing-ads before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
