# Plan — CircleCI semantic lane repair R2

## TDD slices

1. **Red — semantic classification constraints.** Add named focused tests that reject the frozen matrix when it classifies a source-backed non-GET bounded read as a mutation, omits ETL for each of the five referenced-response paginators, treats an inline/request-only paging shape as ETL, infers sync from paging, or omits the named non-executable gap for an actual source-backed outbound-webhook registration. Add exact source-ID/backlink/count reconciliation assertions.
2. **Green — source-backed corrections.** Change only the affected CircleCI matrix cells and supporting cited facts. Resolve retained response `$ref` facts in the test reader, not by adding or editing source locks. Preserve full source-row and seven-lane-cell denominator. Classify `createWebhook` and `updateWebhook` as source-backed `missing_foundation` only when their provider summary/description and typed request body prove outbound event delivery.
3. **Refactor — deterministic evidence.** Keep test helpers minimal and data-driven; ensure no test recognizes eligibility from HTTP method, inline-only schema, an array, `limit`, or request-only paging alone.
4. **Verification/review.** Run focused and complete package tests, race test, vet, strict JSON and exact count checks, contract gate, diff gate, then a manual source-semantic/security scope review. Push only after the green local evidence exists and post a #4382 proof/re-review request.

## Acceptance evidence

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| All retained source rows and seven cells remain accounted for | live | The test reads the unchanged lock and matrix and fails when any source ID/cell count/backlink is absent or invented. |
| A non-GET bounded read stays `direct_read` and does not become a mutation | live | A source-backed CircleCI GET-exception test asserts the actual operation's direct-read state and rejects mutated expected lane facts. |
| Every referenced-response paginator has independent ETL mapping | live | The test resolves the five named response `$ref` schemas and fails unless each row has a cited ETL `mapped_unproven` decision. |
| Request-only or inline false positives do not get ETL | live | In-memory mutations show an array, limit, or request-only paging input alone cannot satisfy the ETL predicate. |
| Sync remains event/webhook-only | live | The test rejects sync promotion for every pagination-only operation; only source-proven outbound registration may retain a named `missing_foundation` sync cell. |
| Every documented provider mutation remains dual-lane | live | Exact source mutation IDs must retain both direct-write and reverse-ETL decisions, while semantic reads do not enter that cohort. |
| No execution or foundation behavior changes | live | Changed paths are limited to CircleCI matrix/test/evidence; the matrix rejects `implemented`, and permits `missing_foundation` only for source-proven outbound webhook registration. |

## Non-goals

No source-lock or crosswalk changes, no connector-definition projection, no shared generator/parser/runtime/certification/transport/Atlas changes, no new dependencies, no PR creation, and no merge.
