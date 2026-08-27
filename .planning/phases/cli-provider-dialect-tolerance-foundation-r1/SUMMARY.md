---
coverage:
  - id: D1
    description: The exact retained Stripe artifact accounts for all 589 locked source operations with unique source citations.
    verification:
      - kind: unit
        ref: TestSourceImportRetainedStripeCorpus
        status: pass
    human_judgment: false
  - id: D2
    description: A bounded response-reference chain becomes an operation-local source-descriptor condition without fabricating a request or response contract.
    verification:
      - kind: unit
        ref: TestSourceImportRetainedStripeCorpus; TestSourceImportStripeReferenceDepthOperationLocal
        status: pass
    human_judgment: false
  - id: D3
    description: Unused over-depth components are accounted for, while hidden malformed, dynamic, external, cyclic, and resource-exhausting input remains terminal.
    verification:
      - kind: unit
        ref: TestSourceImportRetainsUnusedDepthDisposition; TestSourceImportDoesNotTreatDepthAsDocumentSafetyExemption; TestSourceImportGapLockRejectsUnusedSchemaResourceExhaustion
        status: pass
    human_judgment: false
  - id: D4
    description: A retained source descriptor condition reaches registry-backed commandrunner preflight before credentials, executor dispatch, or provider I/O.
    verification:
      - kind: unit
        ref: TestSourceProjectionRetainedStripeDepthGapStopsAtRegistryPreflight
        status: pass
    human_judgment: false
---

# Summary — Stripe provider-dialect tolerance foundation

The importer retains current-main's full source-grammar preflight and source
document-local normalized-reference memoization. A byte-backed v1/v2 lock can
only declare a lower `reference_depth_limit`, never enlarge the caller’s
resource or traversal budget. Typed reference-depth outcomes are retained as
source-cited, merge-blocking `cli-source-descriptor-foundation-r1` conditions;
all other unsafe or resource outcomes remain terminal.

The restored immutable Stripe source is imported locally from its retained
7,967,776-byte, SHA-256-pinned artifact. Its 589 source operations remain
visible through the source descriptor, `api_surface`, crosswalk/disposition,
and generated operation evidence. No Stripe route, CLI command, credential
path, or provider-I/O surface is materialized.

Manual inline GSD fallback: the project contract forbids spawning workflow
roles in this runtime, so `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review` prompts were generated with `scripts/gsd` and
their required work was performed inline. See the ledger, verification, UAT,
and review artifacts in this phase directory.
