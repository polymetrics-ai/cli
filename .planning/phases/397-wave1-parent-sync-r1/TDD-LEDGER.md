# Issue #397 Wave 1 TDD Ledger

Status: VERIFY complete; REVIEW pending
Manual GSD fallback: PLAN → RED → GREEN → REFACTOR → VERIFY → REVIEW → INTEGRATE (`programming-loop` is absent from the 69-command adapter registry).

## Planned RED / regression evidence

| Conflict / risk | RED or regression contract | GREEN target | Status |
|---|---|---|---|
| `internal/cli/cli.go` | Combined Gong dynamic dispatch through the Cobra root plus native/help/golden/certify/reverse contracts | Gong routes and Architecture v2 routing/config/events survive together | green: `TestWave1ParentSyncGong...`, native Cobra, certify, reverse, and golden suites |
| `internal/connectors/connectors.go` | Gong operation/direct-read/write discovery plus existing connector contracts | Main Gong capabilities remain available without dropping parent integration behavior | green: both `RuntimeConfig` policy families and full connector suites |
| `internal/connectors/connsdk/http.go` | JSON arrays, multipart, query handling, retries and parent telemetry/redaction tests | Payload behavior and instrumentation coexist | green: multipart retry/telemetry/redaction regression plus existing suites and focused race |
| `go.mod` / `go.sum` | module verify/tidy and build expose dropped requirements or inconsistent sums | union graph verifies and tidy diff is empty | green: `go mod verify`, `go mod tidy -diff`, build, and `make verify` |
| auto-merges | focused CLI/app/connsdk suites detect semantic loss outside conflict markers | all focused suites pass | green: app identities/reverse, CLI, connectors, connsdk, and full suite |
| machine/reverse safety | golden transcript and reverse native tests | stdout/stderr/JSON/exit contracts and plan → preview → approval → execute remain unchanged | green: golden/native tests and local-only `make verify` smoke |

## Evidence log

- 2026-07-23: `git merge-tree --write-tree origin/feat/cli-architecture-v2 origin/main` reported conflicts in `go.mod`, `go.sum`, `internal/cli/cli.go`, `internal/connectors/connectors.go`, and `internal/connectors/connsdk/http.go`; related `internal/app/app.go` and `internal/connectors/connsdk/http_test.go` auto-merged.
- No production edits or merge resolution occurred before this ledger and the plan/checklist were created.
- RED: after the ordinary merge stopped, the first focused test failed before compilation because unresolved `go.mod` retained conflict syntax (`unknown directive: <<<<<<<`). This established that no side-selecting build could pass. Every conflict was then reconciled manually before test execution.
- GREEN: focused Wave 1 regressions passed for Gong-via-Cobra routing, native CLI/certify/reverse coexistence, both connector runtime policies, multipart retry telemetry, and safe redaction. Merge commit: `c545c3740c71b889fd2f1f64cec5491003f7b654`.
- REFACTOR: full `go test -timeout 20m ./...` exposed stale pre-main behavior in `TestCobraRouterShellPreservesLegacyHelpInterceptionForFallback/dynamic_connector_help` (expected exit 1 but Gong correctly emits a `CommandManual` envelope at exit 0). The expectation was reconciled without changing production behavior; commit `2a2e964b17144939b0a42f297de0d2b1c87383e1`.
- VERIFY: formatting/diff/module/build/vet/full tests, focused race suites, representative CLI routes, and `make verify` pass. The make smoke used a temporary local root and retained reverse plan → preview → approval → execute order.
- REVIEW at `f3df1b169625891b60dce15e332c7b535dd6ff21`: fresh exact-head code review was clean and confirmed broader endpoint/approval/preview/retry hardening observations were pre-existing and not worsened by Wave 1; they are escalated for captain follow-up rather than silently broadening this merge-only slice. Independent evidence verification found two Wave 1 artifact defects: stale #462 review state and 12 incoming Gong planning files with terminal blank lines that failed the base-to-head whitespace gate. Both are corrected in the next review-fix checkpoint.
- REVIEW/INTEGRATE at `3fd63fbe0f526873fa3adb8a75fa5f20342d52a6`: fresh local Codex review was clean, Shepherd returned `PROCEED` with trajectory geomean 4.87, and draft PR #495 opened on the identical convention-compliant head; required branch-specific workflows passed.
- Captain-approved scope extension: focused RED proved the old PM route required the unavailable command and GitHub-bot coverage; canonical PM route GREEN is implemented at `d72a93018933541d390884f96b285856e269a1ab`. Prior exact-head review/Shepherd evidence is historical; full final evidence follows `.planning/phases/397-pm-orchestrator-extension/`.
