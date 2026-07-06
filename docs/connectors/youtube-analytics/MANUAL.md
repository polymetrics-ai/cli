# pm connectors inspect youtube-analytics

```text
NAME
  pm connectors inspect youtube-analytics - YouTube Analytics connector manual

SYNOPSIS
  pm connectors inspect youtube-analytics
  pm connectors inspect youtube-analytics --json
  pm credentials add <name> --connector youtube-analytics [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads YouTube Reporting API jobs, report types, and generated reports via the Google OAuth 2.0 refresh-token grant.

ICON
  asset: icons/youtube-analytics.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.google.com/youtube/analytics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  content_owner_id
  job_id
  max_pages
  mode
  page_size
  scopes
  token_url
  client_id (secret)
  client_secret (secret)
  refresh_token (secret)

ETL STREAMS
  jobs:
    primary key: id
    fields: create_time(), expire_time(), id(), name(), report_type_id(), system_managed()
  report_types:
    primary key: id
    fields: deprecate_time(), id(), name(), system_managed()
  reports:
    primary key: id
    fields: create_time(), download_url(), end_time(), id(), job_expire_time(), job_id(), start_time()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external YouTube Reporting API read of reporting-job/report-type/report metadata
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect youtube-analytics

  # Inspect as structured JSON
  pm connectors inspect youtube-analytics --json

AGENT WORKFLOW
  - Run pm connectors inspect youtube-analytics before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
