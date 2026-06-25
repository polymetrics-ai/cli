---
name: pm-source-pagerduty
description: Pagerduty connector knowledge and safe action guide.
---

# pm-source-pagerduty

## Purpose

Pagerduty catalog connector for https://docs.airbyte.com/integrations/sources/pagerduty. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-pagerduty:0.3.41 (metadata only; not executed)

## Runtime Capabilities

- metadata=true
- check=false
- catalog=false
- read=false
- write=false
- query=false
- etl=false
- reverse_etl=false
- unsupported_reason: Native Go port is planned but not enabled; only catalog metadata is available.

## Native Port Plan

- family: declarative_http_source
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- PagerDuty API reference: https://developer.pagerduty.com/api-reference/
- PagerDuty authentication: https://developer.pagerduty.com/docs/ZG9jOjExMDI5NTUw-authentication
- PagerDuty rate limits: https://developer.pagerduty.com/docs/ZG9jOjExMDI5NTUx-rate-limiting
- PagerDuty Status: https://status.pagerduty.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/pagerduty

## Configuration

- cutoff_days (integer): Fetch pipelines updated in the last number of days
- default_severity (string): A default severity category if not present
- exclude_services (array): List of PagerDuty service names to ignore incidents from. If not set, all incidents will be pulled.
- incident_log_entries_overview (boolean): If true, will return a subset of log entries that show only the most important changes to the incident.
- max_retries (integer): Maximum number of PagerDuty API request retries to perform upon connection errors. The source will pause for an exponentially increasing number of seconds before retrying.
- page_size (integer): page size to use when querying PagerDuty API
- service_details (array): List of PagerDuty service additional details to include.
- token (string) required secret: API key for PagerDuty API authentication
- secret fields: token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/pagerduty

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-pagerduty
```

### Inspect as JSON

```bash
pm connectors inspect source-pagerduty --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Pagerduty documentation](https://docs.airbyte.com/integrations/sources/pagerduty)
