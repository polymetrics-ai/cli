---
name: pm-source-onesignal
description: OneSignal connector knowledge and safe action guide.
---

# pm-source-onesignal

## Purpose

OneSignal catalog connector for https://docs.airbyte.com/integrations/sources/onesignal. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-onesignal:1.2.56 (metadata only; not executed)

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

- OneSignal API reference: https://documentation.onesignal.com/reference
- OneSignal authentication: https://documentation.onesignal.com/docs/accounts-and-keys
- OneSignal rate limits: https://documentation.onesignal.com/docs/rate-limits
- OneSignal Status: https://status.onesignal.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/onesignal

## Configuration

- applications (array) required: Applications keys, see the <a href="https://documentation.onesignal.com/docs/accounts-and-keys">docs</a> for more information on how to obtain this data
- outcome_names (string) required: Comma-separated list of names and the value (sum/count) for the returned outcome data. See the <a href="https://documentation.onesignal.com/reference/view-outcomes">docs</a> for...
- start_date (string) required: The date from which you'd like to replicate data for OneSignal API, in the format YYYY-MM-DDT00:00:00Z. All data generated after this date will be replicated.
- user_auth_key (string) required secret: OneSignal User Auth Key, see the <a href="https://documentation.onesignal.com/docs/accounts-and-keys#user-auth-key">docs</a> for more information on how to obtain this key.
- secret fields: applications[].app_api_key, applications[].app_id, user_auth_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/onesignal

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-onesignal
```

### Inspect as JSON

```bash
pm connectors inspect source-onesignal --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [OneSignal documentation](https://docs.airbyte.com/integrations/sources/onesignal)
