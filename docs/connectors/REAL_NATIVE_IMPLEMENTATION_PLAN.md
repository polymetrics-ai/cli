# Real Native Connector Implementation Plan

Status: implementation plan
Owner: Polymetrics engineering
Scope: all 647 catalog connectors

## Why This Exists

The current `pm` connector catalog has Go-native fixture bindings for every catalog slug. That gives every connector a safe CLI/documentation/conformance surface, but it is not the same as a real live native port.

A real native implementation means:

- `check` validates the real external system or database.
- `catalog` discovers real streams/tables/actions from the connector's upstream implementation and schema.
- `read` extracts real records with the upstream connector's auth, pagination, slicing, cursor, schema, and error semantics.
- `write` mutates only explicitly modeled destinations/actions through preview, approval, idempotency, and receipts.
- `query` is enabled only for systems where the connector can validate and execute safe read-only queries.
- Conformance includes fixture-backed tests and, where credentials exist, live sandbox tests.

The current fixture binding must remain as a scaffold, but it must not be confused with a production live native port.

## Status Model Correction

Replace the single `implementation_status=enabled` meaning with a staged status model:

| Status | Meaning | Runtime allowed |
| --- | --- | --- |
| `catalog_imported` | Metadata, docs URL, schema, secret fields imported. | Inspect only |
| `native_fixture_binding` | Safe Go binding exists with fixture check/catalog/read/write/query. | Fixture mode only |
| `runtime_family_ported` | Shared runtime family supports this connector's upstream pattern. | Fixture and selected live tests |
| `connector_live_ported` | Connector-specific auth, streams, state, and writes are implemented. | Live mode behind conformance |
| `live_conformance_passed` | Live/sandbox tests pass for all required operations. | Runnable |
| `enabled` | Connector is production-ready in live mode. | Runnable by default |

Acceptance rule: never use `enabled` for a connector that only has a generic fixture binding.

## Architecture Target

```text
catalog_data.json
  -> upstream source importer
  -> native port spec
  -> runtime family compiler
  -> generated connector binding
  -> conformance suite
  -> docs/manual/skill
  -> enabled live connector
```

Packages to add:

- `internal/connectors/native/spec`: normalized native port spec for auth, streams, state, schemas, actions, and tests.
- `internal/connectors/native/declarative`: Go runtime for manifest-only HTTP API sources.
- `internal/connectors/native/httpclient`: request builder, auth injection, retry, pagination, rate limits, and response decoding.
- `internal/connectors/native/database`: SQL snapshot, cursor incremental, CDC, query validator, and type mapping.
- `internal/connectors/native/fileobject`: local, S3, GCS, Azure Blob, SFTP, file formats, compression, and checkpointed listings.
- `internal/connectors/native/destination`: append, overwrite, dedup/upsert, schema evolution, batching, idempotency, and receipts.
- `internal/connectors/native/custom`: hand-written Go ports for connectors that cannot be represented by shared runtime specs.
- `internal/connectors/native/conformance`: shared connector acceptance tests.

## Upstream Implementation Families

Current catalog split:

| Runtime family | Count | Port strategy |
| --- | ---: | --- |
| `declarative_http_go` | 503 | Import `manifest.yaml`; run through Go declarative HTTP interpreter. |
| `database_go` | 25 | Implement SQL snapshot/query plus database-specific CDC adapters. |
| `file_go` | 11 | Implement object listing, file format readers, schema inference, and checkpointing. |
| `destination_go` | 56 | Implement destination writer runtime with append/overwrite/dedup and safe reverse ETL actions. |
| `native_go` | 52 | Port custom Python/Java/Kotlin connector code into hand-written Go adapters. |

## Declarative HTTP Runtime

The Stripe source is the reference for manifest-only connectors. Its upstream manifest uses:

- `DeclarativeSource`
- `CheckStream`
- reusable `definitions`
- `HttpRequester`
- `BearerAuthenticator`
- request headers such as API version and account ID
- `RecordSelector` with `DpathExtractor`
- `DefaultPaginator`
- `CursorPagination` using response-derived cursor values
- `DatetimeBasedCursor`
- stream `path` templates
- primary keys and cursor fields
- response filters for ignored HTTP codes

Go implementation tasks:

1. Parse YAML into typed structs with raw extension fields preserved.
2. Resolve `$ref` pointers and merge overrides deterministically.
3. Implement interpolation for `{{ config[...] }}`, `stream_partition`, `response`, `record`, `now_utc`, and date formatting macros used by imported manifests.
4. Implement authenticators:
   - bearer token
   - basic auth
   - API key in header/query/body
   - OAuth2 refresh/client credentials where spec provides token URL
   - session/token exchange authenticators
5. Implement request building:
   - `url_base`
   - path templates
   - headers
   - query/body injection
   - stream slices
   - parent-child partitions
6. Implement paginators:
   - cursor pagination
   - page-number pagination
   - offset/limit pagination
   - link-header pagination
   - stop conditions
7. Implement selectors and transforms:
   - dpath extraction
   - add/remove fields
   - schema normalization
   - deleted/tombstone markers
8. Implement cursors:
   - datetime cursor
   - integer cursor
   - nested cursor path
   - lookback window
   - step/window slicing
9. Add golden tests from representative manifests:
   - Stripe for bearer auth, cursor pagination, date slicing, child streams.
   - Calendly or Slack for token auth and simple pagination.
   - Shopify or Zendesk for parent-child streams and rate limits.

## Custom SaaS Runtime

The GitHub upstream connector is code-based. Its source has:

- `spec.json` with OAuth and Personal Access Token auth modes.
- config fields for repositories, start date, API URL, branches, and max waiting time.
- `advanced_auth` OAuth settings and scopes.
- `streams.py` classes for REST and GraphQL streams.
- per-stream path, primary key, cursor field, request params, stream slices, and pagination.
- custom error handlers for auth failures, SAML/SSO, and rate limits.

Our Go GitHub connector already implements part of this pattern. The real native port work is to close parity gaps by importing the upstream stream matrix and adding every missing stream/action.

Custom port process:

1. Import upstream `spec.json`, schemas, and stream source files into a normalized `NativePortSpec`.
2. Generate a Go checklist for every stream class:
   - stream name
   - endpoint path
   - HTTP method
   - primary key
   - cursor field
   - pagination
   - parent stream dependencies
   - REST vs GraphQL
3. Implement stream functions in Go only when the checklist has tests.
4. Implement OAuth/PAT/GitHub App mappings in `pm credentials`.
5. Add live tests behind env vars; keep fixture tests always on.

## Database Runtime

The Postgres upstream source is Kotlin and delegates to a CDK runner. The source implementation includes:

- config spec for host, port, database, schemas, username, password, SSL mode, SSH tunnel, JDBC params, replication method, checkpoint interval, and max DB connections.
- metadata querier.
- select query generator.
- JDBC partition creation and stream state.
- Postgres type mapping.
- Debezium/logical replication operations.
- CDC position and CDC metadata fields.

Go implementation tasks:

1. Build common SQL source runtime:
   - connection config and secret mapping
   - SSL/tunnel options
   - schema/table discovery
   - column discovery and type mapping
   - safe SELECT query generation
   - chunked snapshot reads
   - cursor incremental reads
   - checkpointing by table/partition/cursor
2. Build query runtime:
   - SELECT-only parser/validator
   - identifier quoting
   - row and byte limits
   - timeout and cancellation
3. Build CDC adapters:
   - Postgres logical replication: publication, replication slot, LSN checkpoint, delete semantics.
   - MySQL/TiDB binlog: GTID or file/position, row image validation, server ID.
   - MongoDB change streams: resume token, cluster time, pre-image caveats.
   - SQL Server CDC: capture instance and LSN retention validation.
   - Oracle: SCN, supplemental logging, LogMiner/XStream boundary.
4. Add Podman-first integration tests with GHCR images where possible.

## File/Object Runtime

The S3 source is file-based. Its upstream spec includes:

- provider object for S3/AWS.
- bucket.
- access key and secret key.
- role ARN.
- path prefix.
- endpoint for S3-compatible services.
- region.
- start date filter.
- inherited file-format and stream options from the file-based CDK.

Go implementation tasks:

1. Build object provider interface:
   - local filesystem
   - S3/S3-compatible
   - GCS
   - Azure Blob
   - SFTP
2. Build listing checkpoint:
   - path prefix/pattern
   - modified time filter
   - object version/etag
   - continuation token
3. Build readers:
   - CSV
   - JSON
   - JSONL
   - Parquet
   - gzip/bzip2/zstd where required
4. Build schema inference and user-provided schema support.
5. Add fixture buckets/directories and integration tests.

## Destination Runtime

The Postgres destination is the reference destination writer. Its upstream Kotlin writer:

- creates namespaces/schemas.
- gathers initial table status.
- maps input column names to final column names.
- generates temp table names.
- selects append, append-truncate, dedup, or dedup-truncate loaders.
- supports raw/final table behavior.
- handles schema evolution and value coercion.

Go implementation tasks:

1. Build common destination writer interfaces:
   - `Setup`
   - `CreateStreamLoader`
   - `ValidateWrite`
   - `DryRunWrite`
   - `WriteBatch`
   - `Finalize`
2. Implement sync modes:
   - append
   - overwrite
   - append_dedup
   - overwrite_dedup
3. Implement warehouse destinations:
   - Postgres
   - BigQuery
   - Snowflake
   - Redshift
   - ClickHouse
4. Implement object/vector destinations:
   - S3/GCS/Azure Blob
   - Kafka/PubSub/RabbitMQ
   - Pinecone/Weaviate/Milvus/PGVector
5. Keep reverse ETL writes action-specific. Do not expose generic HTTP or generic SQL write to agents.

## Connector Port Workflow

For each connector:

1. Fetch upstream connector directory metadata, spec, manifest, schemas, tests, and docs.
2. Classify into runtime family.
3. Generate `NativePortSpec`.
4. Generate tests first:
   - spec/config test
   - secret-field redaction test
   - check test
   - catalog/stream test
   - read fixture test
   - state checkpoint test
   - write/dry-run test for destinations
   - query safety test for query-capable systems
   - docs/skill test
5. Implement runtime family support or connector-specific adapter.
6. Add fixture recordings and live sandbox env-var test gate.
7. Run conformance.
8. Move status forward only if the gate passes.

## Acceptance Gates

Catalog-wide:

- Every connector has status at least `native_fixture_binding`.
- Every connector has docs with auth, config, streams/actions, sync modes, security, and upstream docs.
- Every connector has a generated `NativePortSpec`.

Runtime-family:

- Every declarative manifest-only source can parse and render catalog streams.
- Every database connector can check, catalog, snapshot, query, and checkpoint in fixture/integration tests.
- Every file connector can list, infer schema, read records, and checkpoint.
- Every destination can dry-run and write fixture batches for all supported sync modes.

Live enablement:

- Connector-specific live `check`, `catalog`, and at least one `read` or `write` path pass.
- Secrets are never printed.
- Reverse ETL writes are plan/preview/approve/run only.
- Benchmarks record throughput and memory bounds.

## Immediate Next Milestones

1. Correct implementation status semantics in code and docs.
2. Add `NativePortSpec` model and generator.
3. Import upstream manifest/spec/schema files for all catalog connectors into generated artifacts.
4. Implement declarative HTTP runtime enough to make Stripe live.
5. Expand GitHub to full upstream stream/auth parity.
6. Implement Postgres source and destination live ports.
7. Implement S3/file runtime.
8. Replace fixture binding with real runtime family binding connector by connector.

## Source References

- Stripe manifest: https://github.com/airbytehq/airbyte/blob/master/airbyte-integrations/connectors/source-stripe/manifest.yaml
- GitHub source: https://github.com/airbytehq/airbyte/tree/master/airbyte-integrations/connectors/source-github
- GitHub spec: https://github.com/airbytehq/airbyte/blob/master/airbyte-integrations/connectors/source-github/source_github/spec.json
- Postgres source: https://github.com/airbytehq/airbyte/tree/master/airbyte-integrations/connectors/source-postgres
- Postgres destination: https://github.com/airbytehq/airbyte/tree/master/airbyte-integrations/connectors/destination-postgres
- S3 source: https://github.com/airbytehq/airbyte/tree/master/airbyte-integrations/connectors/source-s3
