# TEST-PLAN — Flow Engine (Phase 0)

## Approach

Strict TDD: every behavior task in PLAN.md must have a committed failing test (red evidence in
TDD-LEDGER.md) before implementation begins. Tests use the stdlib `testing` package and the
existing `testify` assertion library already in `go.mod` (do not introduce a different assertion
library).

## Test files

| File | Package | Covers |
|------|---------|--------|
| `internal/flow/flow_test.go` | `flow` | manifest, DAG, checkpoint, engine |
| `internal/cli/flow_cli_test.go` | `cli` | CLI subcommand dispatch |

## Unit test cases

### Manifest validation (T-01)

| # | Input | Expected |
|---|-------|----------|
| 1 | Valid two-step sync→query manifest | No errors |
| 2 | `name: ""` | ErrManifestInvalid |
| 3 | `kind: unknown` | ErrManifestInvalid |
| 4 | sync step, no `connection` | ErrManifestInvalid |
| 5 | sync step, empty `streams` | ErrManifestInvalid |
| 6 | query step, no `sql` | ErrManifestInvalid |
| 7 | Duplicate step IDs | ErrManifestInvalid |
| 8 | `in` references table not in any `out` | ErrManifestInvalid |
| 9 | `version: 2` | ErrManifestInvalid |

### DAG / topological sort (T-02)

| # | Graph | Expected order constraints |
|---|-------|---------------------------|
| 1 | A→B→C (linear) | A before B before C |
| 2 | A, B independent | any order, both present |
| 3 | A→B, B→A (2-cycle) | ErrCyclicDependency |
| 4 | A→B→C→A (3-cycle) | ErrCyclicDependency |
| 5 | Diamond: A→B, A→C, B→D, C→D | A first, D last |

### Checkpoint store (T-03)

| # | Scenario | Expected |
|---|----------|----------|
| 1 | Set then Get same step | value matches |
| 2 | Get unknown step | "" (no error) |
| 3 | Clear flow removes all | Get returns "" |
| 4 | Two flows, clear one | other flow untouched |

### Lease contention (T-04)

| # | Scenario | Expected |
|---|----------|----------|
| 1 | Lock held, call Run | ErrLeaseHeld |
| 2 | Lock not held, Run succeeds, lock released | no error |
| 3 | Run panics (deferred release) | lock file removed |

### Dependency ordering (T-05)

| # | Scenario | Expected |
|---|----------|----------|
| 1 | Two steps A→B, both succeed | A called before B |
| 2 | Step A fails | step B not called |
| 3 | Three steps, middle fails | third step not called |

### Checkpoint/resume (T-06)

| # | Scenario | Expected |
|---|----------|----------|
| 1 | Step A pre-seeded success | A dispatcher not called, StepResult.Status="skipped" |
| 2 | `--force` clears checkpoints | A dispatcher called |
| 3 | After successful run | all steps checkpointed as success |

### Ledger writes (T-07)

| # | Scenario | Expected |
|---|----------|----------|
| 1 | Successful two-step run | 4 ledger entries (2 per step × 2 states) + 1 flow entry |
| 2 | Step fails | failed step has status=failed entry; flow has status=failed entry |

### CLI subcommand (T-08)

| # | Command | Expected |
|---|---------|----------|
| 1 | `pm flow list` (empty dir) | exit 0, JSON `{"flows":[]}` |
| 2 | `pm flow plan --file valid.json` | exit 0, JSON with status=ok |
| 3 | `pm flow plan --file cyclic.json` | exit non-zero, JSON error |
| 4 | `pm flow status missing` | exit non-zero, error |
| 5 | `pm flow preview --file valid.json` | exit 0, dry_run status |

### Integration smoke (T-09)

Full sync→query two-step manifest through the engine with stub AppAdapter:
- correct execution order
- ledger entries correct count
- RunResult.Status == "ok"
- checkpoints persisted

## Red evidence

Before any implementation commit, each T-N test file must be committed with the test defined but
the implementation absent, producing a compile error or failing assertion. Record the failing output
in TDD-LEDGER.md.

## Running tests

```bash
export GOTOOLCHAIN=auto
go test ./internal/flow/... -v -run TestManifest
go test ./internal/flow/... -v -run TestDAG
go test ./internal/flow/... -v -run TestCheckpoint
go test ./internal/flow/... -v -run TestEngine
go test ./internal/cli/... -v -run TestFlow
go test ./...
make verify
```

## Test doubles / stubs

- `stubAppAdapter` — records ETLRun and QuerySQL calls; returns configurable results.
- `stubLedger` — in-memory slice of RunRecord; no file I/O.
- Temp dirs (`t.TempDir()`) for lock files and checkpoint files.
