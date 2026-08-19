# Plan — issue #4273 connector surface sweep batch 1

## Task Delivery Header

- Issue: Refs #4273 — chore(connectors): surface-sweep batch 1 declarative coverage
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main`, containing one committed 20-connector materialization batch, durable sweep/foundation ledgers, batch evidence, and completed local checks.
- Working branch: fm/cli-surface-sweep-batches-r1
- Task: Create the resumable fleet ledger and foundation-gap log, then materialize and gate one evidence-qualified 20-connector declarative surface batch without foundation code or Zoom changes. Completion means a truthful per-connector declared/proven status, a resume pointer, and a reviewable PR.
- Verification: `batch plan`, `batch materialize`, `batch gate`, `connectorgen validate`, `surface-sync --check`, `surface-reconcile --check`, certification-candidate checks, runtime preflight, connector-boundary, targeted generator tests, Go vet/build, individual `make verify` gates, and `make verify` before push.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Durable fleet progress exists before materialization | live | JSON has exactly 552 unique connector rows, all five class keys, a transport flag, timestamps, state/reason fields, totals, and a resume pointer. |
| Foundation limitations are visible rather than masked | live | Structured gap log names the missing executor/policy, affected connectors, evidence, and smallest bounded fix. |
| Candidate selection is evidence-qualified and bounded | live | `batch plan --size 20` emits a manifest with exactly 20 source-ledger candidates and preserves their URLs, versions, dates, and measured totals. |
| Each accepted generated bundle is structurally/runtimely checked | live | Materialize/gate reports name every included or dropped candidate; each included command runs real `commandrunner.Preflight`. |
| The branch has no connector-boundary or generated-surface drift | live | `connectorgen validate`, `surface-sync --check`, `surface-reconcile --check`, and `connector-boundary` exit successfully. |
| Certification is not overstated | live | Certification-candidate output is recorded per included connector and ledger status stays declared/validated/gated, never certified, without live evidence. |

## Required skills and GSD mode

Loaded: `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-documentation`.

GSD commands resolved: `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`. The canonical contract forbids spawning those role agents here, so the generated prompts are executed as an inline/manual fallback. `discuss-phase 4273 --auto` and `plan-phase 4273 --tdd --skip-research` are recorded in the phase trace; research is bounded to the supplied corpus, batch implementation guide, and current provider-artifact ledger.

## TDD slices

1. **RED — fleet visibility:** demonstrate the baseline gap: only 2/552 transport files and no fleet resume ledger. Generate the ledger and require its exact-record/count invariant before any materialization.
2. **GREEN — deterministic evidence batch:** create a 20-record manifest from the external evidence ledger, materialize from the immutable source root, and retain all included/drop records.
3. **GREEN — runtime truth:** gate every candidate, run branch-wide structural checks, and update each progress row to `materialized`, `validated`, `gated`, `failed`, or `skipped` with the exact report reason.
4. **REFACTOR — truthful consolidation:** run surface reconciliation without promotion, write certification-candidate results, update summary totals/resume pointer, and commit the batch once reviewable.

## Ownership and CLI/docs parity

This batch is a dedicated, pipeline-owned multi-connector declaration sweep; it changes no shared runtime/tooling and no Zoom bundle. Its generated connector command metadata is checked by the normal help/manual/website generators and docs checks. No hand-authored public command, flag, output, help topic, manual prose, or website page is introduced; runtime help and generated-doc commands are still verified and the PR records that limited scope.

## Commit checkpoints

1. Plan + initial ledgers committed before materialization.
2. Batch manifest/materialization/gate evidence + accepted connector definition changes committed after clean checks.
3. Review fixes, if any, committed separately and rechecked.
