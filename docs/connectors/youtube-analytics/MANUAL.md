# pm connectors inspect youtube-analytics

```text
NAME
  pm connectors inspect youtube-analytics - YouTube Analytics connector manual

SYNOPSIS
  pm connectors inspect youtube-analytics
  pm connectors inspect youtube-analytics --json
  pm credentials add <name> --connector youtube-analytics [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads YouTube Reporting API jobs, report types, report metadata, YouTube Analytics groups and group items, and safely plans documented job/group mutations via the Google OAuth 2.0 refresh-token grant.

ICON
  id: youtube-analytics
  asset: icons/youtube-analytics.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.google.com/youtube/analytics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  content_owner_id
  created_after
  include_system_managed
  job_id
  max_pages
  mode
  page_size
  report_id
  scopes
  start_time_at_or_after
  start_time_before
  token_url
  client_id (secret)
  client_secret (secret)
  refresh_token (secret)

ETL STREAMS
  jobs:
    primary key: id
    fields: create_time(string), expire_time(string), id(string), name(string), report_type_id(string), system_managed(boolean)
  job:
    primary key: id
    fields: create_time(string), expire_time(string), id(string), name(string), report_type_id(string), system_managed(boolean)
  report_types:
    primary key: id
    fields: deprecate_time(string), id(string), name(string), system_managed(boolean)
  reports:
    primary key: id
    fields: create_time(string), download_url(string), end_time(string), id(string), job_expire_time(string), job_id(string), start_time(string)
  report:
    primary key: id
    fields: create_time(string), download_url(string), end_time(string), id(string), job_expire_time(string), job_id(string), start_time(string)
  groups:
    primary key: id
    fields: etag(string), id(string), item_count(string), item_type(string), kind(string), published_at(string), title(string)
  group_items:
    primary key: id
    fields: etag(string), group_id(string), id(string), kind(string), resource_id(string), resource_kind(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_job:
    endpoint: POST /jobs
    required fields: reportTypeId, name
    risk: creates a YouTube Reporting job that schedules future report generation; requires reverse ETL plan, preview, and explicit approval
  delete_job:
    endpoint: DELETE /jobs/{{ record.job_id }}
    required fields: job_id
    risk: deletes a scheduled YouTube Reporting job and stops future report generation; destructive, redacts job_id in previews/errors, and requires reverse ETL plan, preview, explicit approval, and typed confirmation
  create_group:
    endpoint: POST https://youtubeanalytics.googleapis.com/v2/groups
    required fields: snippet, contentDetails
    risk: creates a YouTube Analytics group; requires reverse ETL plan, preview, and explicit approval
  update_group:
    endpoint: PUT https://youtubeanalytics.googleapis.com/v2/groups
    required fields: id, snippet
    risk: updates a YouTube Analytics group's metadata (for example title); redacts group id in write errors and requires reverse ETL plan, preview, and explicit approval
  delete_group:
    endpoint: DELETE https://youtubeanalytics.googleapis.com/v2/groups?id={{ record.id | urlencode }}
    required fields: id
    risk: deletes a YouTube Analytics group; destructive, redacts group id in previews/errors, and requires reverse ETL plan, preview, explicit approval, and typed confirmation
  create_group_item:
    endpoint: POST https://youtubeanalytics.googleapis.com/v2/groupItems
    required fields: groupId, resource
    risk: adds a channel, playlist, video, or asset to a YouTube Analytics group; redacts group/resource identifiers in write errors and requires reverse ETL plan, preview, and explicit approval
  delete_group_item:
    endpoint: DELETE https://youtubeanalytics.googleapis.com/v2/groupItems?id={{ record.id | urlencode }}
    required fields: id
    risk: removes an item from a YouTube Analytics group; destructive, redacts group-item id in previews/errors, and requires reverse ETL plan, preview, explicit approval, and typed confirmation

SECURITY
  read risk: external YouTube Reporting and YouTube Analytics API reads of reporting jobs, report types, report metadata, groups, and group items
  approval: reverse ETL mutations are available only through plan, preview, explicit approval, and destructive confirmation where required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run YouTube Analytics's declared streams and reverse-ETL actions.
  Usage: pm youtube-analytics <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api get v1 media resource-name - Documented GET /v1/media/{resource_name} (not implemented) [intent=direct_read availability=not_implemented operation=youtube-analytics.get.v1-media-resource-name]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 reports - Documented GET /v2/reports (not implemented) [intent=direct_read availability=not_implemented operation=youtube-analytics.get.v2-reports]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    group-items create - Plan adding a channel, playlist, video, or asset to a YouTube Analytics group. [intent=reverse_etl availability=not_implemented write=create_group_item]; approval: requires plan, preview, approval, and execute; risk: adds a channel, playlist, video, or asset to a YouTube Analytics group; redacts group/resource identifiers in write errors and requires reverse ETL plan, preview, and explicit approval; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    group-items delete - Plan removal of an item from a YouTube Analytics group with typed destructive confirmation. [intent=reverse_etl availability=implemented write=delete_group_item]; approval: requires plan, preview, approval, and execute; risk: removes an item from a YouTube Analytics group; destructive, redacts group-item id in previews/errors, and requires reverse ETL plan, preview, explicit approval, and typed confirmation; flags: --id (required)
    group-items list - Read items in a YouTube Analytics group through the declared ETL stream. [intent=etl availability=implemented stream=group_items]; flags: --group-id (required)
    groups create - Plan creation of a YouTube Analytics group. [intent=reverse_etl availability=not_implemented write=create_group]; approval: requires plan, preview, approval, and execute; risk: creates a YouTube Analytics group; requires reverse ETL plan, preview, and explicit approval; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    groups delete - Plan deletion of a YouTube Analytics group with typed destructive confirmation. [intent=reverse_etl availability=implemented write=delete_group]; approval: requires plan, preview, approval, and execute; risk: deletes a YouTube Analytics group; destructive, redacts group id in previews/errors, and requires reverse ETL plan, preview, explicit approval, and typed confirmation; flags: --id (required)
    groups list - Read YouTube Analytics groups through the declared ETL stream. [intent=etl availability=implemented stream=groups]; flags: --mine (required)
    groups update - Plan update of a YouTube Analytics group's title. [intent=reverse_etl availability=not_implemented write=update_group]; approval: requires plan, preview, approval, and execute; risk: updates a YouTube Analytics group's metadata (for example title); redacts group id in write errors and requires reverse ETL plan, preview, and explicit approval; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    jobs create - Plan creation of a YouTube Reporting job. [intent=reverse_etl availability=implemented write=create_job]; approval: requires plan, preview, approval, and execute; risk: creates a YouTube Reporting job that schedules future report generation; requires reverse ETL plan, preview, and explicit approval; flags: --name (required), --reportTypeId (required)
    jobs delete - Plan deletion of a YouTube Reporting job with typed destructive confirmation. [intent=reverse_etl availability=implemented write=delete_job]; approval: requires plan, preview, approval, and execute; risk: deletes a scheduled YouTube Reporting job and stops future report generation; destructive, redacts job_id in previews/errors, and requires reverse ETL plan, preview, explicit approval, and typed confirmation; flags: --job_id (required)
    jobs get - Read one YouTube Reporting job by configured job_id. [intent=etl availability=implemented stream=job]; notes: Pass --config job_id=<job id>; ETL command flags intentionally map only to query parameters, not connector config path variables.
    jobs list - Read YouTube Reporting jobs through the declared ETL stream. [intent=etl availability=implemented stream=jobs]; flags: --include-system-managed
    report-types list - Read YouTube Reporting report types through the declared ETL stream. [intent=etl availability=implemented stream=report_types]; flags: --include-system-managed
    reports download - Download generated YouTube Reporting report bytes to an explicit local destination. [intent=binary_download availability=implemented operation=download_report]; notes: The required --resource-name maps to the provider resourceName. Runtime download flags require --dest-root; --file-name is optional. The executor refuses path traversal, overwrites, archive extraction, and payloads over 100 MiB, and emits only file metadata with a SHA-256 receipt.; flags: --resource-name (required), --dest-root (required), --file-name, --max-bytes
    reports get - Read one generated report metadata resource for the configured job_id/report_id. [intent=etl availability=implemented stream=report]; notes: Pass --config job_id=<job id> --config report_id=<report id>; use reports download with the documented resourceName to fetch report bytes.
    reports list - Read generated report metadata for the configured YouTube Reporting job. [intent=etl availability=implemented stream=reports]; notes: Pass --config job_id=<job id>; downloaded report bytes are intentionally not emitted by this JSON metadata stream.; flags: --created-after, --start-time-at-or-after, --start-time-before

EXAMPLES
  # Inspect as a manual
  pm connectors inspect youtube-analytics

  # Inspect as structured JSON
  pm connectors inspect youtube-analytics --json

AGENT WORKFLOW
  - Run pm connectors inspect youtube-analytics before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
