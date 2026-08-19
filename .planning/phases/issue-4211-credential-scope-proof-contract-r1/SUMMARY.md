---
coverage:
  - id: D1
    description: Full-parity evidence is derived only from a passed full-parity report stage.
    verification:
      - kind: unit
        ref: cmd/connectorgen TestCertificationFullParityScopeRequiresPassedReportStage
        status: pass
    human_judgment: false
  - id: D2
    description: Bounded observed-operations evidence is publishable from protocol exchanges.
    verification:
      - kind: unit
        ref: cmd/connectorgen TestCertificationBoundedScopePublishesObservedOperations
        status: pass
    human_judgment: false
  - id: D3
    description: PostgreSQL evidence records were freshly re-issued with an honest bounded scope.
    verification:
      - kind: integration
        ref: PostgreSQL transport and CDC live evidence tests
        status: pass
    human_judgment: false
---

# Summary — Issue 4211 provable credential-scope contract

The accepted-evidence schema is now v2. Every record identifies both its claimed
credential scope and the run fact that established it. A full-parity claim can be
constructed only from `Report.FullParityVerified()`; the normal proof writer
publishes the narrower, truthful `observed_operations` scope from its sanitized
protocol exchanges.

The original fourteen PostgreSQL records were replaced after fresh live runs.
They were previously unverified full-parity assertions because those runs omitted
the full-parity stage and their importer hard-coded the assertion. The re-issued
records and their matrix pointers all state `observed_operations` with
`protocol_exchanges` proof.

See `TDD-LEDGER.md`, `VERIFICATION.md`, and `REVIEW.md` for red/green evidence,
local validation, and the inline manual review fallback.
