---
name: pm-source-okta
description: Okta connector knowledge and safe action guide.
---

# pm-source-okta

## Purpose

Okta catalog connector for https://docs.airbyte.com/integrations/sources/okta. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: native_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-okta:0.3.21 (metadata only; not executed)

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
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Okta API reference: https://developer.okta.com/docs/reference/
- Okta authentication: https://developer.okta.com/docs/guides/implement-oauth-for-okta/main/
- Okta rate limits: https://developer.okta.com/docs/reference/rate-limits/
- Okta Status: https://status.okta.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/okta

## Configuration

- credentials (object)
- domain (string): The Okta domain. See the <a href="https://docs.airbyte.com/integrations/sources/okta">docs</a> for instructions on how to find it.
- start_date (string): UTC date and time in the format YYYY-MM-DDTHH:MM:SSZ. Any data before this date will not be replicated.
- secret fields: credentials.api_token, credentials.client_id, credentials.client_secret, credentials.key_id, credentials.private_key, credentials.refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/okta

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-okta
```

### Inspect as JSON

```bash
pm connectors inspect source-okta --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Okta documentation](https://docs.airbyte.com/integrations/sources/okta)
