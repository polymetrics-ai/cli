# Foundation Post-Fix Execution Plan r1

## Task Delivery Header

- Issue: Refs #4302 — status/export foundation; Refs #4303 — declarative reverse ETL; Refs #4305 — structured REST body; Refs #4306 — source import; Refs #4307 — closed operation runtime.
- Base branch: `fm/cli-current-foundations-main-integration-r1` at `c9824b5837f487acaa2c2a39126d29cf401d7fb5`.
- Merges into: `fm/cli-current-foundations-postfix-fix-wave-r1` → `fm/cli-current-foundations-main-integration-r1` → `main`.
- Delivery: every frozen postfix finding is closed by red-green-refactor evidence, atomic commits are non-force-pushed and remote-verified, an independent review of the exact final SHA has zero blockers, then the integration branch fast-forwards to that SHA; no PR or merge to `main` occurs in this lane.
- Working branch: `fm/cli-current-foundations-postfix-fix-wave-r1`.
- Task: adopt the preserved Group 1 recovery diff byte-for-byte, reconcile the reviewed reverse-ETL and structured-body heads, then repair the frozen 38 blockers and 8 warnings in the dependency order in `POSTFIX-REVIEW.md` without adding generic authority or connector-specific runtime branches.
- Verification: red-first production-shaped tests per finding; focused package/race/generator gates per atomic group; final built-binary, App, CLI, transport, source, generator, certification, and exact-SHA independent-review gates described in `POSTFIX-REVIEW.md`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Group 1 source-derived declarations remain complete and declaration-owned | live | Source/projector/generator tests reject gaps, deleted refs, stale flags, and unreachable nested POST schemas before provider I/O. |
| Provider output remains complete while the public projection masks only classified secrets | live | Receipt tests assert exact ordinary headers, IDs, bytes, and occurrence IDs, while only declared exact secret values are masked. |
| Delivery state remains durable, fenced, and retry-safe | live | Concurrent/crash/restart tests assert a loser makes no side effect, stale authorization stops before I/O, and committed delivery is not relabelled failed. |
| Final branch is independently reviewable | live | The exact local and remote SHA are identical and the post-fix deep review records zero blockers/warnings or reasoned non-actionable dispositions. |

## Frozen review workflow

1. The independent deep input review is the immutable `.planning/phases/cli-current-foundations-main-integration-r1/POSTFIX-REVIEW.md` at source `8a8a866ff6d5282c28bda12acceed8a624218f01`.
2. This plan is the canonical convergence/fix-wave ledger. It executes inline because the task assigns one Herdr worker as the project-code owner; no generated GSD role owns files or commits.
3. Each group starts by recording the relevant failing tests in `POSTFIX-TDD-LEDGER.md`, then turns those tests green and runs its focused regressions before its atomic commit.
4. After all groups, an independent exact-SHA deep review is required. A nonzero actionable finding returns the work to its dependency group; it cannot be waived by a prior report.

## Atomic groups and required scope

| Group | Frozen findings | Completion boundary |
| --- | --- | --- |
| 1 | B01-B03, B09, B12, W01 | Authenticated source/projection graph, complete REST/POST schemas, and exact provider-owned flag synchronization. |
| 2 | B04-B08, W02 | GraphQL pagination, Int32, selection/secret classification, and lossless generated documentation. |
| 3 | B13-B14, B17, B19, B24 | Immutable complete receipt and secret-only public projection through REST, GraphQL, status, SQS, App, and CLI. |
| 4 | B15-B16, B18, B21, B23, B25, W03-W04 | Sealed request execution, response-retaining transitions, precise CLI values, continuation admission, and authority limits. |
| 5 | B22, W05 | Race-safe no-overwrite binary publication and reliable multipart boundary coverage. |
| 6 | B20, B26, B33, B36, W06-W07 | Durable commit outcomes, ownership/auth fences, defensive clones, and terminal reauthorization. |
| 7 | B27-B32 | Exhaustion truth, independent delivery proof, private receipt readback, numeric fidelity, and independent deadlines. |
| 8 | B34-B38, W08 | Persisted terminal envelopes, exact CDC recovery, delivered reconciliation, and truthful preflight errors. |
| 9 | B10-B11 plus B01 closure | Exact-SHA evidence/certification subject binding and one complete regenerated artifact closure. |

## Required skills and GSD record

- Loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-lint`, and `golang-documentation`.
- Resolved commands: `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` using `scripts/gsd`; `go run ./cmd/agentcontractgen check` passed.
- Inline/manual fallback: the Firstmate brief gives this single Herdr worker project-code ownership. The Codex runtime cannot bind a GSD specialist to a safe independent worktree, so planning, execution, and review disposition occur inline; the final independent review is separately scoped to the exact final SHA.
- CLI parity: generated CLI, help/manual, docs, website data, skills, JSON output, stdout/stderr, no-action namespace behavior, and write-safety wording are generator-owned surfaces and are checked with their source changes.
