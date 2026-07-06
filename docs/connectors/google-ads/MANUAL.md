# pm connectors inspect google-ads

```text
NAME
  pm connectors inspect google-ads - Google Ads connector manual

SYNOPSIS
  pm connectors inspect google-ads
  pm connectors inspect google-ads --json
  pm credentials add <name> --connector google-ads [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads accessible customers and allow-listed Google Ads GAQL search resources (campaigns, ad groups) through the Google Ads REST API. Read-only; arbitrary GAQL is not accepted.

ICON
  asset: icons/google-adwords.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.google.com/google-ads/api/docs/release-notes

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  customer_id
  login_customer_id
  max_pages
  mode
  page_size
  access_token (secret)
  developer_token (secret)

ETL STREAMS
  accessible_customers:
    primary key: customer_id
    fields: customer_id(), resource_name()
  campaigns:
    primary key: id
    fields: id(), name(), resource_name(), status()
  ad_groups:
    primary key: id
    fields: id(), name(), resource_name(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Google Ads API read of customer/campaign/ad-group metadata
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect google-ads

  # Inspect as structured JSON
  pm connectors inspect google-ads --json

AGENT WORKFLOW
  - Run pm connectors inspect google-ads before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
