---
name: pm-source-trustpilot
description: TrustPilot connector knowledge and safe action guide.
---

# pm-source-trustpilot

## Purpose

TrustPilot catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/trustpilot.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.trustpilot.com/

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

- Trustpilot API documentation: https://developers.trustpilot.com/

## Configuration

- business_units (array) required: The names of business units which shall be synchronized. Some streams e.g. configured_business_units or private_reviews use this configuration.
- credentials (object) required
- start_date (string) required: For streams with sync. method incremental the start date time to be used
- secret fields: credentials.access_token, credentials.client_id, credentials.client_secret, credentials.refresh_token

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
pm connectors inspect source-trustpilot
```

### Inspect as JSON

```bash
pm connectors inspect source-trustpilot --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Trustpilot API documentation](https://developers.trustpilot.com/)
