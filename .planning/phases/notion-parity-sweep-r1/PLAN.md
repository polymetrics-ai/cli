# Notion documented-operation parity — plan

Part of `cli-top50-fixed-schema-sweep-r1`. Issues: parent #3062, with #3063–#3069.

> **Recorded retrospectively.** This phase was written *after* the implementation, not before it.
> The lane was running the no-mistakes half of the contract without a GSD phase; the captain
> corrected that on 2026-08-07 and directed that notion not be rebuilt. These artifacts therefore
> document what was actually done and verified, and deliberately do not claim a red-first sequence
> that did not happen. **dynamodb onward is properly plan → red → green**; see
> `.planning/phases/dynamodb-parity-sweep-r1`, whose failing test was committed before any
> production code.

## Goal

Bring `internal/connectors/defs/notion` from **6 declared endpoints** (3 read streams, 3 excluded,
`capabilities.write: false`, no `cli_surface.json`) to the **full 51-operation documented surface**,
every operation partitioned exactly once and individually reachable as `pm notion <command>`.

## Operation surface, derived from the provider artifact

Artifact: Notion's official **OpenAPI 3.1.0** document, <https://developers.notion.com/openapi.json>
(876 KB, retrieved 2026-08-07). The sweep ledger recorded `html_reference` and a carried-forward
count of **50**; both were stale, and the `html_reference` misclassification is also what kept notion
out of `connectorgen batch materialize`.

**51 documented operations** = 49 OAS operations (20 GET, 17 POST, 8 PATCH, 4 DELETE over 34 paths)
+ 2 legacy endpoints documented only under the nav's explicit "Databases (deprecated)" group.

The 31 top-level `webhooks` entries are webhook **events**, excluded per the counting policy; notion
publishes no webhook management endpoints. Counting documentation *pages* instead of unique
method+path actions yields ~55–57 because four operations are each documented twice — the count
policy counts actions, giving 51.

## Delivered partition

| Bucket | Count |
| --- | --- |
| ETL streams | 6 |
| Direct reads | 18 |
| Write actions | 24 (21 implemented, 3 partial) |
| Blocked with a named dependency | 1 (shared file upload runner) |
| Not executable, source-cited | 5 (3 OAuth `disallowed`, 2 provider-`deprecated`) |

54 rows carry 51 operations: three operations sit on two rows each.

## Safety notes

- Nothing marked `implemented` unless its command runs; blocked rows name their dependency.
- No validation, boundary, certify, or runtime-preflight gate was loosened.
- No credential or token-derived value is emitted; the three OAuth endpoints are refused outright.
- Diff kept scoped to notion; pre-existing amazon-sqs generator drift was reverted.
