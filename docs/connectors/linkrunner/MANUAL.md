# pm connectors inspect linkrunner

```text
NAME
  pm connectors inspect linkrunner - Linkrunner connector manual

SYNOPSIS
  pm connectors inspect linkrunner
  pm connectors inspect linkrunner --json
  pm credentials add <name> --connector linkrunner [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Linkrunner mobile attribution campaigns and attributed users from the Linkrunner Data API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  channel
  display_id
  end_timestamp
  filter
  max_pages
  mode
  page_size
  start_timestamp
  timezone
  linkrunner-key (secret) (required)

ETL STREAMS
  campaigns:
    primary key: display_id
    cursor: update_at
    fields: active(boolean), attributed_users(number), created_at(string), default_link(boolean), display_id(string), domain(string), google(boolean), link(string), meta(boolean), meta_campaign_id(string), name(string), shareable_link(string), update_at(string), website(string)
  attributed_users:
    primary key: campaign_display_id, attributed_at
    cursor: attributed_at
    fields: ad_set_id(string), attributed_at(string), campaign_display_id(string), campaign_name(string), installed_at(string), link(string), store_click_at(string), user_data(object)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Linkrunner API read of mobile attribution campaign and user data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect linkrunner

  # Inspect as structured JSON
  pm connectors inspect linkrunner --json

AGENT WORKFLOW
  - Run pm connectors inspect linkrunner before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
