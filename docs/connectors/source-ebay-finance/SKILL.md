---
name: pm-source-ebay-finance
description: Ebay Finance connector knowledge and safe action guide.
---

# pm-source-ebay-finance

## Purpose

Ebay Finance catalog connector for https://docs.airbyte.com/integrations/sources/ebay-finance. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-ebay-finance:0.0.39 (metadata only; not executed)

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

- eBay Finances API: https://developer.ebay.com/api-docs/sell/finances/overview.html
- eBay authentication: https://developer.ebay.com/api-docs/static/oauth-tokens.html
- eBay rate limits: https://developer.ebay.com/support/app-check
- eBay Developer Status: https://developer.ebay.com/support/api-status
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/ebay-finance

## Configuration

- api_host (string) required: https://apiz.sandbox.ebay.com for sandbox & https://apiz.ebay.com for production
- client_access_token (string) secret
- password (string) secret: Ebay Client Secret
- redirect_uri (string) required
- refresh_token (string) required secret
- start_date (string) required
- token_refresh_endpoint (string) required
- username (string) required: Ebay Developer Client ID
- secret fields: client_access_token, password, refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/ebay-finance

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-ebay-finance
```

### Inspect as JSON

```bash
pm connectors inspect source-ebay-finance --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Ebay Finance documentation](https://docs.airbyte.com/integrations/sources/ebay-finance)
