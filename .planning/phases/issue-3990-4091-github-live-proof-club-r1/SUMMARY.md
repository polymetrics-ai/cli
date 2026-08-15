---
coverage:
  - id: D1
    description: "The production GitHub certification command completes a whole-surface real-provider run under shared admission with observable rate and provider outcomes."
    requirement: ISSUE-3990-LIVE
    verification:
      - kind: e2e
        ref: "LIVE-EVIDENCE.md #3990: built-binary external proof, 1,370 stages and 83 provider exchanges"
        status: pass
      - kind: integration
        ref: "go test -timeout 20m ./internal/connectors/certify -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "The production ETL transport command performs approved set-replace and keyed GitHub writes, then reuses durable exact-scope authorization without a fresh token."
    requirement: ISSUE-4091-LIVE
    verification:
      - kind: e2e
        ref: "LIVE-EVIDENCE.md #4091: first and tokenless real-GitHub runs with independent label read-back"
        status: pass
      - kind: integration
        ref: "TestETLRunTransportApprovalAllowsDurablePlanReferenceWithoutTokenCarrier"
        status: pass
    human_judgment: false
  - id: D3
    description: "Approval replay and real authentication refusal fail with typed errors and leave provider labels and durable checkpoints unchanged."
    requirement: ISSUE-4091-LIVE
    verification:
      - kind: e2e
        ref: "LIVE-EVIDENCE.md #4091: replay and GitHub 401 negative-side-effect ledger"
        status: pass
      - kind: integration
        ref: "TestIssueLabelDestinationRejectsUnapprovedOrMismatchedOrExpiredOrReplayedPlan"
        status: pass
    human_judgment: false
  - id: D4
    description: "Every required edge is live, deterministically simulated, or explicitly identified as unsafe or unavailable to test live, and all run-owned provider state is restored."
    verification:
      - kind: manual_procedural
        ref: "LIVE-EVIDENCE.md captain edge-case matrix and final cleanup ledger"
        status: pass
    human_judgment: false
key_files:
  created:
    - .planning/phases/issue-3990-4091-github-live-proof-club-r1/LIVE-EVIDENCE.md
  modified:
    - internal/connectors/certify/external_proof.go
    - internal/connectors/certify/stages_glue.go
    - internal/connectors/certify/stages_source.go
    - internal/connectors/certify/stages_write.go
    - internal/cli/etl_transport.go
---

# Summary — Issues 3990 and 4091: live GitHub certification proofs

Both audited live-proof gaps now have real GitHub evidence through a fresh production binary.

- The whole-surface certification route passed 1,370 stages with no failures and recorded 83
  fingerprinted real-provider exchanges under shared rate admission. Live REDs led to focused fixes
  for unavailable/empty captures, mandatory destructive cleanup preview and confirmation, and the
  external proof's per-target retry bound.
- The closed issue-label transport performed separately approved `full_overwrite` and
  `incremental_upsert` writes and then completed identical-scope tokenless reruns. A live RED showed
  that the CLI parser blocked the already-supported durable app authorization; the parser now keeps
  the plan and destructive confirmation mandatory while making a fresh token optional only for a
  later forward run.
- Replayed tokens produced typed `AuthorizationTokenReplayError`; a real invalid GitHub credential
  produced typed `connsdk.HTTPError` status 401. Both paths left provider labels and checkpoints
  unchanged.
- GraphQL query and separately approved star/unstar mutations were proven live, and all run-owned
  labels, the star, and the remote fixture branch were removed and independently read back absent.

The complete sanitized observations, call chains, edge classification, artifact hashes, and cleanup
ledger are in `LIVE-EVIDENCE.md`. Exact credential values, approval tokens, request targets, and
rendered rate scope are intentionally absent.
