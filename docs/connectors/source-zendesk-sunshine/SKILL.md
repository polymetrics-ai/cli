---
name: pm-source-zendesk-sunshine
description: Zendesk Sunshine connector knowledge and safe action guide.
---

# pm-source-zendesk-sunshine

## Purpose

Zendesk Sunshine catalog connector for https://docs.airbyte.com/integrations/sources/zendesk-sunshine. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-zendesk-sunshine:0.4.11 (metadata only; not executed)

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

- Zendesk Sunshine API: https://developer.zendesk.com/api-reference/custom-data/custom-objects-api/introduction/
- Zendesk authentication: https://developer.zendesk.com/api-reference/ticketing/introduction/#security-and-authentication
- Zendesk Status: https://status.zendesk.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/zendesk-sunshine

## Configuration

- credentials (object): Zendesk allows three authentication methods. We recommend using `OAuth2.0` for Airbyte Cloud users and `API token` for Airbyte Open Source users.
- start_date (string) required: The date from which you'd like to replicate data for Zendesk Sunshine API, in the format YYYY-MM-DDT00:00:00Z.
- subdomain (string) required: The subdomain for your Zendesk Account.
- secret fields: credentials.access_token, credentials.api_token, credentials.client_id, credentials.client_secret, credentials.refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/zendesk-sunshine

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-zendesk-sunshine
```

### Inspect as JSON

```bash
pm connectors inspect source-zendesk-sunshine --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Zendesk Sunshine documentation](https://docs.airbyte.com/integrations/sources/zendesk-sunshine)
