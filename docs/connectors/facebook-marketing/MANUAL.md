# pm connectors inspect facebook-marketing

```text
NAME
  pm connectors inspect facebook-marketing - Facebook Marketing connector manual

SYNOPSIS
  pm connectors inspect facebook-marketing
  pm connectors inspect facebook-marketing --json
  pm credentials add <name> --connector facebook-marketing [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Facebook Marketing ad accounts, campaigns, ads, ad sets, ad creatives, custom audiences, and performance insights, and creates/updates campaigns and ad sets, through the Graph API.

ICON
  asset: icons/facebook.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.facebook.com/docs/marketing-api/marketing-api-changelog

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  ad_account_id
  base_url
  access_token (secret)

ETL STREAMS
  ad_accounts:
    primary key: id
    fields: account_id(), account_status(), currency(), id(), name(), timezone_name()
  campaigns:
    primary key: id
    fields: created_time(), effective_status(), id(), name(), objective(), status(), updated_time()
  ads:
    primary key: id
    fields: created_time(), effective_status(), id(), name(), status(), updated_time()
  ad_sets:
    primary key: id
    fields: bid_amount(), billing_event(), campaign_id(), created_time(), daily_budget(), effective_status(), end_time(), id(), lifetime_budget(), name(), optimization_goal(), start_time(), status(), updated_time()
  ad_creatives:
    primary key: id
    fields: id(), name(), object_story_id(), object_type(), status(), thumbnail_url()
  custom_audiences:
    primary key: id
    fields: approximate_count_lower_bound(), approximate_count_upper_bound(), description(), id(), name(), operation_status(), subtype(), time_created(), time_updated()
  insights:
    primary key: id
    fields: ad_id(), ad_name(), adset_id(), adset_name(), campaign_id(), campaign_name(), clicks(), cpc(), cpm(), ctr(), date_start(), date_stop(), frequency(), id(), impressions(), reach(), spend()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_campaign:
    endpoint: POST /{{ config.ad_account_id }}/campaigns
    risk: external mutation on a live Facebook ad account; creates a campaign that can incur ad spend once ads are attached; approval required
  update_campaign:
    endpoint: POST /{{ record.id }}
    required fields: id
    risk: external mutation on a live Facebook ad account (e.g. pausing/resuming spend); approval required
  create_ad_set:
    endpoint: POST /{{ config.ad_account_id }}/adsets
    risk: external mutation on a live Facebook ad account; creates an ad set that can incur ad spend once ads are attached; approval required

SECURITY
  read risk: external Facebook Graph API read of ad account, campaign, ad, ad set, ad creative, custom audience, and insights (performance metrics) data
  write risk: external mutation of a live Facebook ad account; creating/updating campaigns and ad sets can incur real ad spend once ads are attached and the campaign/ad set is active
  approval: writes require approval; reads are unrestricted
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect facebook-marketing

  # Inspect as structured JSON
  pm connectors inspect facebook-marketing --json

AGENT WORKFLOW
  - Run pm connectors inspect facebook-marketing before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
