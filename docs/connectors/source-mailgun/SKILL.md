---
name: pm-source-mailgun
description: Mailgun connector knowledge and safe action guide.
---

# pm-source-mailgun

## Purpose

Mailgun catalog connector for https://docs.airbyte.com/integrations/sources/mailgun. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-mailgun:0.3.54 (metadata only; not executed)

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

- Mailgun API reference: https://documentation.mailgun.com/en/latest/api_reference.html
- Mailgun authentication: https://documentation.mailgun.com/en/latest/api-intro.html#authentication
- Mailgun rate limits: https://documentation.mailgun.com/en/latest/api-intro.html#rate-limiting
- Mailgun Status: https://status.mailgun.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/mailgun

## Configuration

- domain_region (string): Domain region code. 'EU' or 'US' are possible values. The default is 'US'.
- private_key (string) required secret: Primary account API key to access your Mailgun data.
- start_date (string): UTC date and time in the format 2020-10-01 00:00:00. Any data before this date will not be replicated. If omitted, defaults to 3 days ago.
- secret fields: private_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/mailgun

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-mailgun
```

### Inspect as JSON

```bash
pm connectors inspect source-mailgun --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Mailgun documentation](https://docs.airbyte.com/integrations/sources/mailgun)
