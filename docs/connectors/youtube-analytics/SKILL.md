---
name: pm-youtube-analytics
description: YouTube Analytics connector knowledge and safe action guide.
---

# pm-youtube-analytics

## Purpose

Reads YouTube Reporting API jobs, report types, report metadata, YouTube Analytics groups and group items, and safely plans documented job/group mutations via the Google OAuth 2.0 refresh-token grant.

## Icon

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

- analytics_base_url
- base_url
- content_owner_id
- created_after
- group_id
- group_ids
- include_system_managed
- job_id
- max_pages
- mine
- mode
- page_size
- report_id
- scopes
- start_time_at_or_after
- start_time_before
- token_url
- client_id (secret)
- client_secret (secret)
- refresh_token (secret)

## ETL Streams

- jobs:
  - primary key: id
  - fields: create_time(), expire_time(), id(), name(), report_type_id(), system_managed()
- job:
  - primary key: id
  - fields: create_time(), expire_time(), id(), name(), report_type_id(), system_managed()
- report_types:
  - primary key: id
  - fields: deprecate_time(), id(), name(), system_managed()
- reports:
  - primary key: id
  - fields: create_time(), download_url(), end_time(), id(), job_expire_time(), job_id(), start_time()
- report:
  - primary key: id
  - fields: create_time(), download_url(), end_time(), id(), job_expire_time(), job_id(), start_time()
- groups:
  - primary key: id
  - fields: etag(), id(), item_count(), item_type(), kind(), published_at(), title()
- group_items:
  - primary key: id
  - fields: etag(), group_id(), id(), kind(), resource_id(), resource_kind()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_job:
  - endpoint: POST /jobs
  - required fields: reportTypeId
  - risk: creates a YouTube Reporting job that schedules future report generation; requires reverse ETL plan, preview, and explicit approval
- delete_job:
  - endpoint: DELETE /jobs/{{ record.job_id }}
  - required fields: job_id
  - risk: deletes a scheduled YouTube Reporting job and stops future report generation; destructive, redacts job_id in previews/errors, and requires reverse ETL plan, preview, explicit approval, and typed confirmation
- create_group:
  - endpoint: POST {{ config.analytics_base_url }}/groups
  - required fields: snippet
  - risk: creates a YouTube Analytics group; requires reverse ETL plan, preview, and explicit approval
- update_group:
  - endpoint: PUT {{ config.analytics_base_url }}/groups
  - required fields: id, snippet
  - risk: updates a YouTube Analytics group's metadata (for example title); redacts group id in write errors and requires reverse ETL plan, preview, and explicit approval
- delete_group:
  - endpoint: DELETE {{ config.analytics_base_url }}/groups?id={{ record.id | urlencode }}
  - required fields: id
  - risk: deletes a YouTube Analytics group; destructive, redacts group id in previews/errors, and requires reverse ETL plan, preview, explicit approval, and typed confirmation
- create_group_item:
  - endpoint: POST {{ config.analytics_base_url }}/groupItems
  - required fields: groupId, resource
  - risk: adds a channel, playlist, video, or asset to a YouTube Analytics group; redacts group/resource identifiers in write errors and requires reverse ETL plan, preview, and explicit approval
- delete_group_item:
  - endpoint: DELETE {{ config.analytics_base_url }}/groupItems?id={{ record.id | urlencode }}
  - required fields: id
  - risk: removes an item from a YouTube Analytics group; destructive, redacts group-item id in previews/errors, and requires reverse ETL plan, preview, explicit approval, and typed confirmation

## Security

- read risk: external YouTube Reporting and YouTube Analytics API reads of reporting jobs, report types, report metadata, groups, and group items
- approval: reverse ETL mutations are available only through plan, preview, explicit approval, and destructive confirmation where required
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Read YouTube Reporting/Analytics metadata and safely plan documented job/group mutations.
- Usage: pm youtube-analytics <command> [flags]
- Source CLI: YouTube Analytics and Reporting APIs (Google Discovery: YouTube Analytics API v2 revision 20260729 and YouTube Reporting API v1 revision 20260729)
- Global flags:
  - --credential (string): Credential name to use for the YouTube request.
  - --connection (string): Credential name alias used only when --credential is omitted; does not resolve pm connections.
  - --config (string_array): Connector config override as key=value; never pass secret values here.
  - --json (boolean): Emit machine-readable JSON output.
  - --limit (integer): Maximum records to emit from stream commands.
  - --preview (boolean): Preview a reverse-ETL write command without making a network mutation.
  - --approve (string): Approval token required to execute a reverse-ETL plan.
  - --confirm (string): Typed confirmation challenge for destructive reverse-ETL writes.
- YouTube Reporting
  - jobs list - Read YouTube Reporting jobs through the declared ETL stream. [intent=etl availability=implemented stream=jobs]; flags: --include-system-managed
  - jobs get - Read one YouTube Reporting job by configured job_id. [intent=etl availability=implemented stream=job]; notes: Pass --config job_id=<job id>; ETL command flags intentionally map only to query parameters, not connector config path variables.
  - jobs create - Plan creation of a YouTube Reporting job. [intent=reverse_etl availability=implemented write=create_job]; approval: Plan first, inspect preview output, then run only with the generated approval token.; risk: Creates a scheduled YouTube Reporting job; requires reverse ETL plan, preview, explicit approval, then execute.; flags: --report-type-id, --name
  - jobs delete - Plan deletion of a YouTube Reporting job with typed destructive confirmation. [intent=reverse_etl availability=implemented write=delete_job]; approval: Plan first, inspect preview output, then run only with the generated approval token and typed --confirm destructive challenge.; risk: Destructive job deletion stops future report generation; redacts job_id from previews/errors; requires reverse ETL plan, preview, explicit approval, and --confirm destructive before execute.; flags: --job-id
  - report-types list - Read YouTube Reporting report types through the declared ETL stream. [intent=etl availability=implemented stream=report_types]; flags: --include-system-managed
  - reports list - Read generated report metadata for the configured YouTube Reporting job. [intent=etl availability=implemented stream=reports]; notes: Pass --config job_id=<job id>; downloaded report bytes are intentionally not emitted by this JSON metadata stream.; flags: --created-after, --start-time-at-or-after, --start-time-before
  - reports get - Read one generated report metadata resource for the configured job_id/report_id. [intent=etl availability=implemented stream=report]; notes: Pass --config job_id=<job id> --config report_id=<report id>; downloaded report bytes remain a blocked binary operation.
  - reports query - Documented YouTube Analytics reports.query provider query operation (blocked by default). [intent=direct_read availability=planned]; notes: Blocked until provider-query/direct-read foundation supports bounded typed cross-host query execution without raw query escape hatches.
  - reports download - Documented YouTube Reporting media.download operation for generated report bytes (blocked by default). [intent=direct_read availability=planned]; notes: Blocked until a bounded binary file-download executor exists with destination path safety, size limits, digest/audit evidence, and explicit approval.
- YouTube Analytics groups
  - groups list - Read YouTube Analytics groups through the declared ETL stream. [intent=etl availability=implemented stream=groups]; flags: --mine, --group-ids
  - groups create - Plan creation of a YouTube Analytics group. [intent=reverse_etl availability=implemented write=create_group]; approval: Plan first, inspect preview output, then run only with the generated approval token.; risk: Creates a YouTube Analytics group; requires reverse ETL plan, preview, explicit approval, then execute.; flags: --title
  - groups update - Plan update of a YouTube Analytics group's title. [intent=reverse_etl availability=implemented write=update_group]; approval: Plan first, inspect preview output, then run only with the generated approval token.; risk: Mutates YouTube Analytics group metadata; redacts group id from write errors; requires reverse ETL plan, preview, explicit approval, then execute.; flags: --id, --title
  - groups delete - Plan deletion of a YouTube Analytics group with typed destructive confirmation. [intent=reverse_etl availability=implemented write=delete_group]; approval: Plan first, inspect preview output, then run only with the generated approval token and typed --confirm destructive challenge.; risk: Destructive group deletion; redacts group id from previews/errors; requires reverse ETL plan, preview, explicit approval, and --confirm destructive before execute.; flags: --id
  - group-items list - Read items in a YouTube Analytics group through the declared ETL stream. [intent=etl availability=implemented stream=group_items]; flags: --group-id
  - group-items create - Plan adding a channel, playlist, video, or asset to a YouTube Analytics group. [intent=reverse_etl availability=implemented write=create_group_item]; approval: Plan first, inspect preview output, then run only with the generated approval token.; risk: Adds an item to a YouTube Analytics group; redacts group/resource identifiers from write errors; requires reverse ETL plan, preview, explicit approval, then execute.; flags: --group-id, --resource-kind, --resource-id
  - group-items delete - Plan removal of an item from a YouTube Analytics group with typed destructive confirmation. [intent=reverse_etl availability=implemented write=delete_group_item]; approval: Plan first, inspect preview output, then run only with the generated approval token and typed --confirm destructive challenge.; risk: Destructive removal of a group item; redacts group-item id from previews/errors; requires reverse ETL plan, preview, explicit approval, and --confirm destructive before execute.; flags: --id
- Help topics:
  - binary-downloads - Generated report metadata includes download_url, but media.download report bytes remain blocked until a bounded binary file-download executor exists.
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
