---
phase: issue-4072-github-app-auth-admission-r1
plan: "02"
type: standard
wave: 2
gap_closure: true
depends_on:
  - "01"
autonomous: true
requirements:
  - ISSUE-4072
files_modified:
  - .planning/phases/issue-4072-github-app-auth-admission-r1/VERIFICATION.md
  - .planning/phases/issue-4072-github-app-auth-admission-r1/TDD-LEDGER.md
  - .planning/phases/issue-4072-github-app-auth-admission-r1/RUN-STATE.md
  - .planning/phases/issue-4072-github-app-auth-admission-r1/SUMMARY.md
  - .planning/phases/issue-4072-github-app-auth-admission-r1/PR-BODY.md
  - .planning/phases/issue-4072-github-app-auth-admission-r1/NO-MISTAKES-HANDOFF.md
---

# Gap plan — broad local acceptance and guarded delivery handoff

## Scope

Close only the deferred local-verification and handoff gaps for Issue #4072.
The production GREEN remains `3f83bf3afc6efa0ebc323e385e4345f588a41db1`;
the prior focused-evidence checkpoint remains
`72c573bca90f3803ccfe09e914a6bb411c903430`. This plan does not change Go
source, tests, provider policy, credential semantics, CLI/docs surface, parent
route, or the separate UDS finding.

The historical phase snapshot remains SHA-256
`939f14f61defd993f8ad0335a5aeb617d97083c9f73a6a75259d0e312ae8f408`.
At this resumed intake, the current canonical plan file hashes
`5c7aeeacdb5792ad259abb709f3d732f183f5d19b03db4c9863b3ef566044e06`.
The live execution ledger explicitly resumes #4072 **local** acceptance and
keeps the report-defined requirements unchanged; this record preserves both
provenance facts rather than rewriting historical evidence.

## Inline GSD fallback

The named issue phase is outside the numeric roadmap and the canonical
single-worker contract forbids planner/executor/verifier/reviewer spawning.
`scripts/gsd prompt plan-phase ... --gaps` and
`scripts/gsd prompt execute-phase ... --gaps-only` were resolved and are
executed inline. This is a documentation-and-validation gap plan, not a new
behavioral slice, so it adds no RED/GREEN cycle and does not consume the fresh
0/5 correction budget.

## Tasks

### 1. Run only the reconciliation report’s local acceptance matrix

Run sequentially, using the report’s exact package boundaries and timeouts:

1. `gofmt -l` check of the five changed Go files.
2. Twenty-repeat GitHub App auth selector.
3. Bounded four-package functional matrix with `GOMAXPROCS=2 -p 1`.
4. Bounded engine/GitHub race matrix with `GOMAXPROCS=2 -p 1 -race -count=3`.
5. Four-package `go vet`, `internal/cli`, and `go build ./cmd/pm`.
6. The report’s repository gates one at a time, never `make verify` or
   `go test ./...` as a single machine-contending command.

Use only local synthetic auth, secret-blind fakes, and normal static/generator
checks. Do not inspect credentials, call GitHub, mutate a provider, or broaden
the test scope beyond the report.

### 2. Record exact results and prepare, but do not start, delivery

After the matrix passes, update the existing verification, TDD ledger, run
state, and summary with exact commands, head, date, and outcomes. Prepare the
no-mistakes intent/handoff and draft PR body locally, but do not run
no-mistakes while #3856 is in a heavy validation stage. Do not push, create a
PR, publish the #3754 parent, create the UDS child, merge, or interact with
the parked #3754 run.

## Acceptance criteria

- Every command under “Remaining minimum local acceptance” in the #4072
  reconciliation report passes at the unchanged implementation head.
- `scripts/verify-gsd-workflow da8a8ff07aaf00e5c7965cd4d1d3c7252017d785`,
  `make docs-check`, `git diff --check`, and the focused auth selector pass
  after final evidence edits.
- The handoff contains the exact safe `no-mistakes axi run --skip=push,pr,ci`
  vector but records it as held until #3856 heavy validation releases.
- The parent-route block remains explicit: the stale remote #3754 ref and
  absent correct parent draft prevent child push/PR until the captain chooses
  the bounded parent-publication route.

## Verification order

Run no-mistakes, any push, draft PR, CI, or review-service action only after a
later live release verifies that #3856 is no longer in its heavy validation
stage and the exact #3754 parent publication decision has been resolved. Those
conditions are intentionally not inferred from this local plan.
