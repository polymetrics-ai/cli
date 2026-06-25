---
name: pm-source-twitter
description: Twitter connector knowledge and safe action guide.
---

# pm-source-twitter

## Purpose

Twitter catalog connector for https://docs.airbyte.com/integrations/sources/twitter. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: beta
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-twitter:0.3.11 (metadata only; not executed)

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

- Twitter API v2: https://developer.twitter.com/en/docs/twitter-api
- Twitter authentication: https://developer.twitter.com/en/docs/authentication/overview
- Twitter rate limits: https://developer.twitter.com/en/docs/twitter-api/rate-limits
- Twitter API Status: https://api.twitterstat.us/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/twitter

## Configuration

- api_key (string) required secret: App only Bearer Token. See the <a href="https://developer.twitter.com/en/docs/authentication/oauth-2-0/bearer-tokens">docs</a> for more information on how to obtain this token.
- end_date (string): The end date for retrieving tweets must be a minimum of 10 seconds prior to the request time.
- query (string) required: Query for matching Tweets. You can learn how to build this query by reading <a href="https://developer.twitter.com/en/docs/twitter-api/tweets/search/integrate/build-a-query"> bu...
- start_date (string): The start date for retrieving tweets cannot be more than 7 days in the past.
- secret fields: api_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/twitter

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-twitter
```

### Inspect as JSON

```bash
pm connectors inspect source-twitter --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Twitter documentation](https://docs.airbyte.com/integrations/sources/twitter)
