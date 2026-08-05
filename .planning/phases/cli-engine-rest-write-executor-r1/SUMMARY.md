# REST write no-redaction correction

## Outcome

Captain decision `remove-write-redaction` is implemented for `rest_write` only.
The executor now returns decoded response content intact for the three legacy
redaction-named policies, ignores direct-write request/sensitive-policy field
redaction, preserves direct-write command errors and plan samples, and stores
the complete direct-write failure text in its reverse-run report.

## Boundaries preserved

- `none` still returns no response body.
- Unknown output policies still fail closed.
- Read-path code is unchanged.
- No connector declaration changed.
- Preview digest binding, approval, destructive confirmation, one-shot
  execution, no retries, redirect refusal, endpoint binding, and size caps are
  unchanged.

## Evidence

`TDD-LEDGER.md` captures the initial failures and green retest. Local scoped
package tests, `internal/cli`, vet, build, and all individual verification
gates passed; aggregate `go test ./...` and aggregate `make verify` were not
performed and remain CI coverage.

## Not performed

- No CI result has been observed for this correction yet; it is expected after
  the corrective commit is pushed to the existing PR branch.
- Runtime-backed integrations, aggregate `go test ./...`, aggregate `make
  verify`, and a fresh no-mistakes pipeline run were not performed.
- The project-instructions maintenance helper made no change because it found
  both a real `AGENTS.md` and `CLAUDE.md` in this worktree.

## Workflow note

The required GSD `programming-loop` adapter command was absent even though
`scripts/gsd doctor` passed, so the repository-permitted manual GSD/TDD
fallback was used and recorded in `RUN-STATE.json`.
