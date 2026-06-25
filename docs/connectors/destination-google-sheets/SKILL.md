---
name: pm-destination-google-sheets
description: Google Sheets connector knowledge and safe action guide.
---

# pm-destination-google-sheets

## Purpose

Google Sheets catalog connector for https://docs.airbyte.com/integrations/destinations/google-sheets. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: destination_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/destination-google-sheets:0.3.5 (metadata only; not executed)

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

- family: destination_writer
- priority_wave: 3
- etl_operations: catalog, check, write_append, write_dedup, write_overwrite
- reverse_etl_operations: none until native write conformance passes
- conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

## Official Application Documentation

- No upstream application documentation URL was listed in the imported connector registry.
- Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/google-sheets

## Configuration

- credentials (object) required: Authentication method to access Google Sheets
- spreadsheet_id (string) required: The link to your spreadsheet. See <a href='https://docs.airbyte.com/integrations/destinations/google-sheets#sheetlink'>this guide</a> for more details.
- secret fields: credentials.client_id, credentials.client_secret, credentials.refresh_token, credentials.service_account_info

## Sync Modes

- supported sync modes: append, append_dedup, overwrite
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/destinations/google-sheets

## Commands

### Inspect catalog entry

```bash
pm connectors inspect destination-google-sheets
```

### Inspect as JSON

```bash
pm connectors inspect destination-google-sheets --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Google Sheets documentation](https://docs.airbyte.com/integrations/destinations/google-sheets)
