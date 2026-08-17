# pm connectors inspect criteo-marketing

```text
NAME
  pm connectors inspect criteo-marketing - Criteo Marketing connector manual

SYNOPSIS
  pm connectors inspect criteo-marketing
  pm connectors inspect criteo-marketing --json
  pm credentials add <name> --connector criteo-marketing [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Criteo Marketing Solutions ad sets, advertisers, campaigns, audiences, ad spend statistics, and Marketplace Performance Outcomes (MPO) advertisers/sellers/budgets/seller-campaigns through the Criteo REST API using OAuth2 client-credentials auth.

ICON
  asset: icons/criteo.svg
  source: official
  review_status: official_verified
  review_url: https://developers.criteo.com/marketing-solutions/reference/getting-started

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  currency
  end_date
  start_date
  token_url
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  ad_sets:
    primary key: id
    fields: advertiserId(), campaignId(), datasetId(), destinationEnvironment(), id(), mediaType(), name(), objective(), status(), type()
  advertisers:
    primary key: id
    fields: country(), currency(), id(), name(), timezone(), type()
  campaigns:
    primary key: id
    fields: advertiserId(), goal(), id(), name(), objective(), spendLimit(), type()
  audiences:
    primary key: id
    fields: advertiserId(), description(), id(), name(), nbActiveUsers(), type()
  statistics:
    primary key: AdvertiserId, CampaignId, Day
    cursor: Day
    fields: AdvertiserId(), CampaignId(), Clicks(), Currency(), Day(), Displays(), Spend()
  mpo_advertisers:
    primary key: id
    fields: advertiserName(), currencyName(), id(), timeZoneId()
  mpo_sellers:
    primary key: id
    fields: id(), sellerName()
  mpo_budgets:
    primary key: id
    fields: amount(), budgetType(), campaignIds(), endDate(), id(), isSuspended(), sellerId(), spend(), startDate(), status()
  mpo_seller_campaigns:
    primary key: id
    fields: bid(), campaignId(), id(), sellerId(), suspendedSince(), suspensionReasons()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Criteo Marketing Solutions API read of advertiser, campaign, and ad spend data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect criteo-marketing

  # Inspect as structured JSON
  pm connectors inspect criteo-marketing --json

AGENT WORKFLOW
  - Run pm connectors inspect criteo-marketing before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
