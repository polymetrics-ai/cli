# Issue #4072 Run State

**Phase:** issue-4072-github-app-auth-admission-r1

**Issue:** #4072 (direct child of #3754)

**Branch:** `fix/4072-github-app-auth-rate-admission`

**Recovered base:** `da8a8ff07aaf00e5c7965cd4d1d3c7252017d785`

**Correction ledger:** 1/5 active — configured-linter RED recorded; no
no-mistakes correction run started

**Canonical private finish-plan snapshot SHA256:**
`939f14f61defd993f8ad0335a5aeb617d97083c9f73a6a75259d0e312ae8f408`

## Lifecycle State

| Stage | Status | Evidence |
|---|---|---|
| Isolation gate | complete | allocated worktree and repository root match |
| Issue-first gate | complete | #4072 created and verified direct child of #3754 |
| Recovery-base gate | complete | branch starts at exact preserved `da8a8ff…` |
| discuss/context | complete | `CONTEXT.md`, `DISCUSSION-LOG.md` |
| plan-phase --tdd | complete (manual inline fallback) | `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md` |
| execute RED | observed | causal no/lost tests each fail with one premature token transport send; granting case observes zero rate decisions |
| execute GREEN | complete | focused secret-blind transport/coordinator matrix passes; GREEN commit `3f83bf3af` |
| verify-work | complete (inline automatic fallback) | `UAT.md`, bounded functional/race/vet evidence |
| code-review | complete (inline manual deep review) | `REVIEW.md` |
| gap plan / generated artifact synchronization | complete (inline manual fallback) | canonical generator/check produced only six source-location updates, preserved semantics, and all released bounded local acceptance gates pass at `f52745f…` |
| no-mistakes | ready, not started | `NO-MISTAKES-HANDOFF.md`; safe local-only vector has no `--yes` and must be driven synchronously |

## Manual GSD Fallback

`gsd-sdk query init.phase-op issue-4072-github-app-auth-admission-r1` reports
`phase_found: false` because project roadmap phases are numeric. The canonical
delivery contract also disallows spawning the GSD roles for this lane. The
required lifecycle therefore runs inline with equivalent committed context,
plan, TDD ledger, verification, UAT, summary, and review artifacts.

## Guardrails

- Do not use the preserved exhausted no-mistakes run or alter its worktree.
- The reconciliation report's bounded local acceptance matrix is complete;
  run only the fresh #4072 no-mistakes vector, never the parked #3754 run.
- Do not push, create a PR, merge, or CI-drive until the exact #3754 parent
  publication decision is authoritative.
- Do not select a parent PR route; record `needs-decision` at delivery only if
  Firstmate has not supplied an authoritative safe target.

## Focused GREEN Evidence

- `go test ./internal/connectors/hooks/github -run '^TestGitHubAppAuthRateAdmission' -count=1` — pass.
- `go test ./internal/connectors/engine ./internal/connectors/hooks/github -run 'TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubDeclaredRateLimits' -count=1` — pass.
- `go test ./internal/connectors/hooks/github -count=1` — pass.
- `GOMAXPROCS=2 go test -p 1 ./internal/connectors/engine ./internal/connectors/hooks/github -run '^(TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubWriteHook|TestGitHubDeclaredRateLimits|TestGitHubRateLimitAdmission)' -count=1` — pass.
- `GOMAXPROCS=2 go test -p 1 -race ./internal/connectors/engine ./internal/connectors/hooks/github -run '^(TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubWriteHook|TestGitHubDeclaredRateLimits|TestGitHubRateLimitAdmission)' -count=1` — pass.
- `go vet ./internal/connectors/engine ./internal/connectors/hooks/github` — pass.
- The later released bounded race and local acceptance matrix is recorded
  below; no no-mistakes run, push, PR, CI, or merge has started.

## Resumed Intake Provenance

The historic phase snapshot SHA-256 remains
`939f14f61defd993f8ad0335a5aeb617d97083c9f73a6a75259d0e312ae8f408`.
At broad-validation intake, the current canonical plan file hashes
`5c7aeeacdb5792ad259abb709f3d732f183f5d19b03db4c9863b3ef566044e06` and
the current live ledger explicitly releases local #4072 acceptance. The
reconciliation report's unchanged bounded matrix is the only authorized test
scope.

## Active correction

`make lint` found only two unused private forwarding wrappers in
`internal/connectors/engine/auth.go`. The causal lint RED is recorded in
`TDD-LEDGER.md`; correction 1/5 removed them and passed the focused auth matrix
plus `make lint`. The complete report-defined matrix still reruns before the
handoff is ready. No pipeline, push, PR, UDS, or parked-parent action is part
of this correction.

## Active generated-only closure

The `certification-matrix --check` failure is inherited from both recovered
base and #4072 head. It is not a new #4072 behavior defect and does not change
the existing **1/5** correction accounting. The only permitted production
artifact action is the canonical generator's six-line `discovery_source`
synchronization, preserving the known stripped semantic hash
`bc5d14758c26755d83a9dc4dcbb715da31d95f67de38e352bc652b752c0819bc` and
recording #4026/#4034 as a non-imported generator precedent.

The recorded RED exits 1 at both recovered base and child source; the canonical
generator then passes its `--check` at the child worktree and produces matrix
SHA-256 `e63b906cb640b8fb4fc8fd46c1076b77b7dbced7889919d60527f9b4335d520a`.
The six-line `discovery_source` update does not alter the stripped semantic
SHA-256 `bc5d14758c26755d83a9dc4dcbb715da31d95f67de38e352bc652b752c0819bc`.

## Released local acceptance

At `f52745f…`, formatting/diff checks, the 20× GitHub App auth selector,
serial four-package functional test, three-pass focused race test, scoped vet,
`internal/cli`, real-binary build/smoke, and every report-defined individual
repository gate passed. The fresh ledger remains **1/5**. The next state is
only the local no-mistakes vector; no parent ref, PR, external CI, UDS child,
parked #3754 run, provider call, or merge has been touched.
