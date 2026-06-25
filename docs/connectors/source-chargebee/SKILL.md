---
name: pm-source-chargebee
description: Chargebee connector knowledge and safe action guide.
---

# pm-source-chargebee

## Purpose

Chargebee catalog connector for https://docs.airbyte.com/integrations/sources/chargebee. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-chargebee:0.10.38 (metadata only; not executed)

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
- priority_wave: 1
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Versioning Docs: https://apidocs.chargebee.com/docs/api/versioning
- Changelog: https://www.chargebee.com/help/api-updates/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/chargebee

## Configuration

- num_workers (integer): The number of worker threads to use for the sync. The performance upper boundary is based on the limit of your Chargebee plan. More info about the rate limit plan tiers can be f...
- product_catalog (string): Product Catalog version of your Chargebee site. Instructions on how to find your version you may find <a href="https://apidocs.chargebee.com/docs/api?prod_cat_ver=2">here</a> un...
- site (string) required: The site prefix for your Chargebee instance.
- site_api_key (string) required secret: Chargebee API Key. See the <a href="https://docs.airbyte.com/integrations/sources/chargebee">docs</a> for more information on how to obtain this key.
- start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00.000Z. Any data before this date will not be replicated.
- secret fields: site_api_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/chargebee

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-chargebee
```

### Inspect as JSON

```bash
pm connectors inspect source-chargebee --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Chargebee documentation](https://docs.airbyte.com/integrations/sources/chargebee)
