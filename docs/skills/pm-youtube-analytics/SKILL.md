---
name: pm-youtube-analytics
description: YouTube Analytics connector knowledge and safe action guide.
---

# pm-youtube-analytics

## Purpose

Reads YouTube Reporting API jobs, report types, report metadata, YouTube Analytics groups and group items, and safely plans documented job/group mutations via the Google OAuth 2.0 refresh-token grant.

## Icon

- id: youtube-analytics
- asset: icons/youtube-analytics.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.google.com/youtube/analytics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- content_owner_id
- created_after
- include_system_managed
- job_id
- max_pages
- mode
- page_size
- report_id
- scopes
- start_time_at_or_after
- start_time_before
- token_url
- client_id (secret) (required)
- client_secret (secret)
- refresh_token (secret) (required)

## ETL Streams

- jobs:
  - primary key: id
  - fields: create_time(string), expire_time(string), id(string), name(string), report_type_id(string), system_managed(boolean)
- job:
  - primary key: id
  - fields: create_time(string), expire_time(string), id(string), name(string), report_type_id(string), system_managed(boolean)
- report_types:
  - primary key: id
  - fields: deprecate_time(string), id(string), name(string), system_managed(boolean)
- reports:
  - primary key: id
  - fields: create_time(string), download_url(string), end_time(string), id(string), job_expire_time(string), job_id(string), start_time(string)
- report:
  - primary key: id
  - fields: create_time(string), download_url(string), end_time(string), id(string), job_expire_time(string), job_id(string), start_time(string)
- groups:
  - primary key: id
  - fields: etag(string), id(string), item_count(string), item_type(string), kind(string), published_at(string), title(string)
- group_items:
  - primary key: id
  - fields: etag(string), group_id(string), id(string), kind(string), resource_id(string), resource_kind(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_job:
  - endpoint: POST /jobs
  - required fields: reportTypeId, name
  - risk: creates a YouTube Reporting job that schedules future report generation; requires reverse ETL plan, preview, and explicit approval
- delete_job:
  - endpoint: DELETE /jobs/{{ record.job_id }}
  - required fields: job_id
  - risk: deletes a scheduled YouTube Reporting job and stops future report generation; destructive, redacts job_id in previews/errors, and requires reverse ETL plan, preview, explicit approval, and typed confirmation
- create_group:
  - endpoint: POST https://youtubeanalytics.googleapis.com/v2/groups
  - required fields: snippet, contentDetails
  - risk: creates a YouTube Analytics group; requires reverse ETL plan, preview, and explicit approval
- update_group:
  - endpoint: PUT https://youtubeanalytics.googleapis.com/v2/groups
  - required fields: id, snippet
  - risk: updates a YouTube Analytics group's metadata (for example title); redacts group id in write errors and requires reverse ETL plan, preview, and explicit approval
- delete_group:
  - endpoint: DELETE https://youtubeanalytics.googleapis.com/v2/groups?id={{ record.id | urlencode }}
  - required fields: id
  - risk: deletes a YouTube Analytics group; destructive, redacts group id in previews/errors, and requires reverse ETL plan, preview, explicit approval, and typed confirmation
- create_group_item:
  - endpoint: POST https://youtubeanalytics.googleapis.com/v2/groupItems
  - required fields: groupId, resource
  - risk: adds a channel, playlist, video, or asset to a YouTube Analytics group; redacts group/resource identifiers in write errors and requires reverse ETL plan, preview, and explicit approval
- delete_group_item:
  - endpoint: DELETE https://youtubeanalytics.googleapis.com/v2/groupItems?id={{ record.id | urlencode }}
  - required fields: id
  - risk: removes an item from a YouTube Analytics group; destructive, redacts group-item id in previews/errors, and requires reverse ETL plan, preview, explicit approval, and typed confirmation

## Security

- read risk: external YouTube Reporting and YouTube Analytics API reads of reporting jobs, report types, report metadata, groups, and group items
- approval: reverse ETL mutations are available only through plan, preview, explicit approval, and destructive confirmation where required
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Read YouTube Reporting/Analytics metadata and safely plan documented job/group mutations.
- Usage: pm youtube-analytics <command> [flags]
- Source CLI: YouTube Analytics and Reporting APIs (Google Discovery: YouTube Analytics API v2 revision 20260803 and YouTube Reporting API v1 revision 20260803)
- Global flags:
  - --credential (string): Credential name to use for the YouTube request.
  - --connection (string): Credential name alias used only when --credential is omitted; does not resolve pm connections.
  - --config (string_array): Connector config override as key=value; never pass secret values here.
  - --json (boolean): Emit machine-readable JSON output.
  - --limit (integer): Maximum records to emit from stream commands.
  - --preview (boolean): Preview a reverse-ETL write command without making a network mutation.
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  - --confirm (string): Typed confirmation challenge for destructive reverse-ETL writes.
- YouTube Reporting
  - jobs list - Read YouTube Reporting jobs through the declared ETL stream. [intent=etl availability=implemented stream=jobs]; flags: --include-system-managed (max 4096 bytes) (boolean): Include system-managed reporting jobs.: maps_to=query.includeSystemManaged
  - jobs get - Read one YouTube Reporting job by configured job_id. [intent=etl availability=implemented stream=job]; notes: Pass --config job_id=<job id>; ETL command flags intentionally map only to query parameters, not connector config path variables.
  - jobs create - Plan creation of a YouTube Reporting job. [intent=reverse_etl availability=implemented write=create_job]; approval: Plan first, inspect preview output, then run only with the generated approval token.; risk: Creates a scheduled YouTube Reporting job; requires reverse ETL plan, preview, explicit approval, then execute.; flags: --report-type-id (required, non-empty) (string): ReportType id the job should generate.: maps_to=record.reportTypeId, --name (required, non-empty) (string): Reporting job display name required by the provider.: maps_to=record.name
  - jobs delete - Plan deletion of a YouTube Reporting job with typed destructive confirmation. [intent=reverse_etl availability=implemented write=delete_job]; approval: Plan first, inspect preview output, then run only with the generated approval token and typed --confirm destructive challenge.; risk: Destructive job deletion stops future report generation; redacts job_id from previews/errors; requires reverse ETL plan, preview, explicit approval, and --confirm destructive before execute.; flags: --job-id (required, non-empty) (string): Reporting job ID to delete.: maps_to=record.job_id
  - report-types list - Read YouTube Reporting report types through the declared ETL stream. [intent=etl availability=implemented stream=report_types]; flags: --include-system-managed (max 4096 bytes) (boolean): Include system-managed report types.: maps_to=query.includeSystemManaged
  - reports list - Read generated report metadata for the configured YouTube Reporting job. [intent=etl availability=implemented stream=reports]; notes: Pass --config job_id=<job id>; downloaded report bytes are intentionally not emitted by this JSON metadata stream.; flags: --created-after (non-empty, max 4096 bytes, format=date-time) (string): createdAfter filter for reports.list.: maps_to=query.createdAfter, --start-time-at-or-after (non-empty, max 4096 bytes, format=date-time) (string): startTimeAtOrAfter filter for reports.list.: maps_to=query.startTimeAtOrAfter, --start-time-before (non-empty, max 4096 bytes, format=date-time) (string): startTimeBefore filter for reports.list.: maps_to=query.startTimeBefore
  - reports get - Read one generated report metadata resource for the configured job_id/report_id. [intent=etl availability=implemented stream=report]; notes: Pass --config job_id=<job id> --config report_id=<report id>; use reports download with the documented resourceName to fetch report bytes.
  - reports query - Documented YouTube Analytics reports.query provider query operation planned for issue #2985. [intent=direct_read availability=planned operation=reports_query]; notes: Planned solely until typed provider-query foundation issue #2985 supports bounded provider-owned query fields without raw query escape hatches.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - reports download - Download generated YouTube Reporting report bytes to an explicit local destination. [intent=binary_download availability=implemented operation=download_report]; notes: The required --resource-name maps to the provider resourceName. Runtime download flags require --dest-root; --file-name is optional. The executor refuses path traversal, overwrites, archive extraction, and payloads over 100 MiB, and emits only file metadata with a SHA-256 receipt.; flags: --resource-name (required, max 4096 bytes) (string): Provider resourceName identifying the generated report; may contain safe path segments.: maps_to=path.path, --dest-root (required) (string): directory the download is written beneath; traversal outside it is refused., --file-name (string): name for the downloaded file within --dest-root; must be a single path segment., --max-bytes (integer): lower the operation's declared size cap; it can never raise it.
- YouTube Analytics groups
  - groups list - Read YouTube Analytics groups through the declared ETL stream. [intent=etl availability=implemented stream=groups]; flags: --mine (required, max 4096 bytes) (enum): Required closed groups.list selector; only the provider-valid mine=true mode is supported.: values=true: maps_to=query.mine
  - groups create - Plan creation of a YouTube Analytics group. [intent=reverse_etl availability=implemented write=create_group]; approval: Plan first, inspect preview output, then run only with the generated approval token.; risk: Creates a YouTube Analytics group; requires reverse ETL plan, preview, explicit approval, then execute.; flags: --title (required, non-empty) (string): Group title.: maps_to=record.snippet.title, --item-type (required) (enum): Provider resource type the group will contain.: values=youtube#channel|youtube#playlist|youtube#video|youtubePartner#asset: maps_to=record.contentDetails.itemType
  - groups update - Plan update of a YouTube Analytics group's title. [intent=reverse_etl availability=implemented write=update_group]; approval: Plan first, inspect preview output, then run only with the generated approval token.; risk: Mutates YouTube Analytics group metadata; redacts group id from write errors; requires reverse ETL plan, preview, explicit approval, then execute.; flags: --id (required, non-empty) (string): YouTube Analytics group id to update.: maps_to=record.id, --title (required, non-empty) (string): Replacement group title.: maps_to=record.snippet.title
  - groups delete - Plan deletion of a YouTube Analytics group with typed destructive confirmation. [intent=reverse_etl availability=implemented write=delete_group]; approval: Plan first, inspect preview output, then run only with the generated approval token and typed --confirm destructive challenge.; risk: Destructive group deletion; redacts group id from previews/errors; requires reverse ETL plan, preview, explicit approval, and --confirm destructive before execute.; flags: --id (required, non-empty) (string): YouTube Analytics group id to delete.: maps_to=record.id
  - group-items list - Read items in a YouTube Analytics group through the declared ETL stream. [intent=etl availability=implemented stream=group_items]; flags: --group-id (required, non-empty, max 4096 bytes) (string): Group ID for the required groupItems.list groupId query value.: maps_to=query.groupId
  - group-items create - Plan adding a channel, playlist, video, or asset to a YouTube Analytics group. [intent=reverse_etl availability=implemented write=create_group_item]; approval: Plan first, inspect preview output, then run only with the generated approval token.; risk: Adds an item to a YouTube Analytics group; redacts group/resource identifiers from write errors; requires reverse ETL plan, preview, explicit approval, then execute.; flags: --group-id (required, non-empty) (string): Group ID to add the resource to.: maps_to=record.groupId, --resource-kind (enum): Optional provider resource kind being added.: values=youtube#channel|youtube#playlist|youtube#video|youtubePartner#asset: maps_to=record.resource.kind, --resource-id (required, non-empty) (string): Resource ID being added.: maps_to=record.resource.id
  - group-items delete - Plan removal of an item from a YouTube Analytics group with typed destructive confirmation. [intent=reverse_etl availability=implemented write=delete_group_item]; approval: Plan first, inspect preview output, then run only with the generated approval token and typed --confirm destructive challenge.; risk: Destructive removal of a group item; redacts group-item id from previews/errors; requires reverse ETL plan, preview, explicit approval, and --confirm destructive before execute.; flags: --id (required, non-empty) (string): Group item id to remove.: maps_to=record.id
- Help topics:
  - binary-downloads - Generated report metadata includes download_url; use reports download with the documented resourceName and an explicit destination root to fetch report bytes safely.
  - content-owner-writes - Read streams support onBehalfOfContentOwner from content_owner_id; declarative write actions omit that optional query until write-action query templates are available.
  - destructive-confirmation - YouTube Reporting job deletes and YouTube Analytics group/group-item deletes require reverse ETL plan, preview, explicit approval, and typed destructive confirmation.

## Commands

### Inspect as a manual

```bash
pm connectors inspect youtube-analytics
```

### Inspect as structured JSON

```bash
pm connectors inspect youtube-analytics --json
```

## Agent Rules

- Run pm connectors inspect youtube-analytics before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
