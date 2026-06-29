---
name: pm-source-microsoft-teams
description: Microsoft Teams connector knowledge and safe action guide.
---

# pm-source-microsoft-teams

## Purpose

Microsoft Teams catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/microsoft-teams.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://learn.microsoft.com/en-us/graph/api/resources/teams-api-overview

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.

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

- Microsoft Teams API: https://learn.microsoft.com/en-us/graph/api/resources/teams-api-overview
- Microsoft Graph authentication: https://learn.microsoft.com/en-us/graph/auth/
- Microsoft Graph throttling: https://learn.microsoft.com/en-us/graph/throttling
- Microsoft 365 Status: https://status.office365.com/

## Configuration

- credentials (object): Choose how to authenticate to Microsoft
- period (string) required: Specifies the length of time over which the Team Device Report stream is aggregated. The supported values are: D7, D30, D90, and D180.
- secret fields: credentials.client_secret, credentials.refresh_token, credentials.tenant_id

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-microsoft-teams
```

### Inspect as JSON

```bash
pm connectors inspect source-microsoft-teams --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Microsoft Teams API](https://learn.microsoft.com/en-us/graph/api/resources/teams-api-overview)
- [Microsoft Graph authentication](https://learn.microsoft.com/en-us/graph/auth/)
- [Microsoft Graph throttling](https://learn.microsoft.com/en-us/graph/throttling)
- [Microsoft 365 Status](https://status.office365.com/)
