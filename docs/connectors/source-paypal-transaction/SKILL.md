---
name: pm-source-paypal-transaction
description: Paypal Transaction connector knowledge and safe action guide.
---

# pm-source-paypal-transaction

## Purpose

Paypal Transaction catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/paypal.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.paypal.com/api/rest/

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

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
- priority_wave: 1
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- PayPal API reference: https://developer.paypal.com/api/rest/
- PayPal authentication: https://developer.paypal.com/api/rest/authentication/
- PayPal rate limits: https://developer.paypal.com/api/rest/rate-limiting/
- PayPal Status: https://www.paypal-status.com/

## Configuration

- client_id (string) required secret: The Client ID of your Paypal developer application.
- client_secret (string) required secret: The Client Secret of your Paypal developer application.
- dispute_start_date (string): Start Date parameter for the list dispute endpoint in <a href=\"https://datatracker.ietf.org/doc/html/rfc3339#section-5.6\">ISO format</a>. This Start Date must be in range with...
- end_date (string): End Date for data extraction in <a href=\"https://datatracker.ietf.org/doc/html/rfc3339#section-5.6\">ISO format</a>. This can be help you select specific range of time, mainly ...
- is_sandbox (boolean) required: Determines whether to use the sandbox or production environment.
- refresh_token (string) secret: The key to refresh the expired access token.
- start_date (string) required: Start Date for data extraction in <a href=\"https://datatracker.ietf.org/doc/html/rfc3339#section-5.6\">ISO format</a>. Date must be in range from 3 years till 12 hrs before pre...
- time_window (integer): The number of days per request. Must be a number between 1 and 31.
- secret fields: client_id, client_secret, refresh_token

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
pm connectors inspect source-paypal-transaction
```

### Inspect as JSON

```bash
pm connectors inspect source-paypal-transaction --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [PayPal API reference](https://developer.paypal.com/api/rest/)
- [PayPal authentication](https://developer.paypal.com/api/rest/authentication/)
- [PayPal rate limits](https://developer.paypal.com/api/rest/rate-limiting/)
- [PayPal Status](https://www.paypal-status.com/)
