---
name: pm-source-strava
description: Strava connector knowledge and safe action guide.
---

# pm-source-strava

## Purpose

Strava catalog connector for https://docs.airbyte.com/integrations/sources/strava. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: beta
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-strava:0.3.51 (metadata only; not executed)

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
- priority_wave: 2
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Strava API reference: https://developers.strava.com/docs/reference/
- Strava authentication: https://developers.strava.com/docs/authentication/
- Strava rate limits: https://developers.strava.com/docs/rate-limits/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/strava

## Configuration

- athlete_id (integer) required: The Athlete ID of your Strava developer application.
- auth_type (string)
- client_id (string) required: The Client ID of your Strava developer application.
- client_secret (string) required secret: The Client Secret of your Strava developer application.
- refresh_token (string) required secret: The Refresh Token with the activity: read_all permissions.
- start_date (string) required: UTC date and time. Any data before this date will not be replicated.
- secret fields: client_secret, refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/strava

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-strava
```

### Inspect as JSON

```bash
pm connectors inspect source-strava --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Strava documentation](https://docs.airbyte.com/integrations/sources/strava)
