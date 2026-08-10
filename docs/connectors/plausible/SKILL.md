---
name: pm-plausible
description: Plausible connector knowledge and safe action guide.
---

# pm-plausible

## Purpose

Reads Plausible Analytics sites and stats reports through the Stats API.

## Icon

- id: plausible
- asset: icons/plausible.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://plausible.io/docs/stats-api

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- compare
- date
- filters
- metrics
- mode
- period
- property
- site_id
- api_token (secret)

## ETL Streams

- sites:
  - primary key: site_id
  - fields: domain(string), site_id(string)
- aggregate:
  - primary key: site_id
  - fields: bounce_rate(number), events(integer), pageviews(integer), site_id(string), visit_duration(number), visitors(integer), visits(integer)
- timeseries:
  - primary key: date
  - fields: bounce_rate(number), date(string), events(integer), pageviews(integer), site_id(string), visit_duration(number), visitors(integer), visits(integer)
- breakdown:
  - primary key: property_value
  - fields: bounce_rate(number), events(integer), pageviews(integer), property_value(string), site_id(string), visit_duration(number), visitors(integer), visits(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Plausible Analytics API read of site analytics data
- approval: none; read-only analytics sync
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Plausible's declared streams and reverse-ETL actions.
- Usage: pm plausible <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - aggregate list - Run the aggregate ETL stream [intent=etl availability=implemented stream=aggregate]; notes: discrepancy=present-in-surface-absent-from-artifact
  - api delete v1 sites - Documented DELETE /v1/sites (not implemented) [intent=direct_write availability=not_implemented operation=plausible.delete.v1-sites]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get api v1 stats aggregate - Documented GET /api/v1/stats/aggregate (not implemented) [intent=direct_read availability=not_implemented operation=plausible.get.api-v1-stats-aggregate]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 stats breakdown - Documented GET /api/v1/stats/breakdown (not implemented) [intent=direct_read availability=not_implemented operation=plausible.get.api-v1-stats-breakdown]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 stats realtime visitors - Documented GET /api/v1/stats/realtime/visitors (not implemented) [intent=direct_read availability=not_implemented operation=plausible.get.api-v1-stats-realtime-visitors]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 stats timeseries - Documented GET /api/v1/stats/timeseries (not implemented) [intent=direct_read availability=not_implemented operation=plausible.get.api-v1-stats-timeseries]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 stats goal breakdown - Documented GET /v1/stats/goal/breakdown (not implemented) [intent=direct_read availability=not_implemented operation=plausible.get.v1-stats-goal-breakdown]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 stats realtime visitors - Documented GET /v1/stats/realtime/visitors (not implemented) [intent=direct_read availability=not_implemented operation=plausible.get.v1-stats-realtime-visitors]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post v1 sites - Documented POST /v1/sites (not implemented) [intent=direct_write availability=not_implemented operation=plausible.post.v1-sites]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - breakdown list - Run the breakdown ETL stream [intent=etl availability=implemented stream=breakdown]; notes: discrepancy=present-in-surface-absent-from-artifact
  - sites list - Run the sites ETL stream [intent=etl availability=implemented stream=sites]; notes: discrepancy=present-in-surface-absent-from-artifact
  - timeseries list - Run the timeseries ETL stream [intent=etl availability=implemented stream=timeseries]; notes: discrepancy=present-in-surface-absent-from-artifact

## Commands

### Inspect as a manual

```bash
pm connectors inspect plausible
```

### Inspect as structured JSON

```bash
pm connectors inspect plausible --json
```

## Agent Rules

- Run pm connectors inspect plausible before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
