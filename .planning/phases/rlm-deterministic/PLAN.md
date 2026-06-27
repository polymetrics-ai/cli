# PLAN — Phase 2: RLM Deterministic Backend

All tasks follow red-first TDD: the test task must be committed (failing) before the behavior task.
Tags: [BEHAVIOR] = paired with a test; [DOCS-ONLY] = no behavior change; [HUMAN-GATE] = stop for approval.

---

## Wave 0 — Scaffolding (no behavior)

### T0.1 [DOCS-ONLY] Create package skeleton
- Create `internal/rlm/` directory.
- Write `internal/rlm/rlm.go` with package declaration, `ErrNotImplemented` sentinel, `Analyzer` interface, `RunRequest`, `RunResult` types — all exported, no logic.
- Write `internal/rlm/spec.go` with `Spec`, `Feature` types — struct definitions only, no parsing logic yet.
- Write `internal/rlm/model.go` with `ModelAnalyzer` stub (returns `ErrNotImplemented`).
- Verify: `go build ./internal/rlm/...` compiles.

---

## Wave 1 — Spec parsing (red-first)

### T1.1 [BEHAVIOR] Red test: spec parsing
- File: `internal/rlm/spec_test.go`
- Table-driven tests for `ParseSpec([]byte)`:
  - valid YAML with 2 features and weights → returns `*Spec`, nil error
  - missing `name` field → returns descriptive error
  - feature with negative weight → returns validation error
  - feature with both `score_if_set` and `score_if_gt` set → allowed (validate threshold presence)
  - empty features list → returns validation error
- Run `go test ./internal/rlm/...` — must FAIL (ParseSpec not implemented). Record in TDD-LEDGER.md.

### T1.2 [BEHAVIOR] Implement `ParseSpec`
- File: `internal/rlm/spec.go`
- Use `encoding/json` with YAML-as-JSON via `encoding/json` — wait, spec files are YAML. Use stdlib only.
- DECISION: Spec files use JSON (not YAML) to stay stdlib-only. Update CLI flag doc accordingly (--spec accepts JSON). Alternatively, accept both: detect `{` prefix → JSON, else treat as... this requires a YAML parser.
- RESOLUTION: Spec files are JSON. The `--spec` flag description says "JSON spec file". This removes any YAML dependency. Update SPEC.md note. The `likely_customers.json` testdata file is JSON.
- Implement `ParseSpec(data []byte) (*Spec, error)` using `encoding/json`. Add `Validate() error` on `*Spec`.
- Tests from T1.1 pass.

---

## Wave 2 — Scoring engine (red-first)

### T2.1 [BEHAVIOR] Red test: deterministic scoring
- File: `internal/rlm/deterministic_test.go`
- Tests:
  - `TestDeterminismSameInputSameOutput`: run scorer twice on identical input slice → identical `_rlm_score` values and identical ordering
  - `TestScoringWeightedSum`: 2 features, known weights, known field values → expected normalized score (hand-calculated)
  - `TestScoringScoreIfSet`: field present/absent → correct branch
  - `TestScoringScoreIfGT`: numeric field above/below threshold → correct branch
  - `TestScoringAllZeroWeights`: edge case — all weights zero → score = 0.0, no panic
  - `TestScoringEmptyRecords`: zero rows → RunResult.RecordsRead=0, no error
- These tests call a `scoreRecords(spec *Spec, records []connectors.Record) ([]connectors.Record, error)` internal function.
- Run `go test` — must FAIL. Record in TDD-LEDGER.md.

### T2.2 [BEHAVIOR] Implement scoring engine
- File: `internal/rlm/deterministic.go`
- Implement `scoreRecords`, `DeterministicAnalyzer.Run`.
- Sorting: by `_rlm_score` desc, then `_polymetrics_raw_id` asc.
- All tests from T2.1 pass.

---

## Wave 3 — Materialization (red-first)

### T3.1 [BEHAVIOR] Red test: materialization schema
- File: `internal/rlm/deterministic_test.go` (add cases)
- `TestMaterializationWritesNDJSON`: create a temp warehouse dir, write 3 fixture rows to InTable NDJSON, call `DeterministicAnalyzer.Run`, read OutTable NDJSON, assert:
  - row count matches
  - each row has `_rlm_score`, `_rlm_mode`, `_rlm_spec`, `_rlm_scored_at` fields
  - original source fields are preserved
- `TestMaterializationAtomic`: partial failure mid-write must not leave a corrupt OutTable (simulate by checking temp file cleanup).
- `TestDryRunDoesNotWrite`: `DryRun=true` → OutTable file is NOT created.
- Run `go test` — must FAIL. Record in TDD-LEDGER.md.

### T3.2 [BEHAVIOR] Implement materialization in DeterministicAnalyzer.Run
- Read InTable NDJSON, parse `localRawRecord.Record` fields.
- Score, sort, write OutTable atomically (temp + rename).
- Append `ledger.RunRecord` if a ledger is provided (optional field on `DeterministicAnalyzer`).
- All tests from T3.1 pass.

---

## Wave 4 — Fixture backend (red-first)

### T4.1 [BEHAVIOR] Red test: fixture backend
- File: `internal/rlm/fixture_test.go`
- `TestFixtureRunReturnsRows`: `FixtureAnalyzer.Run` with any spec → RunResult.RecordsScored > 0, OutTable written.
- `TestFixtureScoresMatchDeterministic`: apply same spec to `DefaultFixtureRows` via both fixture and deterministic backends → identical `_rlm_score` values.
- `TestFixtureIgnoresInTable`: pass non-existent InTable → no error (fixture ignores it).
- Run `go test` — must FAIL. Record in TDD-LEDGER.md.

### T4.2 [BEHAVIOR] Implement FixtureAnalyzer
- File: `internal/rlm/fixture.go`
- `DefaultFixtureRows` — at least 5 hardcoded `connectors.Record` values representing contacts.
- `FixtureAnalyzer.Run` — scores `DefaultFixtureRows` using shared `scoreRecords`, writes OutTable.
- All tests from T4.1 pass.

---

## Wave 5 — CLI wiring (red-first)

### T5.1 [BEHAVIOR] Red test: CLI rlm verb
- File: `internal/cli/rlm_cli_test.go` (new)
- Use existing CLI test pattern (`cli.Run(args, stdout, stderr)` from `internal/cli/cli_test.go`).
- `TestRLMRunDeterministic`: set up temp warehouse dir with InTable, call `pm rlm run --spec ... --in contacts --out lead_scores --mode deterministic`, assert exit 0, JSON envelope has `records_scored > 0`.
- `TestRLMRunFixture`: `--mode fixture` → exit 0, OutTable created.
- `TestRLMRunModelStub`: `--mode model` → exit non-zero, error contains "not implemented".
- `TestRLMRunMissingSpec`: `--spec /nonexistent.json` → exit 1, error message.
- `TestRLMRunMissingRequired`: missing `--out` flag → exit 1.
- `TestRLMRunDryRun`: `--dry-run` → exit 0, OutTable NOT written.
- Run `go test ./internal/cli/...` — must FAIL. Record in TDD-LEDGER.md.

### T5.2 [BEHAVIOR] Wire CLI dispatch
- Add `case "rlm":` to `switch cmd` in `internal/cli/cli.go`.
- Create `internal/cli/rlm_cli.go`: `runRLM(ctx, a, rest, stdout, jsonOut)` function.
- Parse flags: `--spec`, `--in`, `--out`, `--mode`, `--dry-run`.
- Build `rlm.RunRequest`, select backend by mode, call `analyzer.Run`.
- Write JSON envelope to stdout if `--json`; human-readable summary otherwise.
- All T5.1 tests pass.

---

## Wave 6 — Flow step kind (coordinate, no new tests in this phase)

### T6.1 [DOCS-ONLY] Document rlm flow step kind
- Add `rlm` to the recognized step kinds in `internal/flow` (if Phase 0 flow-engine is present).
- If `internal/flow` does not yet exist (Phase 0 not merged), document the expected YAML schema in `testdata/likely_customers_flow.yaml` only.
- No new behavior in `internal/rlm`; `internal/rlm` must NOT import `internal/flow`.

---

## Wave 7 — End-to-end fixture flow

### T7.1 [BEHAVIOR] Red test: end-to-end offline flow
- File: `internal/rlm/e2e_test.go`
- `TestLikelyCustomersFlowOffline`: using only fixture backend:
  1. Call `FixtureAnalyzer.Run` with `testdata/likely_customers.json` spec.
  2. Assert `RunResult.RecordsScored >= 5`.
  3. Read OutTable, assert top-scored record has `_rlm_score >= 0.5`.
  4. Re-run with identical inputs → identical scores (determinism assertion on fixture path).
- Run `go test` — must FAIL until T4.2 is done (but add now to track state).

### T7.2 [BEHAVIOR] Testdata: create `likely_customers.json` spec
- File: `internal/rlm/testdata/likely_customers.json`
- Features: `email` (weight 0.3), `company` (weight 0.4), `title` (weight 0.3) — all `score_if_set`.
- Name: `"likely-customers"`.
- This is a [BEHAVIOR] task only in that it unblocks T7.1; no Go logic changes.

---

## HUMAN GATE — Model backend (Phase 4)

**STOP HERE** before implementing `ModelAnalyzer.Run`. The model backend:
- Makes outbound network calls to the Claude API.
- Requires credential configuration (API key in vault).
- Changes the network/credential surface of `pm`.

A human must explicitly approve Phase 4 before any of the following are written:
- HTTP client code in `internal/rlm/model.go`
- Credential lookup for model API key
- Any caching of model responses

Document this gate in `TDD-LEDGER.md` and in the phase summary. Do NOT proceed to Phase 4 automatically.

---

## Verification gate (run after all waves)

```bash
export GOTOOLCHAIN=auto
gofmt -w internal/rlm internal/cli
go vet ./internal/rlm/... ./internal/cli/...
go test ./internal/rlm/... ./internal/cli/...
go build ./cmd/pm
make verify
```

All must exit 0 before this phase is considered complete.
