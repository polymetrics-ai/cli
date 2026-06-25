---
name: pm-source-news-api
description: News Api connector knowledge and safe action guide.
---

# pm-source-news-api

## Purpose

News Api catalog connector for https://docs.airbyte.com/integrations/sources/news-api. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-news-api:0.2.23 (metadata only; not executed)

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

- News API documentation: https://newsapi.org/docs
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/news-api

## Configuration

- api_key (string) required secret: API Key
- category (string) required: The category you want to get top headlines for.
- country (string) required: The 2-letter ISO 3166-1 code of the country you want to get headlines for. You can't mix this with the sources parameter.
- domains (array): A comma-seperated string of domains (eg bbc.co.uk, techcrunch.com, engadget.com) to restrict the search to.
- end_date (string): A date and optional time for the newest article allowed. This should be in ISO 8601 format.
- exclude_domains (array): A comma-seperated string of domains (eg bbc.co.uk, techcrunch.com, engadget.com) to remove from the results.
- language (string): The 2-letter ISO-639-1 code of the language you want to get headlines for. Possible options: ar de en es fr he it nl no pt ru se ud zh.
- search_in (array): Where to apply search query. Possible values are: title, description, content.
- search_query (string): Search query. See https://newsapi.org/docs/endpoints/everything for information.
- sort_by (string) required: The order to sort the articles in. Possible options: relevancy, popularity, publishedAt.
- sources (array): Identifiers (maximum 20) for the news sources or blogs you want headlines from. Use the `/sources` endpoint to locate these programmatically or look at the sources index: https:...
- start_date (string): A date and optional time for the oldest article allowed. This should be in ISO 8601 format.
- secret fields: api_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/news-api

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-news-api
```

### Inspect as JSON

```bash
pm connectors inspect source-news-api --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [News Api documentation](https://docs.airbyte.com/integrations/sources/news-api)
