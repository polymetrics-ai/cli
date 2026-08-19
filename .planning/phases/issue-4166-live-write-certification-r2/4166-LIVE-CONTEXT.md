# Issue 4166 Live Write Certification — Context

## Task Delivery Header

- Issue: Refs #4166 — test(certification): prove the three unexercised-coverage gaps are closed
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: A PR from `fm/cli-truth-certification-live-coverage-r1` is open against the stated base, with the safe live-write set executed through the production certification entry point, non-live residue distinguished from passes, and all required local gates recorded.
- Working branch: `fm/cli-truth-certification-live-coverage-r1`
- Task: Replace GitHub certification’s prepared-only write coverage with definition-driven live execution for every safe, reversible action; explicitly classify and report the remaining actions; make `--external-proof --full-parity` enable and require full write parity; and prove a post-schema deliberate defect fails certification.
- Verification: Action inventory and classification review; focused red/green Go tests for the certification production path, report/batch aggregation, and CLI parsing; built-binary help and external-proof checks; live private-GitHub read-backs for the safe set; formatting, vet, lint, generator, website, boundary, contract, and workflow gates.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every safely executable GitHub write reaches the production request builder and bounded provider mutation/read-back | live | A private, run-owned resource changes through `pm connectors certify github`, then an independent GitHub read-back observes the expected state and cleanup restores the run-owned resource. |
| A deliberately broken post-schema write action fails certification | fake | A test-scoped definition seam corrupts one action after schema compilation without shipping a broken bundle; the production certification entry point returns a failed report naming that action. |
| Non-live actions cannot be confused with coverage | fake | Deterministic report and batch tests assert an explicit non-pass outcome and exclude it from pass aggregation; a live provider action is unsafe or needs unprovisioned disposable infrastructure as recorded in the action inventory. |
| `--full-parity` runs full writes or refuses the claim | fake | CLI tests drive the exact external-child argument path and assert full/write options plus rejection whenever an applicable stage is skipped. |

## Fixed constraints

- Target connector: GitHub. Shared certification changes must be driven by bundle-owned action metadata and must not hard-code `github` into shared production code.
- The base branch’s prior 607-action `DryRunWrite` proof is preparation coverage, not live-write coverage, and must not be retained as a substitute.
- No resource outside a private repository or other infrastructure created solely for this task may be mutated. The retained certification repository itself must not be deleted.
- The action inventory must precede implementation. If safely executable actions are materially fewer than 605, the status record must carry a `needs-decision` before a live harness is built.
- GSD commands are executed inline/manual because the canonical single-worker contract forbids workflow-role spawning. Loaded skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-documentation`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-lint`, `vercel-react-best-practices`, and `vercel-composition-patterns`. The otherwise routed `frontend-design` and `web-design-guidelines` skills are unavailable in this environment; no React component change is planned.

## Initial discussion decisions

1. The safe-set runner must use the same `runCertify` → `certify.Runner.Run` path as the shipped binary. A direct engine client is not acceptance evidence.
2. Action scenarios must be derived from connector definitions plus an explicit certification-owned safety manifest, never a shared-provider name switch.
3. The runner must be resumable: each action has a stable identifier, durable lifecycle ledger state, a bounded rate budget, and independently verifiable cleanup.
4. Any action lacking a reversible, bounded provider state transition and independent read-back is non-live, with its exact reason carried to the report/help.
5. A full-parity proof may be written only when all applicable full/write stages ran and the report has no non-live residue falsely represented as live coverage.
