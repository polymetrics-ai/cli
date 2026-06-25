---
name: pm-source-smartsheets
description: Smartsheets connector knowledge and safe action guide.
---

# pm-source-smartsheets

## Purpose

Smartsheets catalog connector for https://docs.airbyte.com/integrations/sources/smartsheets. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: beta
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: native_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-smartsheets:1.1.46 (metadata only; not executed)

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

- family: custom_go_port
- priority_wave: 2
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Smartsheet API reference: https://smartsheet.redoc.ly/
- Smartsheet authentication: https://smartsheet.redoc.ly/#section/API-Basics/Authentication-and-Access-Tokens
- Smartsheet rate limits: https://smartsheet.redoc.ly/#section/API-Basics/Rate-Limiting
- Smartsheet Status: https://status.smartsheet.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/smartsheets

## Configuration

- credentials (object) required
- is_report (boolean): If true, the source will treat the provided sheet_id as a report. If false, the source will treat the provided sheet_id as a sheet.
- metadata_fields (array): A List of available columns which metadata can be pulled from.
- spreadsheet_id (string) required: The spreadsheet ID. Find it by opening the spreadsheet then navigating to File > Properties
- start_datetime (string): Only rows modified after this date/time will be replicated. This should be an ISO 8601 string, for instance: `2000-01-01T13:00:00`
- secret fields: credentials.access_token, credentials.client_id, credentials.client_secret, credentials.refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/smartsheets

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-smartsheets
```

### Inspect as JSON

```bash
pm connectors inspect source-smartsheets --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Smartsheets documentation](https://docs.airbyte.com/integrations/sources/smartsheets)
