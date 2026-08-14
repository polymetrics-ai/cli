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
  id: amazonads
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
  client_id (secret) (required)
  client_secret (secret) (required)
  refresh_token (secret) (required)

ETL STREAMS
  profiles:
    primary key: profile_id
    fields: account_id(string), account_name(string), account_type(string), country_code(string), currency_code(string), daily_budget(number), marketplace_string_id(string), profile_id(integer), timezone(string)
  campaigns:
    primary key: campaign_id
    fields: campaign_id(integer), campaign_type(string), daily_budget(number), end_date(string), name(string), portfolio_id(integer), premium_bid_adjustment(boolean), start_date(string), state(string), targeting_type(string)
  ad_groups:
    primary key: ad_group_id
    fields: ad_group_id(integer), campaign_id(integer), default_bid(number), name(string), state(string)
  portfolios:
    primary key: portfolio_id
    fields: in_budget(boolean), name(string), portfolio_id(integer), state(string)
  keywords:
    primary key: keyword_id
    fields: ad_group_id(integer), bid(number), campaign_id(integer), keyword_id(integer), keyword_text(string), match_type(string), state(string)
  product_ads:
    primary key: ad_id
    fields: ad_group_id(integer), ad_id(integer), asin(string), campaign_id(integer), serving_status(string), sku(string), state(string)
  negative_keywords:
    primary key: keyword_id
    fields: ad_group_id(integer), campaign_id(integer), keyword_id(integer), keyword_text(string), match_type(string), state(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

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
