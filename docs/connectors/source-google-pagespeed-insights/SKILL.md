---
name: pm-source-google-pagespeed-insights
description: Google PageSpeed Insights connector knowledge and safe action guide.
---

# pm-source-google-pagespeed-insights

## Purpose

Google PageSpeed Insights catalog connector for https://docs.airbyte.com/integrations/sources/google-pagespeed-insights. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-google-pagespeed-insights:0.2.52 (metadata only; not executed)

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

- PageSpeed Insights API: https://developers.google.com/speed/docs/insights/v5/get-started
- PageSpeed Insights quotas: https://developers.google.com/speed/docs/insights/v5/get-started#quota
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/google-pagespeed-insights

## Configuration

- api_key (string) secret: Google PageSpeed API Key. See <a href="https://developers.google.com/speed/docs/insights/v5/get-started#APIKey">here</a>. The key is optional - however the API is heavily rate l...
- categories (array) required: Defines which Lighthouse category to run. One or many of: "accessibility", "best-practices", "performance", "pwa", "seo".
- strategies (array) required: The analyses strategy to use. Either "desktop" or "mobile".
- urls (array) required: The URLs to retrieve pagespeed information from. The connector will attempt to sync PageSpeed reports for all the defined URLs. Format: https://(www.)url.domain
- secret fields: api_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/google-pagespeed-insights

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-google-pagespeed-insights
```

### Inspect as JSON

```bash
pm connectors inspect source-google-pagespeed-insights --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Google PageSpeed Insights documentation](https://docs.airbyte.com/integrations/sources/google-pagespeed-insights)
