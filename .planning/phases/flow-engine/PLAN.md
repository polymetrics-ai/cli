# PLAN — Flow Engine (Phase 0)

Each behavior task (B-N) is immediately paired with a test task (T-N). Test task must be
committed red (failing) before the implementation. Docs-only tasks are tagged [DOCS].

Design direction: not applicable (CLI/backend only).

---

## Wave 0 — Errors + manifest parse/validate

### T-01 [TEST] Manifest validation table-driven tests
- File: `internal/flow/flow_test.go`
- Cover: valid manifest round-trips; missing `name`; unknown kind; missing `connection` on sync;
  missing `sql` on query; duplicate step IDs; `in` references non-existent table.
- Must be RED before B-01 is written.

### B-01 [BEHAVIOR] Sentinel errors + manifest parse/validate
- Files: `internal/flow/errors.go`, `internal/flow/manifest.go`
- Implement `ParseManifest([]byte) (FlowManifest, error)` using `encoding/json` (manifests are
  normalized to JSON internally; accept YAML via `encoding/...` — stdlib only: use `gopkg.in/yaml`
  only if already in go.mod, else use a minimal hand-rolled YAML-to-map approach via `encoding/json`
  after converting with `os.ReadFile`).

  GATE: check `go.mod` before implementing YAML decode. If no YAML library is in go.mod, implement
  `ParseManifestYAML` using `gopkg.in/yaml.v3` ONLY IF already present; otherwise author manifests
  as JSON in tests and document YAML as a future add. Do not add a new dependency.

- Expose `ValidateManifest(FlowManifest) []error`.
- Gates T-01 green.

---

## Wave 1 — DAG build + topological sort + cycle detection

### T-02 [TEST] DAG tests
- File: `internal/flow/flow_test.go`
- Cover: linear chain A→B→C produces correct order; independent steps (any valid order);
  two-step cycle A↔B returns `ErrCyclicDependency`; three-node diamond (A→B, A→C, B→D, C→D)
  produces a valid ordering where A is first and D is last.
- Must be RED before B-02.

### B-02 [BEHAVIOR] DAG build + Kahn's topo sort
- File: `internal/flow/dag.go`
- `BuildDAG(manifest FlowManifest) ([]string, error)` — returns ordered step IDs or
  `ErrCyclicDependency`.
- Gates T-02 green.

---

## Wave 2 — Checkpoint store

### T-03 [TEST] Checkpoint store tests
- File: `internal/flow/flow_test.go`
- Cover: `Set` then `Get` returns same value; `Clear` removes all entries for flow; unknown
  step returns `""` not an error; concurrent sets do not corrupt (use a temp dir per test).
- Must be RED before B-03.

### B-03 [BEHAVIOR] Checkpoint store (file-backed)
- File: `internal/flow/checkpoint.go`
- Implement `FileCheckpointStore` satisfying `CheckpointStore` interface, backed by a JSON file
  at `<dir>/flow-checkpoints.json`. Reuse `internal/state.JSONStore` pattern (copy pattern,
  do not import internal/state to avoid coupling).
- Gates T-03 green.

---

## Wave 3 — Engine (lease + execute + ledger)

### T-04 [TEST] Engine lease contention test
- File: `internal/flow/flow_test.go`
- Cover: holding the lock file before calling `engine.Run` returns `ErrLeaseHeld`; after
  release, second run proceeds.
- Must be RED before B-04.

### B-04 [BEHAVIOR] Engine lease acquisition
- File: `internal/flow/engine.go` (partial — just lock/unlock)
- `Engine.acquireLease()` and `Engine.releaseLease()` using `internal/state.FileLock`.
- Gates T-04 green.

### T-05 [TEST] Engine dependency ordering test
- File: `internal/flow/flow_test.go`
- Cover: with a stub `AppAdapter`, assert steps execute in DAG order (record call order);
  assert a step whose dependency failed is not executed.
- Must be RED before B-05.

### B-05 [BEHAVIOR] Engine step dispatch + ordering
- File: `internal/flow/engine.go` (extend)
- Wire `runSync` and `runQuery` dispatchers; iterate in topo order; skip on checkpoint hit;
  stop on failure.
- Gates T-05 green.

### T-06 [TEST] Engine checkpoint/resume test
- File: `internal/flow/flow_test.go`
- Cover: pre-seed step-A as `success` in checkpoint store; run a two-step flow; assert step-A
  dispatcher is never called but step-B is; assert `StepResult.Status == "skipped"` for step-A.
- Must be RED before B-06.

### B-06 [BEHAVIOR] Engine checkpoint skip logic
- File: `internal/flow/engine.go` (extend)
- Add checkpoint read before dispatch; `--force` clears checkpoint store.
- Gates T-06 green.

### T-07 [TEST] Engine ledger write test
- File: `internal/flow/flow_test.go`
- Cover: after a successful run, ledger contains two entries per step (status=running,
  status=success) and one entry for the overall flow run.
- Must be RED before B-07.

### B-07 [BEHAVIOR] Engine ledger write
- File: `internal/flow/engine.go` (extend)
- Write `ledger.RunRecord` at step start and finish; write flow-level record on completion.
  Use `LedgerAdapter` interface to keep tests dependency-free (stub in tests).
- Gates T-07 green.

---

## Wave 4 — CLI dispatcher

### T-08 [TEST] CLI flow subcommand tests
- File: `internal/cli/flow_cli_test.go` (new)
- Cover: `pm flow list` with an empty flows dir returns `[]`; `pm flow plan` with a valid
  two-step manifest file returns `{"status":"ok",...}`; `pm flow plan` with cyclic manifest
  returns non-zero exit and error JSON; `pm flow status <name>` returns last run.
- Use existing CLI test pattern (e.g. `internal/cli/cli_test.go` style with `bytes.Buffer`).
- Must be RED before B-08.

### B-08 [BEHAVIOR] CLI flow dispatcher
- Files: `internal/cli/flow_cli.go` (new), `internal/cli/cli.go` (add `case "flow":`)
- Implement `runFlow(ctx, a, rest, stdout, jsonOut)` routing `plan|preview|run|status|list`.
- Wire `Engine` with real `app.App` adapter and `ledger.JSONLedger`.
- Gates T-08 green.

---

## Wave 5 — Verification + docs

### T-09 [TEST — integration smoke] Full sync→query chain test
- File: `internal/flow/flow_test.go` (or `internal/cli/flow_cli_test.go`)
- Use stub `AppAdapter` that records calls. Run a two-step manifest (sync→query). Assert:
  - both steps execute in order
  - ledger has correct entries
  - `RunResult.Status == "ok"`
  - checkpoint persists after run
- This is the gate test referenced in the phase goal.

### D-01 [DOCS] Update `internal/cli/docs.go`
- Register `flow` command in the manual index with subcommand help text for
  `plan|preview|run|status|list`.

### D-02 [DOCS] Planner trace
- Write `.planning/phases/flow-engine/traces/planner-trace.md`.

---

## Human gates in this phase

None triggered in Phase 0. The following are flagged for future phases:
- Any new go.mod dependency (YAML lib check in B-01 above is a conditional gate).
- First network-write enablement (Phase 1).

## Execution order (wave summary)

| Wave | Tasks | Depends on |
|------|-------|-----------|
| 0 | T-01, B-01 | — |
| 1 | T-02, B-02 | Wave 0 |
| 2 | T-03, B-03 | Wave 0 |
| 3 | T-04→T-07, B-04→B-07 | Waves 1 + 2 |
| 4 | T-08, B-08 | Wave 3 |
| 5 | T-09, D-01, D-02 | Wave 4 |
