---
name: pm-source-iterable
description: Iterable connector knowledge and safe action guide.
---

# pm-source-iterable

## Purpose

Iterable catalog connector for https://docs.airbyte.com/integrations/sources/iterable. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: native_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-iterable:0.7.2 (metadata only; not executed)

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
- priority_wave: 1
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Iterable API reference: https://api.iterable.com/api/docs
- Iterable authentication: https://support.iterable.com/hc/en-us/articles/360043464871-API-Keys-
- Iterable rate limits: https://support.iterable.com/hc/en-us/articles/360045714132-API-Rate-Limits
- Iterable Status: https://status.iterable.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/iterable

## Configuration

- api_key (string) required secret: Iterable API Key. See the <a href=\"https://docs.airbyte.com/integrations/sources/iterable\">docs</a> for more information on how to obtain this key.
- lookback_window (integer): Number of minutes to re-read from the current time when determining the end of each sync window for export-based streams. This accounts for eventual consistency delays in Iterab...
- region (string): The region where your Iterable account is hosted. Select 'EU' if your account is on the European data center.
- start_date (string) required: The date from which you'd like to replicate data for Iterable, in the format YYYY-MM-DDT00:00:00Z. All data generated after this date will be replicated.
- secret fields: api_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/iterable

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-iterable
```

### Inspect as JSON

```bash
pm connectors inspect source-iterable --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Iterable documentation](https://docs.airbyte.com/integrations/sources/iterable)
