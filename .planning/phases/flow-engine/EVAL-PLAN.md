# EVAL-PLAN — Flow Engine (Phase 0)

## Phase gate criteria

All of the following must be true before declaring Phase 0 done and requesting human review
of the Phase 1 gate.

### Functional correctness

| Check | Verification method |
|-------|-------------------|
| Two-step sync→query manifest executes both steps in order | T-09 passes |
| Cycle detection returns ErrCyclicDependency | T-02 passes |
| Lease contention returns ErrLeaseHeld | T-04 passes |
| Checkpoint skip works on re-run | T-06 passes |
| Ledger contains correct entries after run | T-07 passes |
| CLI returns structured JSON with correct fields | T-08 passes |
| All validation error cases caught | T-01 passes |

### Quality gate

| Check | Command |
|-------|---------|
| No formatting issues | `gofmt -w cmd internal; git diff --exit-code` |
| No vet issues | `go vet ./...` |
| All tests pass | `go test ./...` |
| Binary builds | `go build ./cmd/pm` |
| Docs validate | `make docs-check` |
| Full verify | `make verify` |

### Coverage targets

- `internal/flow` package: all exported functions exercised by tests.
- No test skips or `t.Skip()` calls without documented justification.

### Prompt-eval notes

For agent use of `pm flow plan --json`:
- Response must be valid JSON parseable with `json.Unmarshal`.
- `status` field must be `"ok"`, `"failed"`, or `"dry_run"` (no other values).
- Each `steps[]` entry must have `id`, `kind`, `status`, `records_read`, `records_written`,
  `duration_ns`.
- Error responses must have `error` field.
- Exit code must be 0 on success, non-zero on any error.

These are verified by T-08 and T-09. If an agent consuming the output reports a parse error,
the envelope contract in API-CONTRACT.md is the source of truth.

## Definition of done

- `make verify` exits 0.
- TDD-LEDGER.md contains red evidence for every T-N task before its paired B-N commit.
- VERIFICATION.md updated with final `make verify` output.
- SUMMARY.md written with phase outcomes.
- All artifacts in `.planning/phases/flow-engine/` committed.
