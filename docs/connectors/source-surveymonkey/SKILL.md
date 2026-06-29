---
name: pm-source-surveymonkey
description: SurveyMonkey connector knowledge and safe action guide.
---

# pm-source-surveymonkey

## Purpose

SurveyMonkey catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/surveymonkey.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.surveymonkey.com/api/v3/

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
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
- priority_wave: 1
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- SurveyMonkey API reference: https://developer.surveymonkey.com/api/v3/
- SurveyMonkey authentication: https://developer.surveymonkey.com/api/v3/#authentication
- SurveyMonkey API Changelog: https://developer.surveymonkey.com/api/v3/#changelog
- SurveyMonkey rate limits: https://developer.surveymonkey.com/api/v3/#rate-limits

## Configuration

- credentials (object) required: The authorization method to use to retrieve data from SurveyMonkey
- origin (string): Depending on the originating datacenter of the SurveyMonkey account, the API access URL may be different.
- start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated.
- survey_ids (array): IDs of the surveys from which you'd like to replicate data. If left empty, data from all boards to which you have access will be replicated.
- secret fields: credentials.access_token, credentials.client_id, credentials.client_secret

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
pm connectors inspect source-surveymonkey
```

### Inspect as JSON

```bash
pm connectors inspect source-surveymonkey --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [SurveyMonkey API reference](https://developer.surveymonkey.com/api/v3/)
- [SurveyMonkey authentication](https://developer.surveymonkey.com/api/v3/#authentication)
- [SurveyMonkey API Changelog](https://developer.surveymonkey.com/api/v3/#changelog)
- [SurveyMonkey rate limits](https://developer.surveymonkey.com/api/v3/#rate-limits)
