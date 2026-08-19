---
coverage:
  - id: D1
    description: Real PostgreSQL pgoutput CDC reaches a live PostgreSQL history target with source/target restart and replay.
    verification:
      - kind: integration
        ref: internal/connectors/native/postgres/cdc_managed_target_route_integration_test.go:TestPostgresCDCToManagedTargetHistoryRouteLive
        status: pass
    human_judgment: false
  - id: D2
    description: Every declared destination change_capture route cell rejects before connector I/O.
    verification:
      - kind: unit
        ref: internal/app/change_capture_dispatch_test.go:TestDeclaredChangeCaptureRoutesRefuseBeforeIO
        status: pass
    human_judgment: false
  - id: D3
    description: PostgreSQL CDC to GitHub remains non-executable under the current declarations.
    verification:
      - kind: other
        ref: internal/connectors/defs/github/sync_transport.json and VERIFICATION.md declaration inspection
        status: pass
    human_judgment: false
---

# Summary — Issue 4095 residual

The live R4 route and all four named destination-mode refusal rows are delivered. R3 remains explicitly non-executable: GitHub has no logical-CDC source binding and its destination does not support deletes. No API writer or capability declaration was added.

The requested restart counterfactuals showed the initial apparent stall was a test-route page-limit mismatch, not a PostgreSQL restart regression. The engine had already rejected the unsafe page before target mutation; the test now fails immediately when its reader returns an error.

GSD was completed as an inline/manual single-worker fallback because this environment forbids the reviewer subagent that the canonical command would spawn. See PLAN.md, TDD-LEDGER.md, VERIFICATION.md, and REVIEW.md.
