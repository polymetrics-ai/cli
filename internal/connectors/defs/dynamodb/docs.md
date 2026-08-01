# Overview

DynamoDB is a Tier-3 native connector because the official AWS JSON APIs are not ordinary REST resources: every operation is a `POST /` request with an `X-Amz-Target` operation name and AWS Signature Version 4 signing. The connector keeps that protocol in `internal/connectors/native/dynamodb` while the definition bundle owns the reviewed API ledger, schemas, CLI metadata, docs, fixtures, and write-action contracts.

Officially reviewed scope: DynamoDB `DynamoDB_20120810` plus DynamoDB Streams `DynamoDBStreams_20120810`, API version 2012-08-10. DAX is a separate AWS service and is not counted.

Implemented fixture-backed surface counts:

| surface | count |
| --- | ---: |
| ETL/read + DynamoDB Streams changefeed streams | 27 |
| Bounded direct/keyed reads | 3 |
| Reverse-ETL write actions | 26 |
| Binary/import-export operations blocked on shared runtime | 2 |
| Raw PartiQL statement operations disallowed | 3 |
| Certified live operations | 0 |

## Auth setup

Required secret fields:

- `access_key_id` — AWS access key id used in the SigV4 credential scope.
- `secret_access_key` — AWS secret access key used only to derive SigV4 HMAC signing keys.

Required live config:

- `region` — AWS region in the SigV4 scope.
- `endpoint` or `base_url` — DynamoDB JSON API endpoint, for example `https://dynamodb.us-east-1.amazonaws.com`.

Common optional config:

- `streams_endpoint` — DynamoDB Streams endpoint when it differs from `endpoint`.
- `table_name` / `table` — default table for table-scoped streams and writes.
- `table_arn`, `resource_arn`, `stream_arn`, `shard_id`, and operation-specific ARN/name fields for metadata or Streams operations.
- `page_size` (default `100`) and `max_pages` (default `100`) bound read/changefeed loops.
- `mode=fixture` runs deterministic credential-free tests and never contacts AWS.

Never put secret values in prompt text or docs. Load credentials from environment variables, stdin, or the credential store.

## Streams notes

The connector declares 27 read/changefeed stream surfaces. Native code builds closed JSON-RPC bodies for each stream; there is no raw target/body escape hatch.

DynamoDB read streams:

- `items` (`Scan`) and `query_items` (`Query`) are bounded by `page_size`/`max_pages` and stop on `LastEvaluatedKey`.
- Metadata/list streams cover `Describe*`, `GetResourcePolicy`, and `List*` operations from the official Smithy model. Records expose the official response member names directly, plus `id` and `operation` metadata; they do not wrap provider responses in a fabricated `response` envelope.
- `query_items` uses typed `KeyConditions` from `query_key_name`, `query_key_type`, and `query_key_value`; it does not expose raw key-condition expressions.

DynamoDB Streams/changefeed surfaces:

- `streams_list_streams`, `streams_describe_stream`, `streams_get_shard_iterator`, and `streams_get_records` track the official Streams operations.
- `ReadCDC` uses `GetShardIterator` then bounded `GetRecords`, emits `INSERT`/`MODIFY`/`REMOVE` events, and stores each emitted record's `SequenceNumber` with `AFTER_SEQUENCE_NUMBER` resume state.

## Write actions & risks

`writes.json` declares 26 named reverse-ETL actions. Each action has a closed top-level schema, risk text, dry-run preview support, and connector-local request construction. Reverse ETL remains plan -> preview -> explicit approval -> execute.

Destructive or high-risk actions declare `confirm: destructive`, including table/resource deletion, restore operations, transaction writes, and table/global-table updates. `delete_item` is provider-idempotent when no condition is supplied: deleting a missing item succeeds according to DynamoDB semantics.

The connector intentionally does not expose raw HTTP, raw AWS JSON bodies, raw PartiQL, unrestricted scans, arbitrary expression strings, generic shell/file operations, or generic SQL/query surfaces. `UpdateItem` uses typed `AttributeUpdates`; `Query` uses typed `KeyConditions`; PartiQL statement operations are disallowed in the ledger.

## Known limits

- Fixture-only verification is not live certification. `certified=0` until a separately approved live executor supplies redacted artifacts.
- `ExportTableToPointInTime` and `ImportTable` are tracked but blocked because the current shared connector-command runtime has no approved AWS S3 binary/import-export executor for these workflows.
- `BatchExecuteStatement`, `ExecuteStatement`, and `ExecuteTransaction` are blocked/disallowed because they are raw PartiQL statement surfaces, and this task forbids raw PartiQL/query or arbitrary expressions/bodies.
- DynamoDB item attributes are table-specific, so item streams expose generic `pk` plus additional flattened item attributes at runtime.
