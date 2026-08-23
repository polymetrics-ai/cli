# Refs #4303 — Action-binding canonical repair R2 plan

## Task Delivery Header

- Issue: Refs #4303 — connector-neutral typed reverse-ETL destinations.
- Base branch: `origin/fm/cli-reverse-etl-action-binding-foundation-r1` at `846ff9e5d9f56cef3f9835d35b57b1b4d468b379`.
- Merges into: Firstmate-directed integration/rebase after this source-isolated repair and the parallel Foundation repair are independently green and reviewed; never directly into `main`.
- Delivery: committed and non-force-pushed repair branch with all canonical review IDs resolved, full evidence, clean worktree, and no PR opened.
- Working branch: `fm/cli-reverse-etl-action-binding-foundation-r1`.
- Task: repair AB-B01 through AB-B09 and AB-W01 through AB-W05 without weakening declarations, approval visibility, destructive-operation handling, or provider output.
- Verification: report-focused red/green tests; changed-package and race cohorts; `connectorgen validate`; `surface-sync --check`; docs/surface/contract/boundary checks; GSD verification/review artifacts; complete no-mistakes pipeline; remote SHA check.

## Evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Complete physical actions are previewed, digest-bound, and approved | fake | Synthetic typed-destination declaration records a normal action plus tombstone delete; preview/JSON/digest expose both exact actions before approval and apply cannot add another action. Real provider writes are prohibited. |
| Read-back and batches are action-compatible before I/O | fake | Recording provider proves incompatible selected action/fields/maxima cause zero calls, while distinct valid actions each write, read back, and checkpoint. |
| Missing-ok tombstones complete safely | fake | A declared 404 records unchanged, independently reads absence, checkpoints exactly once, and records no replayed DELETE. |
| Receipt/acknowledgement capacities are sealed before writes | fake | Oversize escaped/composite cases refuse or split before a recording provider sees a write; legal units produce attachable acknowledgement receipts. |
| Post-success failures preserve sanitized evidence | fake | A provider success followed by local receipt/composition/ack failure returns its ordered masked result in the persisted run without a checkpoint. |
| Workset idempotency is stable and unique | fake | Same durable workset retry sends the same key while equivalent records from a second workset send a different key. |
| Declarations and operator surfaces are closed and accurate | live | Schema/loader/doc/generator checks reject the specified invalid forms and read the exact batch/tombstone text from help/manual/website surfaces. |

## Required skills and lifecycle

Loaded: `golang-how-to`, `golang-cli`, `golang-documentation`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, `vercel-react-best-practices`, `vercel-composition-patterns`, `no-mistakes`, `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, and `gsd-code-review`.

`scripts/gsd doctor`, all five `scripts/gsd sources` lookups, all five generated lifecycle prompts, and `go run ./cmd/agentcontractgen check` passed. `gsd-sdk query init.phase-op 4303` has no registered roadmap phase; therefore this plan is the required inline/manual GSD fallback. It preserves `discuss-phase --auto` → `plan-phase --tdd` → inline execution → verification → deep code review, and records the worker-isolation limitation without weakening TDD or gates.

## TDD dependency groups

### Group 1 — declaration/action closure (AB-B01, AB-B02, AB-B04, AB-W01–AB-W04)

1. Add production-shaped failing tests for full plan/preview action disclosure; action-owned read-back and compatibility refusal; action/read-back bounded units; optional batch JSON presence; schema mapping exclusivity; strategy/eligible and binding/write-set closure.
2. Run each red test and record failure before production changes.
3. Implement the smallest closed model and pre-I/O validation, then run focused green package tests.
4. Commit `fix(synctransport): close action declaration and approval contracts` and non-force push immediately.

### Group 2 — sealed tombstone/receipt/output execution (AB-B03, AB-B05–AB-B08)

1. Add failing tests for missing-ok absence proof/no replay, escaped receipt sizing, paired composite budgeting, acknowledgement ceiling, and post-success output preservation.
2. Record red executions; implement exact receipt-bounded action units and typed result-carrying error propagation.
3. Run focused App/engine/connectors/synctransport tests and relevant `-race` cohorts.
4. Commit `fix(synctransport): seal tombstone receipts and preserved results` and non-force push immediately.

### Group 3 — durable identity and operator parity (AB-B09, AB-W05)

1. Add failing stable-vs-distinct-workset key test and docs/help/website parity golden assertions.
2. Bind provider idempotency to durable workset occurrence; update every required text surface and generated artifact only when its generator requires it.
3. Run focused green tests, docs/generator/surface checks, commit `fix(synctransport): bind workset identity and document sealed actions`, and non-force push immediately.

## CLI parity checklist

- [ ] `pm help etl` states effective batch clamping and separately sealed tombstone delete/read-back semantics.
- [ ] `pm etl` bare namespace still renders contextual help and exits successfully.
- [ ] relevant `pm etl … --help` remains accurate.
- [ ] `docs/cli/etl.md`, `website/content/docs/etl.mdx`, embedded help, and generated/golden surfaces match.
- [ ] no credentials or reverse-ETL execution are used for help validation.

## Final gates

Run the report’s focused red/green proofs, package tests and `-race`, `gofmt`, `go vet`, `go build ./cmd/pm`, `tidy-check`, `lint`, `docs-check`, `smoke-no-build`, `agent-contract-check`, `connectorgen validate`, `connectorgen surface-sync --check`, `connector-boundary`, and `release-workflow-check`. Execute the full no-mistakes pipeline only after committed local gates are green. Do not open a PR.
