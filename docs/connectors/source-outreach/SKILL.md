---
name: pm-source-outreach
description: Outreach connector knowledge and safe action guide.
---

# pm-source-outreach

## Purpose

Outreach catalog connector for https://docs.airbyte.com/integrations/sources/outreach. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-outreach:1.1.34 (metadata only; not executed)

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

- Outreach API reference: https://api.outreach.io/api/v2/docs
- Outreach authentication: https://api.outreach.io/api/v2/docs#authentication
- Outreach rate limits: https://api.outreach.io/api/v2/docs#rate-limiting
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/outreach

## Configuration

- client_id (string) required: The Client ID of your Outreach developer application.
- client_secret (string) required secret: The Client Secret of your Outreach developer application.
- redirect_uri (string) required: A Redirect URI is the location where the authorization server sends the user once the app has been successfully authorized and granted an authorization code or access token.
- refresh_token (string) required secret: The token for obtaining the new access token.
- start_date (string) required: The date from which you'd like to replicate data for Outreach API, in the format YYYY-MM-DDT00:00:00.000Z. All data generated after this date will be replicated.
- secret fields: client_secret, refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/outreach

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-outreach
```

### Inspect as JSON

```bash
pm connectors inspect source-outreach --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Outreach documentation](https://docs.airbyte.com/integrations/sources/outreach)
