# pm connectors inspect amazon-ads

```text
NAME
  pm connectors inspect amazon-ads - Amazon Ads connector manual

SYNOPSIS
  pm connectors inspect amazon-ads
  pm connectors inspect amazon-ads --json
  pm credentials add <name> --connector amazon-ads [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Amazon Advertising profiles, Sponsored Products campaigns, ad groups, product ads, keywords, negative keywords, and portfolios via the Amazon Ads API using a Login with Amazon (LWA) refresh-token grant. Read-only.

ICON
  asset: icons/amazonads.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://advertising.amazon.com/API/docs/en-us/release-notes/deprecations

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  page_size
  profile_id
  token_url
  client_id (secret)
  client_secret (secret)
  refresh_token (secret)

ETL STREAMS
  profiles:
    primary key: profile_id
    fields: account_id(), account_name(), account_type(), country_code(), currency_code(), daily_budget(), marketplace_string_id(), profile_id(), timezone()
  campaigns:
    primary key: campaign_id
    fields: campaign_id(), campaign_type(), daily_budget(), end_date(), name(), portfolio_id(), premium_bid_adjustment(), start_date(), state(), targeting_type()
  ad_groups:
    primary key: ad_group_id
    fields: ad_group_id(), campaign_id(), default_bid(), name(), state()
  portfolios:
    primary key: portfolio_id
    fields: in_budget(), name(), portfolio_id(), state()
  keywords:
    primary key: keyword_id
    fields: ad_group_id(), bid(), campaign_id(), keyword_id(), keyword_text(), match_type(), state()
  product_ads:
    primary key: ad_id
    fields: ad_group_id(), ad_id(), asin(), campaign_id(), serving_status(), sku(), state()
  negative_keywords:
    primary key: keyword_id
    fields: ad_group_id(), campaign_id(), keyword_id(), keyword_text(), match_type(), state()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Amazon Ads API read of profile, campaign, ad group, product ad, keyword, negative keyword, and portfolio data
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect amazon-ads

  # Inspect as structured JSON
  pm connectors inspect amazon-ads --json

AGENT WORKFLOW
  - Run pm connectors inspect amazon-ads before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
