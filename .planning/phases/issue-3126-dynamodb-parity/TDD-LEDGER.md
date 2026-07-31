# TDD Ledger — issue 3126 DynamoDB parity

## Red/green/refactor log

| Slice | Red test / evidence | Green implementation | Refactor/safety notes |
| --- | --- | --- | --- |
| Official operation inventory | `TestParityOperationRegistryCounts` asserts 61 official operations and disposition counts: 27 stream, 26 write, 3 direct, 2 blocked binary/import-export, 3 disallowed PartiQL. | `api_surface.json`, `streams.json`, `writes.json`, `operations.json`, `cli_surface.json`, and native registries reconcile to the same counts. | Smithy/docs-derived names only; no provider calls. |
| Native read streams | `TestReadFixtureCoversEveryStream` covers all 27 streams; `TestReadListStreamsUsesDocumentedPagination` verifies DynamoDB Streams `Limit`/`ExclusiveStartStreamArn` pagination. | `reader.go` implements closed X-Amz-Target dispatch, page_size/max_pages bounds, operation token fields, and fixture mode. | Bounded by page_size/max_pages and ctx cancellation. |
| Direct reads | `TestOperationDirectReadBuildsClosedBodies`, `TestOperationDirectReadRejectsMissingRequiredFields`, and `TestOperationDirectReadEnforcesMaxBytes` assert `get_item`, `batch_get_item`, and `transact_get_items` build closed bodies, validate required fields, default scalar key type, and enforce max_bytes. | `direct.go` builds typed DynamoDB JSON-RPC bodies and uses `doJSONLimited` for max byte caps. | No raw PartiQL, arbitrary expression, endpoint, or request body passthrough. |
| Typed writes | `TestWriteActionsValidatePreviewAndExecuteAgainstReplay` exercises all 26 write actions through schema validation, dry-run, and `httptest.Server`; it asserts batch/transaction writes are typed builders, not raw `request_items`/`transact_items` passthrough. | `writes.json`, `writer.go`, and `operations.go` implement typed writes, dry-run redaction, SigV4 JSON-RPC execution, and typed BatchWriteItem/TransactWriteItems builders. | Destructive/admin actions carry confirmation where applicable; fixtures use synthetic AttributeValue maps only. |
| CDC/changefeed | `TestReadCDCUsesShardIteratorAndEmitsEvents` asserts shard iterator + GetRecords calls and INSERT event flattening. | `cdc.go` implements DynamoDB Streams CDC using configured stream ARN/shard iterator state and bounded max records. | No live AWS streams; all fixtures/synthetic HTTP. |
| Docs/fixtures/guard | Focused connectorgen temp-root validation, all-defs validation, conformance, generated CLI/docs/website data, and boundary checks exercised. | DynamoDB docs/manual/skill/catalog/website generated surfaces updated; unrelated connector docs were reverted. | Fixture-only remains uncertified; no secret-shaped fixture values added. |

## Manual GSD fallback note

`gsd-programming-loop` is required by AGENTS.md, but the repo-local `scripts/gsd` registry in this worktree reports `unknown GSD command: programming-loop`. The worker therefore generated the `plan-phase` prompt and maintains this explicit TDD ledger before production edits, preserving the programming-loop lifecycle manually.
