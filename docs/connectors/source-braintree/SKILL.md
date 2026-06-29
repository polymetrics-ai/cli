---
name: pm-source-braintree
description: Braintree connector knowledge and safe action guide.
---

# pm-source-braintree

## Purpose

Braintree catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/braintree.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.paypal.com/braintree/docs/reference/general/server-sdk-deprecation-policy

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: native_go
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

- family: custom_go_port
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Braintree API reference: https://developer.paypal.com/braintree/docs/reference/overview
- Braintree authentication: https://developer.paypal.com/braintree/docs/start/authentication
- Server SDK Deprecation Policy: https://developer.paypal.com/braintree/docs/reference/general/server-sdk-deprecation-policy
- Braintree API rate limits: https://developer.paypal.com/braintree/docs/reference/general/rate-limiting
- Braintree Status: https://status.braintreepayments.com/

## Configuration

- environment (string) required: Environment specifies where the data will come from.
- merchant_id (string) required: manual intervention needed
- private_key (string) required secret: manual intervention needed
- public_key (string) required: manual intervention needed
- start_date (string): UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated.
- secret fields: private_key

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
pm connectors inspect source-braintree
```

### Inspect as JSON

```bash
pm connectors inspect source-braintree --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Braintree API reference](https://developer.paypal.com/braintree/docs/reference/overview)
- [Braintree authentication](https://developer.paypal.com/braintree/docs/start/authentication)
- [Server SDK Deprecation Policy](https://developer.paypal.com/braintree/docs/reference/general/server-sdk-deprecation-policy)
- [Braintree API rate limits](https://developer.paypal.com/braintree/docs/reference/general/rate-limiting)
- [Braintree Status](https://status.braintreepayments.com/)
