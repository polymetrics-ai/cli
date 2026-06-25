---
name: pm-source-exchange-rates
description: Exchange Rates Api connector knowledge and safe action guide.
---

# pm-source-exchange-rates

## Purpose

Exchange Rates Api catalog connector for https://docs.airbyte.com/integrations/sources/exchange-rates. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-exchange-rates:1.4.53 (metadata only; not executed)

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

- Exchange Rates API documentation: https://exchangeratesapi.io/documentation/
- Exchange Rates authentication: https://exchangeratesapi.io/documentation/#authentication
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/exchange-rates

## Configuration

- access_key (string) required secret: Your API Key. See <a href="https://apilayer.com/marketplace/exchangerates_data-api">here</a>. The key is case sensitive.
- base (string): ISO reference currency. See <a href="https://www.ecb.europa.eu/stats/policy_and_exchange_rates/euro_reference_exchange_rates/html/index.en.html">here</a>. Free plan doesn't supp...
- ignore_weekends (boolean): Ignore weekends? (Exchanges don't run on weekends)
- start_date (string) required: Start getting data from that date.
- secret fields: access_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/exchange-rates

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-exchange-rates
```

### Inspect as JSON

```bash
pm connectors inspect source-exchange-rates --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Exchange Rates Api documentation](https://docs.airbyte.com/integrations/sources/exchange-rates)
