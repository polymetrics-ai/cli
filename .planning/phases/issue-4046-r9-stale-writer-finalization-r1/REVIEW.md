---
status: clean
phase: issue-4046-r9-stale-writer-finalization-r1
depth: standard
files_reviewed: 2
findings:
  critical: 0
  warning: 0
  info: 0
  total: 0
reviewed_at: 2026-08-11T10:05:48Z
---

# Code review — #4046 R9 stale-writer run finalization

## Method

Manual standard-depth inline review is the documented fallback because the non-numbered phase cannot resolve through `init.phase-op` and the repository contract forbids spawning a reviewer role. The explicit review scope was `internal/app/app.go` and `internal/app/transport_dispatch_test.go`; the review also traced `RunETL → runTransportETL → updateState → failRun`, the JSON-store update semantics, and the R7/R8 regression boundary.

## Findings and disposition

No unresolved Critical, Warning, or Info findings remain.

- **Resolved during review fix:** a definite pre-rename JSON-store failure returns the callback's speculative state, so typed stale-conflict finalization now returns that run only if the callback observed an already-terminal target or the store reports a committed/indeterminate transitioned outcome. The focused real-CAS regression proves a non-committed transition returns `Run{}` and leaves the durable loser `running` without changing winner or unrelated state; companion tests retain the terminal-target and committed/indeterminate cases.
- **Resolved during review:** the first focused concurrent-writer test had the winner and unrelated mutation in one writer update. It was strengthened in `8841ddbd0` to observe the real typed CAS conflict from the stale app, then persist unrelated state in a second writer update before `failRun`; the repeated focused and race checks passed afterward.
- The typed exception is anchored to the private transport conflict sentinel through `errors.Is`; unrelated failure paths retain the exact revision comparison and cannot become a generic last-writer-wins refresh.
- The typed path receives the current state only while the JSON-store lock is held, changes only the matching generated run ID, and requires its status to remain `running`. It does not touch `StreamStates`, project checkpoints, winner data, or unrelated runs.
- Error identity is retained on successful finalization and on a persistence/update error through `errors.Join`; returned runs follow the commit-outcome truth above and never expose an uncommitted callback state.
- The tests exercise the original durable symptom, reopen behavior, typed error identity, winner/unrelated retention, a writer after conflict observation, ordinary-error guarding, terminal-target refusal, cancellation after acknowledgement, all seven modes, the race detector, and the unchanged R7/R8 suite.
- No provider implementation, connector protocol, state-store API, credential surface, warehouse implementation, or external integration changed.

## Review checks

- `go test -race -count=10 -timeout 20m ./internal/app -run '^(TestRunETLTransportStaleWriterFinalizesLosingRun|TestFailRunTransportConflictPreservesLatestConcurrentState|TestRunETLTransportStaleWriterFinalizesAfterCancellation)$'` passed with no race report.
- The exact R7/R8 nine-test command, affected `internal/app` package, vet/build, and each required repository gate passed; details are in `VERIFICATION.md`.
- The macOS linker warning observed under `-race` was a platform toolchain warning only; it did not produce a test or race failure.
