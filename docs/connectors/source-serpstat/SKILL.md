---
name: pm-source-serpstat
description: Serpstat connector knowledge and safe action guide.
---

# pm-source-serpstat

## Purpose

Serpstat catalog connector for https://docs.airbyte.com/integrations/sources/serpstat. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-serpstat:0.2.24 (metadata only; not executed)

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

- No upstream application documentation URL was listed in the imported connector registry.
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/serpstat

## Configuration

- api_key (string) required secret: Serpstat API key can be found here: https://serpstat.com/users/profile/
- domain (string): The domain name to get data for (ex. serpstat.com)
- domains (array): The list of domains that will be used in streams that support batch operations
- filter_by (string): The field name by which the results should be filtered. Filtering the results will result in fewer API credits spent. Each stream has different filtering options. See https://se...
- filter_value (string): The value of the field to filter by. Each stream has different filtering options. See https://serpstat.com/api/ for more details.
- page_size (integer): The number of data rows per page to be returned. Each data row can contain multiple data points. The max value is 1000. Reducing the size of the page will result in fewer API cr...
- pages_to_fetch (integer): The number of pages that should be fetched. All results will be obtained if left blank. Reducing the number of pages will result in fewer API credits spent.
- region_id (string): The ID of a region to get data from in the form of a two-letter country code prepended with the g_ prefix. See the list of supported region IDs here: https://serpstat.com/api/66...
- sort_by (string): The field name by which the results should be sorted. Each stream has different sorting options. See https://serpstat.com/api/ for more details.
- sort_value (string): The value of the field to sort by. Each stream has different sorting options. See https://serpstat.com/api/ for more details.
- secret fields: api_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/serpstat

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-serpstat
```

### Inspect as JSON

```bash
pm connectors inspect source-serpstat --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Serpstat documentation](https://docs.airbyte.com/integrations/sources/serpstat)
