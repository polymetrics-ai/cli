# Airbyte To pm Native Porting Guide

Status: implementation guide
Audience: engineers and agents porting connectors into `pm`

## Goal

Replicate upstream connector behavior in Go without running Airbyte connector images or Python/Java/Kotlin/Ruby code at product runtime.

This guide describes how to translate upstream auth, streams, state, and writes into the current `pm` architecture.

## Inputs To Import

For each upstream connector directory, import:

- `metadata.yaml`
- `manifest.yaml` for low-code sources
- `spec.json` or generated JSON Schema
- stream schemas
- acceptance test config
- unit/integration fixture data when available
- connector docs
- relevant source files for custom Python/Java/Kotlin connectors

Store normalized output as generated Go-readable JSON:

```text
internal/connectors/native/specs/<slug>.json
```

Do not store secrets. Test configs that reference secret stores must be converted into redacted fixture configs.

## NativePortSpec

Every connector should compile into this shape:

```json
{
  "slug": "source-stripe",
  "type": "source",
  "runtime_family": "declarative_http_source",
  "auth": [],
  "config_fields": [],
  "secret_fields": [],
  "streams": [],
  "actions": [],
  "state": {},
  "rate_limits": {},
  "conformance": {}
}
```

Required sections:

- `auth`: auth modes, secret paths, token refresh behavior, scopes, and injection rules.
- `streams`: name, path/query/body, method, extractor, schema, primary key, cursor, paginator, slicer, parent streams.
- `actions`: destination writes and reverse ETL actions, not generic write-anywhere behavior.
- `state`: cursor/state format, checkpoint frequency, resumability, CDC position when applicable.
- `conformance`: fixture config, fixture responses, live env vars, expected streams, expected records.

## Auth Translation

Map upstream auth into `pm credentials` fields:

| Upstream pattern | pm auth model |
| --- | --- |
| `BearerAuthenticator` | secret token field injected as `Authorization: Bearer <token>`. |
| API key in header | secret field injected into configured header. |
| API key in query | secret field injected into query param; never logged. |
| Basic auth | username config plus password secret. |
| OAuth2 authorization code | credential stores access/refresh/client secret; runtime refreshes token server-side. |
| OAuth2 client credentials | runtime token source with expiry cache. |
| Session login | check/login step creates short-lived session token; never written to docs/logs. |
| Database password | vault secret used only to build driver config. |
| SSH tunnel key/password | vault secret resolved only inside database/file runtime. |
| Cloud role auth | config role ARN/project/account plus optional secret material. |

Rules:

- Never expose secret values in `--json`, manuals, skills, traces, benchmarks, errors, or previews.
- Do not ask users for secrets in chat.
- Prefer `pm credentials add --from-env field=ENV`.
- OAuth refresh tokens are rotated through vault updates, not workflow history.

## Stream Translation

For each upstream stream, create a `NativeStreamSpec`:

```json
{
  "name": "customers",
  "method": "GET",
  "path": "/v1/customers",
  "primary_key": ["id"],
  "cursor_fields": ["updated"],
  "extractor": {"type": "dpath", "path": ["data"]},
  "paginator": {},
  "slicer": {},
  "schema_ref": "schemas/customers.json"
}
```

Translate:

- low-code `DeclarativeStream` to `NativeStreamSpec`.
- Python stream classes to generated checklists plus hand-written Go stream code.
- database tables to discovered `NativeStreamSpec` entries.
- file streams to path-pattern and format specs.

## Declarative HTTP Mapping

| Upstream low-code concept | pm Go runtime component |
| --- | --- |
| `DeclarativeSource` | `declarative.Source` |
| `CheckStream` | `Check(ctx, cfg)` reads configured stream with limit 1. |
| `HttpRequester` | `httpclient.Requester` |
| `BearerAuthenticator` | `httpclient.BearerAuth` |
| `DefaultPaginator` | `httpclient.Paginator` |
| `CursorPagination` | `httpclient.CursorPaginator` |
| `DpathExtractor` | `declarative.Extractor` |
| `DatetimeBasedCursor` | `declarative.DatetimeCursor` |
| `AddFields`/transforms | `declarative.Transform` |
| response filters | `httpclient.ErrorPolicy` |

Minimum supported expression functions:

- config lookup
- record lookup
- response lookup
- stream partition lookup
- `now_utc`
- date parse/format
- string/number/bool coercion
- simple comparisons and boolean operators

Do not implement a full arbitrary Python/Jinja runtime. Implement only the safe expression subset used by imported manifests and reject unsupported expressions during codegen.

## Custom SaaS Mapping

For code-based connectors like GitHub:

1. Convert `spec.json` into config and secret fields.
2. Convert OAuth metadata into credential setup docs and auth providers.
3. Convert every stream class into a Go stream checklist.
4. Implement stream structs by hand using the current GitHub connector style.
5. Add tests per stream against `httptest` or recorded fixture responses.
6. Add live tests behind explicit env vars.

GitHub parity checklist:

- OAuth and PAT config from upstream spec.
- repository patterns and multiple repository handling.
- branch handling.
- start date handling.
- max rate-limit wait.
- REST streams.
- GraphQL streams.
- reaction streams.
- nested parent-child streams.
- per-stream primary keys and cursor fields.

## Database Mapping

Common source operations:

- build connection options from config and vault secrets.
- validate SSL/tunnel config.
- discover schemas, tables, columns, primary keys, and cursor candidates.
- map database types to pm field types and JSON values.
- generate safe snapshot queries.
- generate cursor incremental queries.
- maintain checkpoint state per stream/table/partition.
- expose query only through SELECT validator.

CDC operations:

- setup validation.
- checkpoint position.
- heartbeat/resume.
- delete semantics.
- retention/outage safety checks.
- cleanup docs for slots/publications where relevant.

Postgres native target:

- snapshot: `SELECT <columns> FROM <schema>.<table> ORDER BY pk/cursor`.
- CDC: logical replication with publication/slot and LSN state.
- config: host, port, database, schemas, username, password, SSL, SSH tunnel, replication method, checkpoint interval, max DB connections.

## File/Object Mapping

Common source operations:

- validate provider auth/config.
- list objects with path prefix/pattern.
- filter by modified time/start date.
- checkpoint continuation token/object version/etag.
- infer or apply schema.
- stream rows from file readers with bounded memory.

S3 native target:

- provider config includes bucket, access key, secret key, role ARN, endpoint, region, path prefix, and start date.
- support anonymous/public bucket reads when no secret is provided.
- support S3-compatible endpoints.

## Destination Mapping

Every destination must implement:

- `Check`
- `Catalog` for write capability/actions.
- `ValidateWrite`
- `DryRunWrite`
- `Write`
- per-record receipts.

Warehouse destinations must implement:

- append
- overwrite
- append_dedup
- overwrite_dedup
- schema creation/evolution
- temp table/staging table strategy
- finalization/commit
- idempotent retry

Postgres destination native target:

- config: host, port, database, schema, username, password, SSL, SSL mode, SSH tunnel, raw data schema, CDC deletion mode, disable type/dedupe, drop cascade, unconstrained number.
- setup: create schemas/namespaces.
- write: choose append, append-truncate, dedup, or dedup-truncate loader.
- type mapping/coercion: JSON to Postgres columns.
- deletion handling: hard or soft deletes based on CDC deletion mode.

## Generated Docs Per Connector

Every connector doc must include:

- current implementation status stage.
- upstream implementation source path.
- auth modes.
- config fields.
- secret field names only.
- stream list with primary keys and cursors.
- supported sync modes.
- destination actions or reverse ETL actions.
- query support.
- CDC support.
- fixture tests.
- live test requirements.
- official app docs.

Generated path:

```text
docs/connectors/<slug>/NATIVE_PORT.md
```

## Conformance Gates

Required before `enabled`:

- spec/config validation.
- check success.
- catalog stream/action parity.
- read fixture success for every required stream family.
- state checkpoint/resume test.
- secret redaction test.
- docs/skill validation.
- live check and at least one live read/write where credentials are available.
- benchmark result with memory bound.

Required before reverse ETL live writes:

- action-specific schema.
- dry run or validation endpoint where available.
- approval token.
- idempotency key.
- per-record receipt.
- rollback/compensation note if API supports it.

## Agent Rules

- Agents may inspect `NativePortSpec`, docs, catalog, and conformance reports.
- Agents may run fixture checks without approval.
- Agents may run live reads only with named credentials and output limits.
- Agents may not execute writes without reverse ETL approval.
- Agents may not request or display secret values.
- Agents may not use generic HTTP, shell, or SQL write tools as connector substitutes.

## Implementation Order

1. Status model correction.
2. `NativePortSpec` generator.
3. Declarative HTTP runtime.
4. Stripe live port.
5. GitHub full parity.
6. Postgres source live snapshot/query/CDC.
7. Postgres destination live append/overwrite/dedup.
8. S3 file runtime.
9. Remaining declarative HTTP manifest-only connectors by generated conformance batches.
10. Remaining database/file/destination/custom connectors.

## References

- Airbyte low-code manifest reference: https://docs.airbyte.com/platform/connector-development/config-based/understanding-the-yaml-file/reference
- Airbyte protocol: https://docs.airbyte.com/platform/understanding-airbyte/airbyte-protocol
- Connector metadata: https://docs.airbyte.com/platform/connector-development/connector-metadata-file
- Stripe source manifest: https://github.com/airbytehq/airbyte/blob/master/airbyte-integrations/connectors/source-stripe/manifest.yaml
- GitHub source implementation: https://github.com/airbytehq/airbyte/tree/master/airbyte-integrations/connectors/source-github
- Postgres source implementation: https://github.com/airbytehq/airbyte/tree/master/airbyte-integrations/connectors/source-postgres
- Postgres destination implementation: https://github.com/airbytehq/airbyte/tree/master/airbyte-integrations/connectors/destination-postgres
- S3 source implementation: https://github.com/airbytehq/airbyte/tree/master/airbyte-integrations/connectors/source-s3
