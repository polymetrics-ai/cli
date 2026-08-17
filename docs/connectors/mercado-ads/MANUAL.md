# pm connectors inspect mercado-ads

```text
NAME
  pm connectors inspect mercado-ads - Mercado Ads connector manual

SYNOPSIS
  pm connectors inspect mercado-ads
  pm connectors inspect mercado-ads --json
  pm credentials add <name> --connector mercado-ads [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Mercado Ads brand, display, and product advertisers and daily campaign metrics from the Mercado Libre Advertising API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  end_date
  lookback_days
  mode
  start_date
  client_id (secret)
  client_refresh_token (secret)
  client_secret (secret)

ETL STREAMS
  brand_advertisers:
    primary key: advertiser_id
    fields: account_name(), advertiser_id(), advertiser_name(), site_id()
  display_advertisers:
    primary key: advertiser_id
    fields: account_name(), advertiser_id(), advertiser_name(), site_id()
  product_advertisers:
    primary key: advertiser_id
    fields: account_name(), advertiser_id(), advertiser_name(), site_id()
  brand_campaigns_metrics:
    primary key: date, advertiser_id, campaign_id
    cursor: date
    fields: acos(), advertiser_id(), campaign_id(), clicks(), cost(), cpc(), ctr(), date(), direct_amount(), indirect_amount(), prints(), total_amount(), units_quantity()
  display_campaigns_metrics:
    primary key: date, advertiser_id, campaign_id
    cursor: date
    fields: acos(), advertiser_id(), campaign_id(), clicks(), cost(), cpc(), ctr(), date(), direct_amount(), indirect_amount(), prints(), total_amount(), units_quantity()
  product_campaigns_metrics:
    primary key: date, advertiser_id, campaign_id
    cursor: date
    fields: acos(), advertiser_id(), campaign_id(), clicks(), cost(), cpc(), ctr(), date(), direct_amount(), indirect_amount(), prints(), total_amount(), units_quantity()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Mercado Ads API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mercado-ads

  # Inspect as structured JSON
  pm connectors inspect mercado-ads --json

AGENT WORKFLOW
  - Run pm connectors inspect mercado-ads before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
