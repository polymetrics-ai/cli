# pm connectors inspect linkedin-ads

```text
NAME
  pm connectors inspect linkedin-ads - LinkedIn Ads connector manual

SYNOPSIS
  pm connectors inspect linkedin-ads
  pm connectors inspect linkedin-ads --json
  pm credentials add <name> --connector linkedin-ads [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads LinkedIn Ads accounts, campaign groups, campaigns, and creatives through the LinkedIn Marketing REST API.

ICON
  asset: icons/linkedin.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://learn.microsoft.com/en-us/linkedin/marketing/integrations/recent-changes?view=li-lms-2024-10

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  linkedin_version
  max_pages
  mode
  page_size
  access_token (secret)

ETL STREAMS
  accounts:
    primary key: id
    cursor: last_modified
    fields: created_at(), currency(), id(), last_modified(), name(), reference(), status(), test(), type(), version()
  campaign_groups:
    primary key: id
    cursor: last_modified
    fields: account(), created_at(), id(), last_modified(), name(), run_schedule(), status(), total_budget()
  campaigns:
    primary key: id
    cursor: last_modified
    fields: account(), campaign_group(), cost_type(), created_at(), daily_budget(), format(), id(), last_modified(), name(), objective_type(), run_schedule(), status(), type(), unit_cost()
  creatives:
    primary key: id
    cursor: last_modified
    fields: account(), campaign(), content(), created_at(), id(), intended_status(), is_serving(), last_modified(), review_status(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external LinkedIn Marketing API read of ad account, campaign, and creative data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect linkedin-ads

  # Inspect as structured JSON
  pm connectors inspect linkedin-ads --json

AGENT WORKFLOW
  - Run pm connectors inspect linkedin-ads before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
