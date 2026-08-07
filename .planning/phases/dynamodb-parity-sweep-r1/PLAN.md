# DynamoDB documented-operation parity — plan

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Goal

Bring `internal/connectors/defs/dynamodb` from **9 declared endpoints** (1 covered, 8 legacy
`excluded`, no `operation_ledger_version`) to the **full 58-operation documented surface**, with every
operation partitioned exactly once and individually reachable as `pm dynamodb <command>`.

## Operation surface, derived before authoring

Artifact: AWS service model
`https://raw.githubusercontent.com/boto/botocore/develop/botocore/data/dynamodb/2012-08-10/service-2.json`
(`serviceId: DynamoDB`, `targetPrefix: DynamoDB_20120810`, `apiVersion: 2012-08-10`), byte-identical
at stable tag `1.43.66`, cross-checked set-equal against `API_Operations.html`.

**58 operations. 27 read, 31 write. No deprecated operations.**

The ledger's carried-forward 57 was correct when taken; the delta is exactly `SearchVectors`, absent
in botocore 1.35.0/1.38.0/1.40.0 and present in current. Writes reconcile at 31 with no adjustment.

### Shape constraint that drives everything below

DynamoDB is **JSON-RPC, not REST**: every operation is `POST /` and is selected by the
`X-Amz-Target: DynamoDB_20120810.<Operation>` header. `requestUri` is `/` for all 58, so path alone
cannot distinguish them. Auth is AWS SigV4, which is a Tier-2 `AuthHook` concern, and this bundle is
already hook-backed.

Consequence for `api_surface.json`: rows keyed on `POST /` would collapse to one endpoint. Rows must
carry the target, following the qualified-path convention this repo already uses (notion's
`/v1/search (object=page)`), i.e. `POST /` + target suffix per operation. **This is the first
decision to validate against `connectorgen validate` before bulk authoring** — if the validator or
runtime preflight rejects it, the partition strategy changes and the plan is revised before any
sub-agent authors 58 rows.

### Planned partition

| Bucket | Count | Notes |
| --- | --- | --- |
| ETL streams | ~4–6 | `ListTables`, `ListBackups`, `ListExports`, `ListImports`, `ListContributorInsights`, `Scan` — collection reads |
| Direct reads | ~21 | `GetItem`, `BatchGetItem`, `Query`, `Describe*`, `SearchVectors` |
| Reverse-ETL / direct writes | ~31 | `Put/Update/Delete/Create/Restore/Import/Transact*Write/BatchWriteItem/Tag/Untag`, plus the 3 PartiQL ops gated as mutations |
| Binary | 0 | DynamoDB has no binary upload/download endpoint; exports land in S3, not in the API response. Recorded as genuinely absent, not blocked. |

Counts are planned, not asserted. The red test asserts the **derived** totals; bucket sizes settle
during authoring and the test is updated only alongside evidence.

### Deliberately out of scope

- **DynamoDB Streams** (4 ops) and **DAX** (21 ops) are separate AWS services with their own
  `targetPrefix` and doc slugs. Not merged. A naive scrape of every `API_*.html` link yields 84 —
  that miscount is exactly what this plan refuses.
- Webhook events: none exist here; there are no subscription-management operations either.

## TDD sequence

1. **RED** — add `cmd/connectorgen/dynamodb_api_surface_test.go` asserting 58 rows, the 27/31 split,
   `operation_ledger_version`, no duplicate endpoint key, and zero blank dispositions. It must fail
   against today's 9-row bundle. Capture the failure text in `TDD-LEDGER.md`.
2. **GREEN** — author the bundle to satisfy it (sub-agents do the bulk; this lane judges every diff).
3. **REFACTOR** — docs, fixtures, catalogs, ledger resync.
4. Gates, then no-mistakes.

## Safety notes

- Do not loosen `connectorgen validate`, the connector boundary gate, `certify`, or
  `TestEveryImplementedCommandPassesRuntimePreflight` to make this pass.
- Nothing is marked `implemented` unless its command runs; blocked rows carry a **named** dependency.
- No credential or token-derived value is ever emitted.
- Keep the diff scoped to dynamodb; revert unrelated generator churn.
