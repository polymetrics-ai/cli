# Issue #4063: Flow-authoring discovery metadata - Context

**Gathered:** 2026-08-11
**Status:** Ready for TDD execution
**Mode:** Manual issue-phase fallback after resolved GSD phase-op rejection

## Phase Boundary

Repair only the derived discovery coordinate for the existing flow_authoring
workflow kind. The runtime annotation is already at line 20 and func runFlow
is already at line 21; the checked-in generated projection is stale.

## Implementation Decisions

### Generated artifact ownership

- **D-01:** Use only go run ./cmd/connectorgen certification-matrix to write
  the matrix. JSON is never hand-edited.
- **D-02:** Treat the exact pre-change --check failure as the RED assertion.
  This correction changes generated metadata, not behavior or a production Go
  implementation, so no new production test is authored.
- **D-03:** Require the generated semantic diff to be exactly one scalar in
  internal/connectors/certifications/flow-matrix.json from :20 to :21.

### Delivery and safety

- **D-04:** Preserve existing #4060 branch, required stacked base, and draft
  state. Push only the existing branch after local gates pass.
- **D-05:** No credentials, provider calls, reverse ETL, external writes, new
  dependency, CLI help/manual/website change, or certification promotion is in
  scope.
- **D-06:** Record correction 4 / 5 as reserved before the generator runs;
  never create correction 6.

## Canonical References

- AGENTS.md — issue-first lifecycle, generated-artifact boundary, local gates,
  and stacked-PR safety.
- .agents/agentic-delivery/references/required-skills-routing.md — required
  Go and review skill routing.
- .agents/agentic-delivery/references/gsd-pi-adapter.md — required command
  resolution and manual-fallback rule.
- .agents/agentic-delivery/canonical/delivery-contract.json — single-worker
  execution and no-mistakes constraints.
- .agents/agentic-delivery/contracts/issue-agent-contract.md — GSD/TDD,
  commit, review, and PR evidence contract.
- .planning/phases/issue-3897-flow-connection-scope-r1/TDD-LEDGER.md —
  authoritative previous correction count of 3 / 5.
- .planning/phases/issue-3897-flow-connection-scope-r1/RUN-STATE.json —
  authoritative correction_rounds value of 3.
- cmd/connectorgen/certificationmatrix.go — canonical check/generate command.
- cmd/connectorgen/certificationflow.go — runtime-source discovery and flow
  matrix rendering.
- internal/cli/flow_cli.go — authoritative annotation and func runFlow line.
- internal/connectors/certifications/flow-matrix.json — generated projection
  whose one scalar may change.

## Existing Code Insights

- The main connectorgen dispatcher routes certification-matrix to
  runCertificationMatrix.
- In check mode, the generator builds all three certification artifacts and
  byte-compares each to its generated payload; it fails on the stale flow
  matrix before any write.
- Runtime-source discovery uses the pmcert:workflow annotation and records
  the annotated function coordinate in the flow matrix.

## Deferred Ideas

None. Any change beyond the source-coordinate projection belongs to a separate
issue and cannot consume this correction reservation.
