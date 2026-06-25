---
name: pm-source-paystack
description: Paystack connector knowledge and safe action guide.
---

# pm-source-paystack

## Purpose

Paystack catalog connector for https://docs.airbyte.com/integrations/sources/paystack. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: beta
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-paystack:1.1.31 (metadata only; not executed)

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

- Paystack API reference: https://paystack.com/docs/api/
- Paystack authentication: https://paystack.com/docs/api/#authentication
- Paystack rate limits: https://paystack.com/docs/api/#rate-limiting
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/paystack

## Configuration

- lookback_window_days (integer): When set, the connector will always reload data from the past N days, where N is the value set here. This is useful if your data is updated after creation.
- secret_key (string) required secret: The Paystack API key (usually starts with 'sk_live_'; find yours <a href="https://dashboard.paystack.com/#/settings/developer">here</a>).
- start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated.
- secret fields: secret_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/paystack

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-paystack
```

### Inspect as JSON

```bash
pm connectors inspect source-paystack --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Paystack documentation](https://docs.airbyte.com/integrations/sources/paystack)
