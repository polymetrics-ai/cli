---
coverage:
  - id: D1
    description: Source-cited Stripe GET/DELETE operations stay complete when their local contracts resolve.
    verification:
      - kind: unit
        ref: TestSourceImportStripeReferenceDepthOperationLocal
        status: pass
    human_judgment: false
  - id: D2
    description: An over-depth local Stripe reference retains only the exact operation as a merge-blocked source descriptor.
    verification:
      - kind: unit
        ref: TestSourceImportStripeReferenceDepthOperationLocal
        status: pass
    human_judgment: false
  - id: D3
    description: Unsafe and resource-exhausting references remain rejected, while target memoization preserves traversal counts.
    verification:
      - kind: unit
        ref: TestSourceImportRejectsUnsafeReferences; TestSourceReferenceResolverCachesNormalizedTargetsWithoutBypassingCounts
        status: pass
    human_judgment: false
---

# Summary — Stripe provider-dialect tolerance foundation

The importer no longer walks every component before emitting source operations.
It indexes source grammar and reserves bounded counts first, then resolves each
operation locally. Normalized pointer targets are memoized per source document;
cycle detection, target-kind validation, traversal/reference counts, and byte
budgets remain on each traversal.

A typed depth exhaustion is retained only for a source-lock version that
already permits source contract gaps. It emits a source-cited skeletal
descriptor with `cli-source-descriptor-foundation-r1`, exact method/path/source
location, and a merge-blocking projection gap. Other unsafe reference errors
remain hard import failures. No Stripe bundle or command was generated.

Manual inline GSD fallback: the project contract forbids spawning workflow
roles in this runtime, so `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review` prompts were generated with `scripts/gsd` and
their required work was performed inline. See the ledger, verification, UAT,
and review artifacts in this phase directory.
